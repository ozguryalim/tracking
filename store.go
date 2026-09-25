package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	errNotFound   = errors.New("record not found")
	errValidation = errors.New("invalid input")
	errAmbiguous  = errors.New("ambiguous ID prefix")
	errConflict   = errors.New("record changed; reload and retry")
)

type Store struct {
	db  *sql.DB
	key string
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	TaskCount   int    `json:"task_count"`
	DoneCount   int    `json:"done_count"`
	ActiveCount int    `json:"active_count"`
}

type Plan struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Goal      string `json:"goal"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Task struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	PlanID      string `json:"plan_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

type Event struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	PlanID     string `json:"plan_id,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
	Kind       string `json:"kind"`
	Note       string `json:"note"`
	Actor      string `json:"actor"`
	OccurredAt string `json:"occurred_at"`
}

type ProjectDetail struct {
	Project Project `json:"project"`
	Plans   []Plan  `json:"plans"`
	Tasks   []Task  `json:"tasks"`
	Events  []Event `json:"events"`
}

type TaskPatch struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	PlanID      *string `json:"plan_id"`
	Note        string  `json:"note"`
}

type ImportedTask struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PlanImport struct {
	Title string         `json:"title"`
	Goal  string         `json:"goal"`
	Tasks []ImportedTask `json:"tasks"`
}

func openStore(path string) (*Store, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// One connection keeps connection-scoped pragmas consistent. Separate CLI and
	// dashboard processes still coordinate through SQLite's WAL locking.
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("configure database: %w", err)
		}
	}
	for _, statement := range []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			path TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS plans (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			goal TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			plan_id TEXT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL CHECK(status IN ('todo', 'doing', 'blocked', 'done')),
			created_at TEXT NOT NULL,
			started_at TEXT,
			completed_at TEXT,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			plan_id TEXT REFERENCES plans(id) ON DELETE SET NULL,
			task_id TEXT REFERENCES tasks(id) ON DELETE SET NULL,
			kind TEXT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			actor TEXT NOT NULL,
			occurred_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS tasks_project_status ON tasks(project_id, status, created_at)`,
		`CREATE INDEX IF NOT EXISTS tasks_plan ON tasks(plan_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS plans_project ON plans(project_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS events_project_time ON events(project_id, occurred_at DESC)`,
		`CREATE INDEX IF NOT EXISTS events_task_time ON events(task_id, occurred_at DESC)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			db.Close()
			return nil, fmt.Errorf("initialize database: %w", err)
		}
	}
	return &Store{db: db, key: databaseKey(path)}, nil
}

func databaseKey(path string) string {
	sum := sha256.Sum256([]byte(filepath.Clean(path)))
	return hex.EncodeToString(sum[:])
}

func (s *Store) Close() error { return s.db.Close() }

func nowUTC() string { return time.Now().UTC().Format("2006-01-02T15:04:05.000000000Z") }

func newID(prefix string) (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(b), nil
}

func actorName() string {
	if actor := strings.TrimSpace(os.Getenv("TRACKING_ACTOR")); actor != "" {
		return actor
	}
	if actor := strings.TrimSpace(os.Getenv("USER")); actor != "" {
		return actor
	}
	if actor := strings.TrimSpace(os.Getenv("USERNAME")); actor != "" {
		return actor
	}
	return "local"
}

func insertEvent(ctx context.Context, tx *sql.Tx, projectID, planID, taskID, kind, note, actor, at string) error {
	id, err := newID("ev")
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO events(id, project_id, plan_id, task_id, kind, note, actor, occurred_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`,
		id, projectID, nullString(planID), nullString(taskID), kind, note, actor, at,
	)
	return err
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func (s *Store) CreateProject(ctx context.Context, name, path, actor string) (Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Project{}, fmt.Errorf("%w: project name is required", errValidation)
	}
	id, err := newID("pr")
	if err != nil {
		return Project{}, err
	}
	project := Project{ID: id, Name: name, Path: strings.TrimSpace(path), CreatedAt: nowUTC()}
	project.UpdatedAt = project.CreatedAt
	if err := s.insertProject(ctx, project, actor); err != nil {
		return Project{}, err
	}
	return project, nil
}

