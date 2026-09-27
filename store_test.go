package main

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
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
	for _, empty := range []string{"", " "} {
		if _, err := store.ResolveTaskID(context.Background(), empty); !errors.Is(err, errValidation) {
			t.Errorf("empty prefix %q: got %v, want validation error", empty, err)
		}
	}
}

func TestPlanIDPrefixAndPlanTasks(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Plans")
	first := mustPlan(t, store, project.ID)
	second := mustPlan(t, store, project.ID)
	firstTask := mustTask(t, store, project.ID, first.ID, "First task")
	_ = mustTask(t, store, project.ID, second.ID, "Second task")
	resolved, err := store.ResolvePlanID(ctx, first.ID[:len(first.ID)-1])
	if err != nil || resolved != first.ID {
		t.Fatalf("plan prefix: id=%q err=%v", resolved, err)
	}
	if _, err := store.ResolvePlanID(ctx, "pl-"); !errors.Is(err, errAmbiguous) {
		t.Fatalf("shared plan prefix: got %v, want ambiguity", err)
	}
	for _, prefix := range []string{"", " ", "%", "pl_"} {
		if _, err := store.ResolvePlanID(ctx, prefix); err == nil {
			t.Errorf("invalid plan prefix %q resolved", prefix)
		}
	}
	tasks, err := store.PlanTasks(ctx, first.ID)
	if err != nil || len(tasks) != 1 || tasks[0].ID != firstTask.ID {
		t.Fatalf("plan tasks: tasks=%+v err=%v", tasks, err)
	}
	if _, err := store.PlanTasks(ctx, "missing"); !errors.Is(err, errNotFound) {
		t.Fatalf("missing plan tasks: got %v, want not found", err)
	}
}

