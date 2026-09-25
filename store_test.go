package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func testStore(t *testing.T) (*Store, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tracking.db")
	store, err := openStore(path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store, path
}

func mustProject(t *testing.T, store *Store, name string) Project {
	t.Helper()
	project, err := store.CreateProject(context.Background(), name, "", "test")
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func mustPlan(t *testing.T, store *Store, projectID string) Plan {
	t.Helper()
	plan, err := store.CreatePlan(context.Background(), projectID, "Release", "Ship safely", "test")
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}
	return plan
}

func mustTask(t *testing.T, store *Store, projectID, planID, title string) Task {
	t.Helper()
	task, err := store.CreateTask(context.Background(), projectID, planID, title, "", "test")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}

func TestProjectIsolationAndCounters(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	first := mustProject(t, store, "First")
	second := mustProject(t, store, "Second")
	firstPlan := mustPlan(t, store, first.ID)
	secondPlan := mustPlan(t, store, second.ID)
	firstTask := mustTask(t, store, first.ID, firstPlan.ID, "First task")
	secondTask := mustTask(t, store, second.ID, secondPlan.ID, "Second task")

	if _, err := store.CreateTask(ctx, first.ID, secondPlan.ID, "Wrong project", "", "test"); !errors.Is(err, errValidation) {
		t.Fatalf("cross-project task creation: got %v, want validation error", err)
	}
	if _, err := store.PatchTask(ctx, firstTask.ID, TaskPatch{PlanID: &secondPlan.ID}, "test"); !errors.Is(err, errValidation) {
		t.Fatalf("cross-project task move: got %v, want validation error", err)
	}
	status := "done"
	if _, err := store.PatchTask(ctx, firstTask.ID, TaskPatch{Status: &status, Note: "Validated on the device"}, "test"); err != nil {
		t.Fatalf("complete first task: %v", err)
	}
	firstDetail, err := store.ProjectDetail(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	secondDetail, err := store.ProjectDetail(ctx, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(firstDetail.Plans) != 1 || len(firstDetail.Tasks) != 1 || firstDetail.Tasks[0].ID != firstTask.ID {
		t.Fatalf("first project leaked records: %+v", firstDetail)
	}
	if len(secondDetail.Plans) != 1 || len(secondDetail.Tasks) != 1 || secondDetail.Tasks[0].ID != secondTask.ID {
		t.Fatalf("second project leaked records: %+v", secondDetail)
	}
	if firstDetail.Project.DoneCount != 1 || secondDetail.Project.DoneCount != 0 {
		t.Fatalf("wrong project counters: first=%+v second=%+v", firstDetail.Project, secondDetail.Project)
	}
	for _, event := range firstDetail.Events {
		if event.ProjectID != first.ID || event.TaskID == secondTask.ID {
			t.Fatalf("foreign event in first project: %+v", event)
		}
	}
	projects, err := store.ListProjects(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	for _, project := range projects {
		if project.ID == first.ID && (project.TaskCount != 1 || project.DoneCount != 1) {
			t.Fatalf("first project list counters: %+v", project)
		}
		if project.ID == second.ID && (project.TaskCount != 1 || project.DoneCount != 0) {
			t.Fatalf("second project list counters: %+v", project)
		}
	}
}

func TestImportPlanIsAtomic(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Import")

	_, _, err := store.ImportPlan(ctx, project.ID, PlanImport{Title: "Bad plan", Tasks: []ImportedTask{{Title: "Valid"}, {Title: " "}}}, "test")
	if !errors.Is(err, errValidation) {
		t.Fatalf("invalid import: got %v, want validation error", err)
	}
	// Fail after the plan and first task have been inserted. The transaction must
	// roll back all of those rows and their events.
	_, err = store.db.Exec(`CREATE TRIGGER reject_task BEFORE INSERT ON tasks WHEN NEW.title = 'Reject' BEGIN SELECT RAISE(ABORT, 'test failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = store.ImportPlan(ctx, project.ID, PlanImport{Title: "Interrupted plan", Tasks: []ImportedTask{{Title: "Valid"}, {Title: "Reject"}}}, "test")
	if err == nil {
		t.Fatal("import should have failed on the second task")
	}
	detail, err := store.ProjectDetail(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Plans) != 0 || len(detail.Tasks) != 0 || len(detail.Events) != 1 || detail.Events[0].Kind != "project_created" {
		t.Fatalf("partial import survived rollback: %+v", detail)
	}
}

func TestTaskLifecycleAndPersistence(t *testing.T) {
	store, path := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Lifecycle")
	plan := mustPlan(t, store, project.ID)
	task := mustTask(t, store, project.ID, plan.ID, "Implement feature")
	if task.StartedAt != "" || task.CompletedAt != "" {
		t.Fatalf("new task unexpectedly has work timestamps: %+v", task)
	}

	doing := "doing"
	started, err := store.PatchTask(ctx, task.ID, TaskPatch{Status: &doing, Note: "Started implementation"}, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if started.Status != "doing" || started.StartedAt == "" || started.CompletedAt != "" {
		t.Fatalf("start timestamps: %+v", started)
	}
	if _, err := time.Parse(time.RFC3339Nano, started.StartedAt); err != nil {
		t.Fatalf("start time format: %v", err)
	}
	if _, err := store.AddNote(ctx, task.ID, "Implementation complete", "codex"); err != nil {
		t.Fatal(err)
	}
	done := "done"
	if _, err := store.PatchTask(ctx, task.ID, TaskPatch{Status: &done}, "codex"); !errors.Is(err, errValidation) {
		t.Fatalf("completion without note: got %v, want validation error", err)
	}
	completed, err := store.PatchTask(ctx, task.ID, TaskPatch{Status: &done, Note: "Targeted tests passed"}, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "done" || completed.CompletedAt == "" || completed.StartedAt != started.StartedAt {
		t.Fatalf("completion timestamps: %+v", completed)
	}
	startTime, _ := time.Parse(time.RFC3339Nano, completed.StartedAt)
	completionTime, err := time.Parse(time.RFC3339Nano, completed.CompletedAt)
	if err != nil || completionTime.Before(startTime) {
		t.Fatalf("completion time %q precedes start %q: %v", completed.CompletedAt, completed.StartedAt, err)
	}

	detail, err := store.ProjectDetail(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantKinds := map[string]string{"task_doing": "Started implementation", "task_note": "Implementation complete", "task_done": "Targeted tests passed"}
	for kind, note := range wantKinds {
		found := false
		for _, event := range detail.Events {
			if event.TaskID == task.ID && event.Kind == kind && event.Note == note && event.Actor == "codex" && event.OccurredAt != "" {
				found = true
			}
		}
		if !found {
			t.Errorf("missing %s event with note %q and actor codex", kind, note)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := openStore(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reopened.Close() })
	persisted, err := reopened.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Status != "done" || persisted.CompletedAt != completed.CompletedAt || persisted.StartedAt != started.StartedAt {
		t.Fatalf("task changed after reopen: %+v", persisted)
	}
	reopenedStatus := "todo"
	reopenedTask, err := reopened.PatchTask(ctx, task.ID, TaskPatch{Status: &reopenedStatus, Note: "More work needed"}, "claude")
	if err != nil {
		t.Fatal(err)
	}
	if reopenedTask.CompletedAt != "" || reopenedTask.StartedAt != started.StartedAt {
		t.Fatalf("reopen timestamps: %+v", reopenedTask)
	}
	next, err := reopened.NextTasks(ctx, project.ID)
	if err != nil || len(next) != 1 || next[0].ID != task.ID {
		t.Fatalf("reopened task should be next: tasks=%+v err=%v", next, err)
	}
}

func TestTaskIDPrefixTreatsWildcardCharactersLiterally(t *testing.T) {
	store, _ := testStore(t)
	project := mustProject(t, store, "Prefixes")
	plan := mustPlan(t, store, project.ID)
	task := mustTask(t, store, project.ID, plan.ID, "One task")
	if resolved, err := store.ResolveTaskID(context.Background(), task.ID[:8]); err != nil || resolved != task.ID {
		t.Fatalf("ordinary prefix did not resolve task: id=%q err=%v", resolved, err)
	}
	for _, wildcard := range []string{"%", "tk_"} {
		if resolved, err := store.ResolveTaskID(context.Background(), wildcard); err == nil {
			t.Errorf("wildcard %q unexpectedly resolved task %q", wildcard, resolved)
		}
	}
}