func (s *Store) insertProject(ctx context.Context, project Project, actor string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO projects(id, name, path, created_at, updated_at) VALUES(?, ?, ?, ?, ?)`,
		project.ID, project.Name, project.Path, project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if err := insertEvent(ctx, tx, project.ID, "", "", "project_created", project.Name, actor, project.CreatedAt); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) RegisterProject(ctx context.Context, project Project, actor string) error {
	if project.ID == "" || strings.TrimSpace(project.Name) == "" {
		return fmt.Errorf("%w: project ID and name are required", errValidation)
	}
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ?`, project.ID).Scan(&exists)
	if err == nil {
		_, err = s.db.ExecContext(ctx, `UPDATE projects SET name = ?, path = ? WHERE id = ?`,
			project.Name, project.Path, project.ID)
		return err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	project.CreatedAt = nowUTC()
	project.UpdatedAt = project.CreatedAt
	return s.insertProject(ctx, project, actor)
}

func (s *Store) ListProjects(ctx context.Context) ([]Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.id, p.name, p.path, p.created_at, p.updated_at,
		COUNT(t.id), COALESCE(SUM(CASE WHEN t.status = 'done' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN t.status = 'doing' THEN 1 ELSE 0 END), 0)
		FROM projects p LEFT JOIN tasks t ON t.project_id = p.id
		GROUP BY p.id ORDER BY p.updated_at DESC, p.name COLLATE NOCASE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := []Project{}
	for rows.Next() {
		var project Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Path, &project.CreatedAt, &project.UpdatedAt,
			&project.TaskCount, &project.DoneCount, &project.ActiveCount); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (s *Store) getProject(ctx context.Context, id string) (Project, error) {
	var project Project
	err := s.db.QueryRowContext(ctx, `SELECT id, name, path, created_at, updated_at FROM projects WHERE id = ?`, id).
		Scan(&project.ID, &project.Name, &project.Path, &project.CreatedAt, &project.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, errNotFound
	}
	return project, err
}

func (s *Store) ProjectDetail(ctx context.Context, id string) (ProjectDetail, error) {
	project, err := s.getProject(ctx, id)
	if err != nil {
		return ProjectDetail{}, err
	}
	detail := ProjectDetail{Project: project, Plans: []Plan{}, Tasks: []Task{}, Events: []Event{}}
	plans, err := s.db.QueryContext(ctx, `SELECT id, project_id, title, goal, created_at, updated_at
		FROM plans WHERE project_id = ? ORDER BY created_at DESC`, id)
	if err != nil {
		return ProjectDetail{}, err
	}
	for plans.Next() {
		var plan Plan
		if err := plans.Scan(&plan.ID, &plan.ProjectID, &plan.Title, &plan.Goal, &plan.CreatedAt, &plan.UpdatedAt); err != nil {
			plans.Close()
			return ProjectDetail{}, err
		}
		detail.Plans = append(detail.Plans, plan)
	}
	if err := plans.Err(); err != nil {
		plans.Close()
		return ProjectDetail{}, err
	}
	plans.Close()
	tasks, err := s.db.QueryContext(ctx, `SELECT id, project_id, plan_id, title, description, status,
		created_at, started_at, completed_at, updated_at FROM tasks WHERE project_id = ? ORDER BY created_at`, id)
	if err != nil {
		return ProjectDetail{}, err
	}
	for tasks.Next() {
		task, err := scanTask(tasks)
		if err != nil {
			tasks.Close()
			return ProjectDetail{}, err
		}
		detail.Tasks = append(detail.Tasks, task)
	}
	if err := tasks.Err(); err != nil {
		tasks.Close()
		return ProjectDetail{}, err
	}
	tasks.Close()
	events, err := s.db.QueryContext(ctx, `SELECT id, project_id, plan_id, task_id, kind, note, actor, occurred_at
		FROM events WHERE project_id = ? ORDER BY occurred_at DESC LIMIT 500`, id)
	if err != nil {
		return ProjectDetail{}, err
	}
	for events.Next() {
		event, err := scanEvent(events)
		if err != nil {
			events.Close()
			return ProjectDetail{}, err
		}
		detail.Events = append(detail.Events, event)
	}
	if err := events.Err(); err != nil {
		events.Close()
		return ProjectDetail{}, err
	}
	events.Close()
	detail.Project.TaskCount = len(detail.Tasks)
	for _, task := range detail.Tasks {
		if task.Status == "done" {
			detail.Project.DoneCount++
		}
		if task.Status == "doing" {
			detail.Project.ActiveCount++
		}
	}
	return detail, nil
}

func scanEvent(row taskScanner) (Event, error) {
	var event Event
	var planID, taskID sql.NullString
	err := row.Scan(&event.ID, &event.ProjectID, &planID, &taskID,
		&event.Kind, &event.Note, &event.Actor, &event.OccurredAt)
	event.PlanID, event.TaskID = planID.String, taskID.String
	return event, err
}

func (s *Store) TaskEvents(ctx context.Context, id string) ([]Event, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, plan_id, task_id, kind, note, actor, occurred_at
		FROM events WHERE project_id = ? AND task_id = ? ORDER BY occurred_at DESC, id DESC`, task.ProjectID, task.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) CreatePlan(ctx context.Context, projectID, title, goal, actor string) (Plan, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Plan{}, fmt.Errorf("%w: plan title is required", errValidation)
	}
	if _, err := s.getProject(ctx, projectID); err != nil {
		return Plan{}, err
	}
	id, err := newID("pl")
	if err != nil {
		return Plan{}, err
	}
	plan := Plan{ID: id, ProjectID: projectID, Title: title, Goal: strings.TrimSpace(goal), CreatedAt: nowUTC()}
	plan.UpdatedAt = plan.CreatedAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Plan{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO plans(id, project_id, title, goal, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)`, plan.ID, plan.ProjectID, plan.Title, plan.Goal, plan.CreatedAt, plan.UpdatedAt); err != nil {
		return Plan{}, err
	}
	if err := insertEvent(ctx, tx, projectID, plan.ID, "", "plan_created", plan.Title, actor, plan.CreatedAt); err != nil {
		return Plan{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, plan.CreatedAt, projectID); err != nil {
		return Plan{}, err
	}
	if err := tx.Commit(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (s *Store) ImportPlan(ctx context.Context, projectID string, input PlanImport, actor string) (Plan, []Task, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || len(input.Tasks) == 0 || len(input.Tasks) > 500 {
		return Plan{}, nil, fmt.Errorf("%w: plan title and 1–500 tasks are required", errValidation)
	}
	for index, task := range input.Tasks {
		if strings.TrimSpace(task.Title) == "" {
			return Plan{}, nil, fmt.Errorf("%w: task %d needs a title", errValidation, index+1)
		}
	}
	if _, err := s.getProject(ctx, projectID); err != nil {
		return Plan{}, nil, err
	}
	planID, err := newID("pl")
	if err != nil {
		return Plan{}, nil, err
	}
	plan := Plan{ID: planID, ProjectID: projectID, Title: input.Title, Goal: strings.TrimSpace(input.Goal), CreatedAt: nowUTC()}
	plan.UpdatedAt = plan.CreatedAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Plan{}, nil, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO plans(id, project_id, title, goal, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)`, plan.ID, projectID, plan.Title, plan.Goal, plan.CreatedAt, plan.UpdatedAt); err != nil {
		return Plan{}, nil, err
	}
	if err := insertEvent(ctx, tx, projectID, plan.ID, "", "plan_created", plan.Title, actor, plan.CreatedAt); err != nil {
		return Plan{}, nil, err
	}
	tasks := make([]Task, 0, len(input.Tasks))
	for _, item := range input.Tasks {
		id, err := newID("tk")
		if err != nil {
			return Plan{}, nil, err
		}
		at := nowUTC()
		task := Task{ID: id, ProjectID: projectID, PlanID: plan.ID, Title: strings.TrimSpace(item.Title),
			Description: strings.TrimSpace(item.Description), Status: "todo", CreatedAt: at, UpdatedAt: at}
		if _, err := tx.ExecContext(ctx, `INSERT INTO tasks(id, project_id, plan_id, title, description, status, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, task.ID, projectID, plan.ID, task.Title, task.Description,
			task.Status, task.CreatedAt, task.UpdatedAt); err != nil {
			return Plan{}, nil, err
		}
		if err := insertEvent(ctx, tx, projectID, plan.ID, task.ID, "task_created", task.Title, actor, at); err != nil {
			return Plan{}, nil, err
		}
		tasks = append(tasks, task)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, nowUTC(), projectID); err != nil {
		return Plan{}, nil, err
	}
	if err := tx.Commit(); err != nil {
		return Plan{}, nil, err
	}
	return plan, tasks, nil
}

func (s *Store) UpdatePlan(ctx context.Context, id, title, goal, actor string) (Plan, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Plan{}, fmt.Errorf("%w: plan title is required", errValidation)
	}
	var plan Plan
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, created_at, updated_at FROM plans WHERE id = ?`, id).
		Scan(&plan.ID, &plan.ProjectID, &plan.CreatedAt, &plan.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Plan{}, errNotFound
	}
	if err != nil {
		return Plan{}, err
	}
	previousUpdatedAt := plan.UpdatedAt
	plan.Title, plan.Goal, plan.UpdatedAt = title, strings.TrimSpace(goal), nowUTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Plan{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE plans SET title = ?, goal = ?, updated_at = ? WHERE id = ? AND updated_at = ?`,
		plan.Title, plan.Goal, plan.UpdatedAt, plan.ID, previousUpdatedAt)
	if err != nil {
		return Plan{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return Plan{}, errConflict
	}
	if err := insertEvent(ctx, tx, plan.ProjectID, plan.ID, "", "plan_edited", plan.Title, actor, plan.UpdatedAt); err != nil {
		return Plan{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, plan.UpdatedAt, plan.ProjectID); err != nil {
		return Plan{}, err
	}
	if err := tx.Commit(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (s *Store) LatestPlanID(ctx context.Context, projectID string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM plans WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`, projectID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: this project has no plan", errValidation)
	}
	return id, err
}

func (s *Store) GetPlan(ctx context.Context, id string) (Plan, error) {
	var plan Plan
	err := s.db.QueryRowContext(ctx, `SELECT id, project_id, title, goal, created_at, updated_at FROM plans WHERE id = ?`, id).
		Scan(&plan.ID, &plan.ProjectID, &plan.Title, &plan.Goal, &plan.CreatedAt, &plan.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Plan{}, errNotFound
	}
	return plan, err
}

func (s *Store) CreateTask(ctx context.Context, projectID, planID, title, description, actor string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("%w: task title is required", errValidation)
	}
	if planID == "" {
		var err error
		planID, err = s.LatestPlanID(ctx, projectID)
		if err != nil {
			return Task{}, err
		}
	}
	var planProjectID string
	err := s.db.QueryRowContext(ctx, `SELECT project_id FROM plans WHERE id = ?`, planID).Scan(&planProjectID)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, errNotFound
	}
	if err != nil {
		return Task{}, err
	}
	if planProjectID != projectID {
		return Task{}, fmt.Errorf("%w: plan belongs to another project", errValidation)
	}
	id, err := newID("tk")
	if err != nil {
		return Task{}, err
	}
	task := Task{ID: id, ProjectID: projectID, PlanID: planID, Title: title,
		Description: strings.TrimSpace(description), Status: "todo", CreatedAt: nowUTC()}
	task.UpdatedAt = task.CreatedAt
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO tasks(id, project_id, plan_id, title, description, status, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, task.ID, task.ProjectID, task.PlanID, task.Title,
		task.Description, task.Status, task.CreatedAt, task.UpdatedAt); err != nil {
		return Task{}, err
	}
	if err := insertEvent(ctx, tx, projectID, planID, task.ID, "task_created", task.Title, actor, task.CreatedAt); err != nil {
		return Task{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, task.UpdatedAt, projectID); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return task, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(row taskScanner) (Task, error) {
	var task Task
	var startedAt, completedAt sql.NullString
	err := row.Scan(&task.ID, &task.ProjectID, &task.PlanID, &task.Title, &task.Description,
		&task.Status, &task.CreatedAt, &startedAt, &completedAt, &task.UpdatedAt)
	task.StartedAt, task.CompletedAt = startedAt.String, completedAt.String
	return task, err
}

func (s *Store) ResolveTaskID(ctx context.Context, prefix string) (string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM tasks WHERE substr(id, 1, length(?)) = ? LIMIT 2`, prefix, prefix)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", errNotFound
	}
	if len(ids) > 1 {
		return "", errAmbiguous
	}
	return ids[0], nil
}

func (s *Store) GetTask(ctx context.Context, id string) (Task, error) {
	task, err := scanTask(s.db.QueryRowContext(ctx, `SELECT id, project_id, plan_id, title, description, status,
		created_at, started_at, completed_at, updated_at FROM tasks WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, errNotFound
	}
	return task, err
}

func validStatus(status string) bool {
	switch status {
	case "todo", "doing", "blocked", "done":
		return true
	default:
		return false
	}
}

func (s *Store) PatchTask(ctx context.Context, id string, patch TaskPatch, actor string) (Task, error) {
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return Task{}, err
	}
	previous := task
	if patch.Title != nil {
		task.Title = strings.TrimSpace(*patch.Title)
		if task.Title == "" {
			return Task{}, fmt.Errorf("%w: task title is required", errValidation)
		}
	}
	if patch.Description != nil {
		task.Description = strings.TrimSpace(*patch.Description)
	}
	if patch.PlanID != nil {
		var projectID string
		err := s.db.QueryRowContext(ctx, `SELECT project_id FROM plans WHERE id = ?`, *patch.PlanID).Scan(&projectID)
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, errNotFound
		}
		if err != nil {
			return Task{}, err
		}
		if projectID != task.ProjectID {
			return Task{}, fmt.Errorf("%w: plan belongs to another project", errValidation)
		}
		task.PlanID = *patch.PlanID
	}
	if patch.Status != nil {
		if !validStatus(*patch.Status) {
			return Task{}, fmt.Errorf("%w: unknown task status", errValidation)
		}
		task.Status = *patch.Status
	}
	note := strings.TrimSpace(patch.Note)
	if previous.Status != "done" && task.Status == "done" && note == "" {
		return Task{}, fmt.Errorf("%w: a completion note is required", errValidation)
	}
	if task == previous && note == "" {
		return task, nil
	}
	task.UpdatedAt = nowUTC()
	if task.Status == "doing" && task.StartedAt == "" {
		task.StartedAt = task.UpdatedAt
	}
	if task.Status == "done" && previous.Status != "done" {
		task.CompletedAt = task.UpdatedAt
	} else if task.Status != "done" {
		task.CompletedAt = ""
	}
	kind := "task_edited"
	if task.Status != previous.Status {
		kind = "task_" + task.Status
	} else if task.Title == previous.Title && task.Description == previous.Description && task.PlanID == previous.PlanID {
		kind = "task_note"
	}
	if note == "" {
		note = task.Title
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Task{}, err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE tasks SET plan_id = ?, title = ?, description = ?, status = ?,
		started_at = ?, completed_at = ?, updated_at = ? WHERE id = ? AND updated_at = ?`, task.PlanID, task.Title,
		task.Description, task.Status, nullString(task.StartedAt), nullString(task.CompletedAt), task.UpdatedAt, task.ID, previous.UpdatedAt)
	if err != nil {
		return Task{}, err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return Task{}, errConflict
	}
	if err := insertEvent(ctx, tx, task.ProjectID, task.PlanID, task.ID, kind, note, actor, task.UpdatedAt); err != nil {
		return Task{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`, task.UpdatedAt, task.ProjectID); err != nil {
		return Task{}, err
	}
	if err := tx.Commit(); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (s *Store) AddNote(ctx context.Context, id, note, actor string) (Task, error) {
	note = strings.TrimSpace(note)
	if note == "" {
		return Task{}, fmt.Errorf("%w: note is required", errValidation)
	}
	return s.PatchTask(ctx, id, TaskPatch{Note: note}, actor)
}

func (s *Store) NextTasks(ctx context.Context, projectID string) ([]Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, project_id, plan_id, title, description, status,
		created_at, started_at, completed_at, updated_at FROM tasks
		WHERE project_id = ? AND status = 'todo' ORDER BY created_at LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []Task{}
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}