func TestDeleteTaskRemovesHistoryAndPreservesOtherRecords(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "First")
	plan := mustPlan(t, store, project.ID)
	target := mustTask(t, store, project.ID, plan.ID, "Delete me")
	kept := mustTask(t, store, project.ID, plan.ID, "Keep me")
	otherProject := mustProject(t, store, "Second")
	otherPlan := mustPlan(t, store, otherProject.ID)
	otherTask := mustTask(t, store, otherProject.ID, otherPlan.ID, "Other project")
	if _, err := store.AddNote(ctx, target.ID, "Old work", "test"); err != nil {
		t.Fatal(err)
	}
	before, err := store.TaskEvents(ctx, target.ID)
	if err != nil || len(before) < 2 {
		t.Fatalf("target history: events=%+v err=%v", before, err)
	}
	if err := store.DeleteTask(ctx, target.ID, "agent"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetTask(ctx, target.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("deleted task: got %v, want not found", err)
	}
	if _, err := store.GetTask(ctx, kept.ID); err != nil {
		t.Fatalf("other task was deleted: %v", err)
	}
	if _, err := store.GetTask(ctx, otherTask.ID); err != nil {
		t.Fatalf("other project task was deleted: %v", err)
	}
	detail, err := store.ProjectDetail(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Project.TaskCount != 1 || len(detail.Tasks) != 1 || detail.Tasks[0].ID != kept.ID {
		t.Fatalf("task count after deletion: %+v", detail)
	}
	var deletion *Event
	for i := range detail.Events {
		event := &detail.Events[i]
		for _, old := range before {
			if event.ID == old.ID {
				t.Fatalf("deleted task history survived: %+v", *event)
			}
		}
		if event.Kind == "task_deleted" {
			deletion = event
		}
	}
	if deletion == nil || deletion.TaskID != "" || deletion.PlanID != plan.ID || deletion.Actor != "agent" ||
		!strings.Contains(deletion.Note, target.ID) || !strings.Contains(deletion.Note, target.Title) ||
		detail.Project.UpdatedAt != deletion.OccurredAt {
		t.Fatalf("task deletion event or project timestamp: event=%+v project=%+v", deletion, detail.Project)
	}
	if _, err := store.TaskEvents(ctx, kept.ID); err != nil {
		t.Fatalf("other task history lost: %v", err)
	}
	if err := store.DeleteTask(ctx, target.ID, "agent"); !errors.Is(err, errNotFound) {
		t.Fatalf("delete missing task: got %v, want not found", err)
	}
}

func TestDeletePlanRemovesOwnedHistoryAndPreservesMovedTask(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "First")
	target := mustPlan(t, store, project.ID)
	keptPlan := mustPlan(t, store, project.ID)
	first := mustTask(t, store, project.ID, target.ID, "Delete first")
	second := mustTask(t, store, project.ID, target.ID, "Delete second")
	moved := mustTask(t, store, project.ID, target.ID, "Move me")
	kept := mustTask(t, store, project.ID, keptPlan.ID, "Keep me")
	otherProject := mustProject(t, store, "Second")
	otherPlan := mustPlan(t, store, otherProject.ID)
	otherTask := mustTask(t, store, otherProject.ID, otherPlan.ID, "Other project")
	if _, err := store.AddNote(ctx, first.ID, "Old work", "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PatchTask(ctx, moved.ID, TaskPatch{PlanID: &keptPlan.ID}, "test"); err != nil {
		t.Fatal(err)
	}
	movedHistory, err := store.TaskEvents(ctx, moved.ID)
	if err != nil {
		t.Fatal(err)
	}
	firstHistory, err := store.TaskEvents(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	count, err := store.DeletePlan(ctx, target.ID, "agent")
	if err != nil || count != 2 {
		t.Fatalf("delete plan: count=%d err=%v", count, err)
	}
	if _, err := store.GetPlan(ctx, target.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("deleted plan: got %v, want not found", err)
	}
	for _, id := range []string{first.ID, second.ID} {
		if _, err := store.GetTask(ctx, id); !errors.Is(err, errNotFound) {
			t.Fatalf("deleted plan task %s: got %v, want not found", id, err)
		}
	}
	for _, id := range []string{moved.ID, kept.ID, otherTask.ID} {
		if _, err := store.GetTask(ctx, id); err != nil {
			t.Fatalf("unrelated task %s was deleted: %v", id, err)
		}
	}
	detail, err := store.ProjectDetail(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Plans) != 1 || detail.Plans[0].ID != keptPlan.ID || detail.Project.TaskCount != 2 {
		t.Fatalf("project records after plan deletion: %+v", detail)
	}
	var deletion *Event
	for i := range detail.Events {
		event := &detail.Events[i]
		if event.PlanID == target.ID {
			t.Fatalf("deleted plan association survived: %+v", *event)
		}
		for _, old := range firstHistory {
			if event.ID == old.ID {
				t.Fatalf("deleted task history survived: %+v", *event)
			}
		}
		if event.Kind == "plan_deleted" {
			deletion = event
		}
	}
	if deletion == nil || deletion.PlanID != "" || deletion.TaskID != "" || deletion.Actor != "agent" ||
		!strings.Contains(deletion.Note, target.ID) || !strings.Contains(deletion.Note, target.Title) ||
		!strings.Contains(deletion.Note, "2 tasks") || detail.Project.UpdatedAt != deletion.OccurredAt {
		t.Fatalf("plan deletion event or project timestamp: event=%+v project=%+v", deletion, detail.Project)
	}
	afterMoved, err := store.TaskEvents(ctx, moved.ID)
	if err != nil || len(afterMoved) != len(movedHistory) {
		t.Fatalf("moved task history: before=%+v after=%+v err=%v", movedHistory, afterMoved, err)
	}
	if count, err := store.DeletePlan(ctx, target.ID, "agent"); count != 0 || !errors.Is(err, errNotFound) {
		t.Fatalf("delete missing plan: count=%d err=%v", count, err)
	}
}

func TestDeleteRollsBackWhenDatabaseRejectsIt(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Rollback")
	plan := mustPlan(t, store, project.ID)
	task := mustTask(t, store, project.ID, plan.ID, "Keep history")
	if _, err := store.AddNote(ctx, task.ID, "Existing note", "test"); err != nil {
		t.Fatal(err)
	}
	before, err := store.TaskEvents(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`CREATE TRIGGER reject_task_delete BEFORE DELETE ON tasks BEGIN SELECT RAISE(ABORT, 'test failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err := store.DeleteTask(ctx, task.ID, "agent"); err == nil {
		t.Fatal("task deletion unexpectedly succeeded")
	}
	after, err := store.TaskEvents(ctx, task.ID)
	if err != nil || len(after) != len(before) {
		t.Fatalf("task history did not roll back: before=%+v after=%+v err=%v", before, after, err)
	}
	if _, err := store.db.Exec(`DROP TRIGGER reject_task_delete`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`CREATE TRIGGER reject_plan_delete BEFORE DELETE ON plans BEGIN SELECT RAISE(ABORT, 'test failure'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DeletePlan(ctx, plan.ID, "agent"); err == nil {
		t.Fatal("plan deletion unexpectedly succeeded")
	}
	after, err = store.TaskEvents(ctx, task.ID)
	if err != nil || len(after) != len(before) {
		t.Fatalf("plan deletion lost task history: before=%+v after=%+v err=%v", before, after, err)
	}
	if _, err := store.GetPlan(ctx, plan.ID); err != nil {
		t.Fatalf("plan was deleted despite rollback: %v", err)
	}
}
