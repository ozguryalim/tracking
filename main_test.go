package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskCommandStaysWithinCurrentProject(t *testing.T) {
	store, _ := testStore(t)
	first := mustProject(t, store, "First")
	second := mustProject(t, store, "Second")
	secondPlan := mustPlan(t, store, second.ID)
	secondTask := mustTask(t, store, second.ID, secondPlan.ID, "Private task")

	projectDir := t.TempDir()
	manifest := projectManifest{ID: first.ID, Name: first.Name}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(manifestFile(projectDir)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestFile(projectDir), data, 0644); err != nil {
		t.Fatal(err)
	}
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	err = runTask(context.Background(), store, []string{"done", secondTask.ID, "--note", "Wrong project"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("task command modified a task outside the current project")
	}
	actual, err := store.GetTask(context.Background(), secondTask.ID)
	if err != nil {
		t.Fatal(err)
	}
	if actual.Status != "todo" || actual.CompletedAt != "" {
		t.Fatalf("other project's task changed: %+v", actual)
	}
}

func TestShowAndDeleteCommands(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Current")
	plan := mustPlan(t, store, project.ID)
	first := mustTask(t, store, project.ID, plan.ID, "First task")
	second := mustTask(t, store, project.ID, plan.ID, "Second task")
	other := mustProject(t, store, "Other")
	otherPlan := mustPlan(t, store, other.ID)
	otherTask := mustTask(t, store, other.ID, otherPlan.ID, "Other task")

	projectDir := t.TempDir()
	data, err := json.Marshal(projectManifest{ID: project.ID, Name: project.Name})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(manifestFile(projectDir)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestFile(projectDir), data, 0644); err != nil {
		t.Fatal(err)
	}
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	var output bytes.Buffer
	if err := runPlan(ctx, store, []string{"show", plan.ID[:8]}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	var planDetail struct {
		Plan  Plan   `json:"plan"`
		Tasks []Task `json:"tasks"`
	}
	if err := json.Unmarshal(output.Bytes(), &planDetail); err != nil {
		t.Fatal(err)
	}
	if planDetail.Plan.ID != plan.ID || len(planDetail.Tasks) != 2 {
		t.Fatalf("unexpected plan detail: %+v", planDetail)
	}

	output.Reset()
	if err := runTask(ctx, store, []string{"show", first.ID[:8]}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	var taskDetail struct {
		Task   Task    `json:"task"`
		Events []Event `json:"events"`
	}
	if err := json.Unmarshal(output.Bytes(), &taskDetail); err != nil {
		t.Fatal(err)
	}
	if taskDetail.Task.ID != first.ID || len(taskDetail.Events) != 1 || taskDetail.Events[0].Kind != "task_created" {
		t.Fatalf("unexpected task detail: %+v", taskDetail)
	}

	if err := runPlan(ctx, store, []string{"edit", plan.ID[:8], "--goal", "Updated goal"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	updated, err := store.GetPlan(ctx, plan.ID)
	if err != nil || updated.Goal != "Updated goal" {
		t.Fatalf("plan edit failed: %+v, %v", updated, err)
	}
	if err := runTask(ctx, store, []string{"delete", first.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("task deletion without --yes succeeded")
	}
	if err := runPlan(ctx, store, []string{"delete", plan.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("plan deletion without --yes succeeded")
	}
	if err := runPlan(ctx, store, []string{"edit", plan.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("plan edit without fields succeeded")
	}
	if err := runTask(ctx, store, []string{"edit", first.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("task edit without fields succeeded")
	}
	if _, err := store.GetTask(ctx, first.ID); err != nil {
		t.Fatalf("unconfirmed task deletion changed data: %v", err)
	}
	if err := runTask(ctx, store, []string{"show", otherTask.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("showed a task outside the current project")
	}
	if err := runPlan(ctx, store, []string{"show", otherPlan.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("showed a plan outside the current project")
	}
	if err := runTask(ctx, store, []string{"delete", otherTask.ID, "--yes"}, io.Discard, io.Discard); err == nil {
		t.Fatal("deleted a task outside the current project")
	}
	if err := runPlan(ctx, store, []string{"delete", otherPlan.ID, "--yes"}, io.Discard, io.Discard); err == nil {
		t.Fatal("deleted a plan outside the current project")
	}

	if err := runTask(ctx, store, []string{"delete", first.ID[:8], "--yes"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetTask(ctx, first.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("deleted task still exists: %v", err)
	}
	if err := runPlan(ctx, store, []string{"delete", plan.ID[:8], "--yes"}, io.Discard, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetTask(ctx, second.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("plan task still exists: %v", err)
	}
	if _, err := store.GetPlan(ctx, plan.ID); !errors.Is(err, errNotFound) {
		t.Fatalf("deleted plan still exists: %v", err)
	}
	if _, err := store.GetTask(ctx, otherTask.ID); err != nil {
		t.Fatalf("other project task changed: %v", err)
	}
}

func TestAttachDashboardProjectToDirectory(t *testing.T) {
	store, _ := testStore(t)
	project := mustProject(t, store, "Dashboard project")
	other := mustProject(t, store, "Other project")
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldDir) })

	if err := runAttach(context.Background(), store, []string{project.ID}, io.Discard); err != nil {
		t.Fatal(err)
	}
	manifest, root, err := readManifest(".")
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ID != project.ID || manifest.Name != project.Name {
		t.Fatalf("incorrect project manifest: %+v", manifest)
	}
	linked, err := store.getProject(context.Background(), project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if linked.Path != root {
		t.Fatalf("database path = %q, manifest path = %q", linked.Path, root)
	}
	if err := runAttach(context.Background(), store, []string{project.ID}, io.Discard); err != nil {
		t.Fatalf("attaching the same project should be idempotent: %v", err)
	}
	if err := runAttach(context.Background(), store, []string{other.ID}, io.Discard); err == nil {
		t.Fatal("attached a different project to an existing project directory")
	}
}
