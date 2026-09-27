package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// chdirToProject links a temporary directory to project and works inside it.
func chdirToProject(t *testing.T, project Project) {
	t.Helper()
	dir := t.TempDir()
	data, err := json.Marshal(projectManifest{ID: project.ID, Name: project.Name})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(manifestFile(dir)), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestFile(dir), data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

func TestTaskReviewCommandIsListedButNotNext(t *testing.T) {
	store, _ := testStore(t)
	ctx := context.Background()
	project := mustProject(t, store, "Review")
	plan := mustPlan(t, store, project.ID)
	reviewed := mustTask(t, store, project.ID, plan.ID, "Awaiting approval")
	ready := mustTask(t, store, project.ID, plan.ID, "Ready task")
	chdirToProject(t, project)

	if err := runTask(ctx, store, []string{"review", reviewed.ID}, io.Discard, io.Discard); err == nil {
		t.Fatal("review without a note succeeded")
	}
	var output bytes.Buffer
	if err := runTask(ctx, store, []string{"review", reviewed.ID, "--note", "Acceptance tests pass"}, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if want := reviewed.ID + " review: " + reviewed.Title; !strings.Contains(output.String(), want) {
		t.Fatalf("review output %q does not contain %q", output.String(), want)
	}
	output.Reset()
	if err := runNext(ctx, store, nil, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), ready.ID) || strings.Contains(output.String(), reviewed.ID) {
		t.Fatalf("next should list only the ready task:\n%s", output.String())
	}
	output.Reset()
	if err := runStatus(ctx, store, nil, &output, io.Discard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "0/2 done, 1 in review") || !strings.Contains(output.String(), "review  "+reviewed.ID) {
		t.Fatalf("status does not show the task in review:\n%s", output.String())
	}
}

func TestTrackingSkillDefersCompletionToProjectRules(t *testing.T) {
	skill := string(trackingSkill)
	if !strings.Contains(skill, "definition of done") || !strings.Contains(skill, "tracking task review ID") {
		t.Fatal("skill does not point agents to the project's definition of done and the review status")
	}
	if strings.Contains(skill, "Do not require separate approval") {
		t.Fatal("skill still tells agents to skip approval")
	}
	var usage bytes.Buffer
	printUsage(&usage)
	for _, match := range regexp.MustCompile("`tracking ([a-z]+(?: [a-z]+)?)").FindAllStringSubmatch(skill, -1) {
		// help prints the usage itself, so it is not listed there.
		listed := regexp.MustCompile(`(?m)^  tracking ` + match[1] + `( |$)`)
		if match[1] != "help" && !listed.MatchString(usage.String()) {
			t.Errorf("skill mentions tracking %s, which the usage does not list", match[1])
		}
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
