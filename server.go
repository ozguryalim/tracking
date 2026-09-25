package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed web/*
var webFiles embed.FS

func runServer(store *Store, command string, args []string, out, errOut io.Writer) error {
	flags := newFlags(command, errOut)
	listenAddress := flags.String("listen", "127.0.0.1:4157", "HTTP listen address")
	managed := flags.Bool("managed", false, "managed background process")
	if err := flags.Parse(args); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		return err
	}
	url := "http://" + listener.Addr().String()
	fmt.Fprintf(out, "Tracking dashboard: %s\n", url)
	server := &http.Server{Handler: newHTTPHandlerWithRuntime(store, *managed), ReadHeaderTimeout: 5 * time.Second}
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func newHTTPHandler(store *Store) http.Handler {
	return newHTTPHandlerWithRuntime(store, false)
}

func newHTTPHandlerWithRuntime(store *Store, managed bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]any{
			"ok": true, "app": "tracking", "db_key": store.key,
			"managed": managed, "pid": os.Getpid(),
		})
	})
	mux.HandleFunc("GET /api/projects", func(w http.ResponseWriter, r *http.Request) {
		projects, err := store.ListProjects(r.Context())
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"projects": projects})
	})
	mux.HandleFunc("POST /api/projects", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name string `json:"name"`
		}
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, err)
			return
		}
		project, err := store.CreateProject(r.Context(), input.Name, "", "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"project": project})
	})
	mux.HandleFunc("GET /api/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		detail, err := store.ProjectDetail(r.Context(), r.PathValue("id"))
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, detail)
	})
	mux.HandleFunc("POST /api/projects/{id}/plans", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title string `json:"title"`
			Goal  string `json:"goal"`
		}
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, err)
			return
		}
		plan, err := store.CreatePlan(r.Context(), r.PathValue("id"), input.Title, input.Goal, "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"plan": plan})
	})
	mux.HandleFunc("PATCH /api/plans/{id}", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title string `json:"title"`
			Goal  string `json:"goal"`
		}
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, err)
			return
		}
		plan, err := store.UpdatePlan(r.Context(), r.PathValue("id"), input.Title, input.Goal, "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"plan": plan})
	})
	mux.HandleFunc("POST /api/projects/{id}/tasks", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			PlanID      string `json:"plan_id"`
		}
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, err)
			return
		}
		task, err := store.CreateTask(r.Context(), r.PathValue("id"), input.PlanID,
			input.Title, input.Description, "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"task": task})
	})
	mux.HandleFunc("PATCH /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		var patch TaskPatch
		if err := decodeJSON(r, &patch); err != nil {
			respondError(w, err)
			return
		}
		task, err := store.PatchTask(r.Context(), r.PathValue("id"), patch, "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"task": task})
	})
	mux.HandleFunc("POST /api/tasks/{id}/notes", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Note string `json:"note"`
		}
		if err := decodeJSON(r, &input); err != nil {
			respondError(w, err)
			return
		}
		task, err := store.AddNote(r.Context(), r.PathValue("id"), input.Note, "dashboard")
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusCreated, map[string]any{"task": task})
	})
	mux.HandleFunc("GET /api/tasks/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		events, err := store.TaskEvents(r.Context(), r.PathValue("id"))
		if err != nil {
			respondError(w, err)
			return
		}
		respondJSON(w, http.StatusOK, map[string]any{"events": events})
	})
	assets, err := fs.Sub(webFiles, "web")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(assets))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
	return mux
}

func decodeJSON(r *http.Request, destination any) error {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return fmt.Errorf("%w: Content-Type must be application/json", errValidation)
	}
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("%w: %v", errValidation, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: expected one JSON object", errValidation)
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func respondError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, errValidation):
		status = http.StatusBadRequest
	case errors.Is(err, errNotFound):
		status = http.StatusNotFound
	case errors.Is(err, errConflict):
		status = http.StatusConflict
	case errors.Is(err, context.Canceled):
		status = http.StatusRequestTimeout
	}
	if status == http.StatusInternalServerError {
		respondJSON(w, status, map[string]any{"error": "internal error"})
		return
	}
	respondJSON(w, status, map[string]any{"error": err.Error()})
}
