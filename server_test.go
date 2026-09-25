package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func apiRequest(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse[T any](t *testing.T, response *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()
	if response.Code != wantStatus {
		t.Fatalf("status=%d, want %d; body=%s", response.Code, wantStatus, response.Body.String())
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("unexpected JSON content type: %q", contentType)
	}
	var result T
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, response.Body.String())
	}
	return result
}

func TestHTTPProjectPlanTaskFlowAndAssets(t *testing.T) {
	store, _ := testStore(t)
	handler := newHTTPHandler(store)

	projectResult := decodeResponse[struct {
		Project Project `json:"project"`
	}](t,
		apiRequest(t, handler, http.MethodPost, "/api/projects", `{"name":"  Acme  "}`), http.StatusCreated)
	project := projectResult.Project
	if project.ID == "" || project.Name != "Acme" {
		t.Fatalf("project was not created correctly: %+v", project)
	}
	planPath := "/api/projects/" + project.ID + "/plans"
	planResult := decodeResponse[struct {
		Plan Plan `json:"plan"`
	}](t,
		apiRequest(t, handler, http.MethodPost, planPath, `{"title":"Launch","goal":"Ship first version"}`), http.StatusCreated)
	plan := planResult.Plan
	if plan.ID == "" || plan.ProjectID != project.ID {
		t.Fatalf("plan was not linked to project: %+v", plan)
	}
	updatedPlan := decodeResponse[struct {
		Plan Plan `json:"plan"`
	}](t,
		apiRequest(t, handler, http.MethodPatch, "/api/plans/"+plan.ID, `{"title":"Launch v1","goal":"Ship"}`), http.StatusOK)
	if updatedPlan.Plan.Title != "Launch v1" {
		t.Fatalf("plan update lost: %+v", updatedPlan.Plan)
	}
	taskResult := decodeResponse[struct {
		Task Task `json:"task"`
	}](t,
		apiRequest(t, handler, http.MethodPost, "/api/projects/"+project.ID+"/tasks", `{"title":"Build UI","description":"Dashboard"}`), http.StatusCreated)
	task := taskResult.Task
	if task.ID == "" || task.PlanID != plan.ID || task.Status != "todo" {
		t.Fatalf("task was not linked to latest plan: %+v", task)
	}
	started := decodeResponse[struct {
		Task Task `json:"task"`
	}](t,
		apiRequest(t, handler, http.MethodPatch, "/api/tasks/"+task.ID, `{"status":"doing","note":"Started UI"}`), http.StatusOK)
	if started.Task.Status != "doing" || started.Task.StartedAt == "" {
		t.Fatalf("task start failed: %+v", started.Task)
	}
	decodeResponse[struct {
		Task Task `json:"task"`
	}](t,
		apiRequest(t, handler, http.MethodPost, "/api/tasks/"+task.ID+"/notes", `{"note":"Main view ready"}`), http.StatusCreated)
	decodeResponse[struct {
		Error string `json:"error"`
	}](t,
		apiRequest(t, handler, http.MethodPatch, "/api/tasks/"+task.ID, `{"status":"done"}`), http.StatusBadRequest)
	finished := decodeResponse[struct {
		Task Task `json:"task"`
	}](t,
		apiRequest(t, handler, http.MethodPatch, "/api/tasks/"+task.ID, `{"status":"done","note":"Checked dashboard"}`), http.StatusOK)
	if finished.Task.CompletedAt == "" {
		t.Fatalf("completion time absent: %+v", finished.Task)
	}
	detail := decodeResponse[ProjectDetail](t,
		apiRequest(t, handler, http.MethodGet, "/api/projects/"+project.ID, ""), http.StatusOK)
	if detail.Project.DoneCount != 1 || len(detail.Plans) != 1 || len(detail.Tasks) != 1 || len(detail.Events) < 6 {
		t.Fatalf("unexpected project detail: %+v", detail)
	}
	projects := decodeResponse[struct {
		Projects []Project `json:"projects"`
	}](t,
		apiRequest(t, handler, http.MethodGet, "/api/projects", ""), http.StatusOK)
	if len(projects.Projects) != 1 || projects.Projects[0].ID != project.ID || projects.Projects[0].DoneCount != 1 {
		t.Fatalf("unexpected projects list: %+v", projects.Projects)
	}
	page := apiRequest(t, handler, http.MethodGet, "/", "")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "<title>Tracking") {
		t.Fatalf("dashboard not served: status=%d body=%s", page.Code, page.Body.String())
	}
	asset := apiRequest(t, handler, http.MethodGet, "/app.js", "")
	if asset.Code != http.StatusOK || asset.Body.Len() == 0 {
		t.Fatalf("dashboard script not served: status=%d", asset.Code)
	}
}

