package main

import (
	"context"
	"encoding/json"
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