func TestHTTPRejectsInvalidInputAndUnknownRecords(t *testing.T) {
	store, _ := testStore(t)
	handler := newHTTPHandler(store)
	for _, body := range []string{`{"name":""}`, `{"name":"Good","unknown":true}`, `{"name":"Good","path":"/workspace/acme"}`, `{"name":"One"}{"name":"Two"}`} {
		response := apiRequest(t, handler, http.MethodPost, "/api/projects", body)
		decodeResponse[struct {
			Error string `json:"error"`
		}](t, response, http.StatusBadRequest)
	}
	missing := decodeResponse[struct {
		Error string `json:"error"`
	}](t,
		apiRequest(t, handler, http.MethodGet, "/api/projects/pr-missing", ""), http.StatusNotFound)
	if missing.Error == "" {
		t.Fatal("missing project should return an error message")
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(`{"name":"No content type"}`))
	handler.ServeHTTP(response, request)
	decodeResponse[struct {
		Error string `json:"error"`
	}](t, response, http.StatusBadRequest)
}

func TestHTTPTaskEventsPreserveFullTaskHistory(t *testing.T) {
	store, _ := testStore(t)
	project := mustProject(t, store, "History")
	plan := mustPlan(t, store, project.ID)
	target := mustTask(t, store, project.ID, plan.ID, "Original task")
	_ = mustTask(t, store, project.ID, plan.ID, "Busy task")
	foreign := mustProject(t, store, "Other project")
	foreignPlan := mustPlan(t, store, foreign.ID)
	foreignTask := mustTask(t, store, foreign.ID, foreignPlan.ID, "Foreign task")

	handler := newHTTPHandler(store)
	decodeResponse[struct {
		Task Task `json:"task"`
	}](t, apiRequest(t, handler, http.MethodPost, "/api/tasks/"+target.ID+"/notes", `{"note":"Original task note"}`), http.StatusCreated)
	decodeResponse[struct {
		Task Task `json:"task"`
	}](t, apiRequest(t, handler, http.MethodPost, "/api/tasks/"+foreignTask.ID+"/notes", `{"note":"Foreign note"}`), http.StatusCreated)

	// A long-running task's first notes fall outside the project summary's
	// 500-event window, but remain available in the task's full history.
	tx, err := store.db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	base := time.Now().UTC().Add(time.Minute)
	for index := 0; index < 501; index++ {
		id, err := newID("ev")
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.Exec(`INSERT INTO events(id, project_id, plan_id, task_id, kind, note, actor, occurred_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, id, project.ID, plan.ID, target.ID,
			"task_note", fmt.Sprintf("Later note %d", index), "test", base.Add(time.Duration(index)*time.Nanosecond).Format("2006-01-02T15:04:05.000000000Z"))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	summary := decodeResponse[ProjectDetail](t,
		apiRequest(t, handler, http.MethodGet, "/api/projects/"+project.ID, ""), http.StatusOK)
	if len(summary.Events) != 500 {
		t.Fatalf("project summary has %d events, want 500", len(summary.Events))
	}
	for _, event := range summary.Events {
		if event.Note == "Original task note" {
			t.Fatal("old target note unexpectedly remained in project summary")
		}
	}

	result := decodeResponse[struct {
		Events []Event `json:"events"`
	}](t, apiRequest(t, handler, http.MethodGet, "/api/tasks/"+target.ID+"/events", ""), http.StatusOK)
	if len(result.Events) != 503 {
		t.Fatalf("task history has %d events, want 503", len(result.Events))
	}
	for _, event := range result.Events {
		if event.TaskID != target.ID || event.ProjectID != project.ID {
			t.Fatalf("task history includes another task or project: %+v", event)
		}
	}
	if result.Events[0].Kind != "task_note" || result.Events[0].Note != "Later note 500" ||
		result.Events[len(result.Events)-2].Note != "Original task note" {
		t.Fatalf("task history is not newest first: %+v", result.Events)
	}
	decodeResponse[struct {
		Error string `json:"error"`
	}](t, apiRequest(t, handler, http.MethodGet, "/api/tasks/tk-missing/events", ""), http.StatusNotFound)
}
