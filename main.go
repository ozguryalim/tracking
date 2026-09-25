package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed integrations/tracking/SKILL.md
var trackingSkill []byte

type projectManifest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "tracking:", err)
		os.Exit(1)
	}
}

func run(args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return runDashboard(nil, out, errOut)
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		printUsage(out)
		return nil
	}
	if args[0] == "integrate" {
		return runIntegrate(args[1:], out, errOut)
	}
	if args[0] == "dashboard" {
		return runDashboard(args[1:], out, errOut)
	}
	if args[0] == "stop" {
		return runStop(args[1:], out, errOut)
	}
	path, err := databasePath()
	if err != nil {
		return err
	}
	store, err := openStore(path)
	if err != nil {
		return err
	}
	defer store.Close()
	ctx := context.Background()
	switch args[0] {
	case "init":
		return runInit(ctx, store, args[1:], out, errOut)
	case "attach":
		return runAttach(ctx, store, args[1:], out)
	case "projects":
		projects, err := store.ListProjects(ctx)
		if err != nil {
			return err
		}
		return writeJSON(out, map[string]any{"projects": projects})
	case "plan":
		return runPlan(ctx, store, args[1:], out, errOut)
	case "task":
		return runTask(ctx, store, args[1:], out, errOut)
	case "status":
		return runStatus(ctx, store, args[1:], out, errOut)
	case "next":
		return runNext(ctx, store, args[1:], out, errOut)
	case "context":
		return runContext(ctx, store, args[1:], out, errOut)
	case "serve":
		return runServer(store, args[0], args[1:], out, errOut)
	default:
		return fmt.Errorf("unknown command %q (try tracking help)", args[0])
	}
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, `Tracking — project plans and progress for AI agents

Usage:
  tracking
  tracking init --name NAME
  tracking attach PROJECT_ID
  tracking projects
  tracking plan add --title TITLE [--goal GOAL]
  tracking plan import --file plan.json
  tracking plan edit ID --title TITLE [--goal GOAL]
  tracking task add --title TITLE [--description TEXT] [--plan PLAN_ID]
  tracking task edit ID [--title TITLE] [--description TEXT] [--plan PLAN_ID]
  tracking task start ID [--note TEXT]
  tracking task done ID --note TEXT
  tracking task block ID --note TEXT
  tracking task reopen ID [--note TEXT]
  tracking task note ID --note TEXT
  tracking status [--json]
  tracking next [--json]
  tracking context
  tracking dashboard [--listen 127.0.0.1:4157]
  tracking stop [--listen 127.0.0.1:4157]
  tracking serve [--listen 127.0.0.1:4157]
  tracking integrate codex|claude [--scope project|user] [--no-hook]

Set TRACKING_DB to choose the database file and TRACKING_ACTOR to label changes.`)
}

func newFlags(name string, errOut io.Writer) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(errOut)
	return flags
}

func databasePath() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("TRACKING_DB")); configured != "" {
		return filepath.Abs(configured)
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "tracking", "tracking.db"), nil
}

func writeJSON(out io.Writer, value any) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func manifestFile(dir string) string {
	return filepath.Join(dir, ".tracking", "project.json")
}

func readManifest(dir string) (projectManifest, string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return projectManifest{}, "", err
	}
	for {
		data, err := os.ReadFile(manifestFile(dir))
		if err == nil {
			var manifest projectManifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				return projectManifest{}, "", fmt.Errorf("read project manifest: %w", err)
			}
			if manifest.ID == "" || manifest.Name == "" {
				return projectManifest{}, "", fmt.Errorf("invalid project manifest at %s", dir)
			}
			return manifest, dir, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return projectManifest{}, "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return projectManifest{}, "", errNotFound
		}
		dir = parent
	}
}

func currentProject(ctx context.Context, store *Store) (projectManifest, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return projectManifest{}, err
	}
	manifest, root, err := readManifest(cwd)
	if errors.Is(err, errNotFound) {
		return projectManifest{}, fmt.Errorf("%w: no Tracking project here; run tracking init --name NAME", errNotFound)
	}
	if err != nil {
		return projectManifest{}, err
	}
	if err := store.RegisterProject(ctx, Project{ID: manifest.ID, Name: manifest.Name, Path: root}, actorName()); err != nil {
		return projectManifest{}, err
	}
	return manifest, nil
}

func runInit(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	flags := newFlags("init", errOut)
	name := flags.String("name", "", "project name")
	if err := flags.Parse(args); err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := manifestFile(cwd)
	if data, err := os.ReadFile(path); err == nil {
		var manifest projectManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return err
		}
		if manifest.ID == "" || manifest.Name == "" {
			return fmt.Errorf("invalid existing manifest at %s", path)
		}
		if err := store.RegisterProject(ctx, Project{ID: manifest.ID, Name: manifest.Name, Path: cwd}, actorName()); err != nil {
			return err
		}
		fmt.Fprintf(out, "Already initialized: %s (%s)\n", manifest.Name, manifest.ID)
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if strings.TrimSpace(*name) == "" {
		return fmt.Errorf("--name is required")
	}
	id, err := newID("pr")
	if err != nil {
		return err
	}
	manifest := projectManifest{ID: id, Name: strings.TrimSpace(*name)}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return err
	}
	if err := store.RegisterProject(ctx, Project{ID: manifest.ID, Name: manifest.Name, Path: cwd}, actorName()); err != nil {
		return err
	}
	fmt.Fprintf(out, "Initialized %s (%s)\n", manifest.Name, manifest.ID)
	return nil
}

func runAttach(ctx context.Context, store *Store, args []string, out io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("attach requires one project ID")
	}
	project, err := store.getProject(ctx, args[0])
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if existing, _, err := readManifest(cwd); err == nil {
		if existing.ID == project.ID {
			fmt.Fprintf(out, "Already attached: %s (%s)\n", project.Name, project.ID)
			return nil
		}
		return fmt.Errorf("this directory already belongs to %s", existing.Name)
	} else if !errors.Is(err, errNotFound) {
		return err
	}
	manifest := projectManifest{ID: project.ID, Name: project.Name}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	path := manifestFile(cwd)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		return err
	}
	if err := store.RegisterProject(ctx, Project{ID: project.ID, Name: project.Name, Path: cwd}, actorName()); err != nil {
		return err
	}
	fmt.Fprintf(out, "Attached %s (%s)\n", project.Name, project.ID)
	return nil
}

func runPlan(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("plan subcommand required: add or edit")
	}
	switch args[0] {
	case "add":
		flags := newFlags("plan add", errOut)
		title := flags.String("title", "", "plan title")
		goal := flags.String("goal", "", "plan goal")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		manifest, err := currentProject(ctx, store)
		if err != nil {
			return err
		}
		plan, err := store.CreatePlan(ctx, manifest.ID, *title, *goal, actorName())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Created plan %s: %s\n", plan.ID, plan.Title)
		return nil
	case "import":
		flags := newFlags("plan import", errOut)
		file := flags.String("file", "-", "JSON file, or - for stdin")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		var data []byte
		var err error
		if *file == "-" {
			data, err = io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
		} else {
			data, err = os.ReadFile(*file)
		}
		if err != nil {
			return err
		}
		var input PlanImport
		if err := json.Unmarshal(data, &input); err != nil {
			return fmt.Errorf("read plan JSON: %w", err)
		}
		manifest, err := currentProject(ctx, store)
		if err != nil {
			return err
		}
		plan, tasks, err := store.ImportPlan(ctx, manifest.ID, input, actorName())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Imported plan %s with %d tasks: %s\n", plan.ID, len(tasks), plan.Title)
		return nil
	case "edit":
		if len(args) < 2 {
			return fmt.Errorf("plan edit requires an ID")
		}
		manifest, err := currentProject(ctx, store)
		if err != nil {
			return err
		}
		previous, err := store.GetPlan(ctx, args[1])
		if err != nil {
			return err
		}
		if previous.ProjectID != manifest.ID {
			return fmt.Errorf("%w: plan belongs to another project", errValidation)
		}
		flags := newFlags("plan edit", errOut)
		title := flags.String("title", previous.Title, "plan title")
		goal := flags.String("goal", previous.Goal, "plan goal")
		if err := flags.Parse(args[2:]); err != nil {
			return err
		}
		plan, err := store.UpdatePlan(ctx, previous.ID, *title, *goal, actorName())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Updated plan %s\n", plan.ID)
		return nil
	default:
		return fmt.Errorf("unknown plan subcommand %q", args[0])
	}
}

func runTask(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("task subcommand required")
	}
	if args[0] == "add" {
		flags := newFlags("task add", errOut)
		title := flags.String("title", "", "task title")
		description := flags.String("description", "", "task description")
		planID := flags.String("plan", "", "plan ID (defaults to latest plan)")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		manifest, err := currentProject(ctx, store)
		if err != nil {
			return err
		}
		task, err := store.CreateTask(ctx, manifest.ID, *planID, *title, *description, actorName())
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Created task %s: %s\n", task.ID, task.Title)
		return nil
	}
	if len(args) < 2 {
		return fmt.Errorf("task %s requires an ID", args[0])
	}
	id, err := store.ResolveTaskID(ctx, args[1])
	if err != nil {
		return err
	}
	manifest, err := currentProject(ctx, store)
	if err != nil {
		return err
	}
	currentTask, err := store.GetTask(ctx, id)
	if err != nil {
		return err
	}
	if currentTask.ProjectID != manifest.ID {
		return fmt.Errorf("%w: task belongs to another project", errValidation)
	}
	flags := newFlags("task "+args[0], errOut)
	note := flags.String("note", "", "work note")
	title := flags.String("title", "", "task title")
	description := flags.String("description", "", "task description")
	planID := flags.String("plan", "", "plan ID")
	if err := flags.Parse(args[2:]); err != nil {
		return err
	}
	var task Task
	switch args[0] {
	case "start", "done", "block", "reopen":
		status := map[string]string{"start": "doing", "done": "done", "block": "blocked", "reopen": "todo"}[args[0]]
		task, err = store.PatchTask(ctx, id, TaskPatch{Status: &status, Note: *note}, actorName())
	case "note":
		task, err = store.AddNote(ctx, id, *note, actorName())
	case "edit":
		patch := TaskPatch{Note: *note}
		flags.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "title":
				patch.Title = title
			case "description":
				patch.Description = description
			case "plan":
				patch.PlanID = planID
			}
		})
		task, err = store.PatchTask(ctx, id, patch, actorName())
	default:
		return fmt.Errorf("unknown task subcommand %q", args[0])
	}
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s %s: %s\n", task.ID, task.Status, task.Title)
	return nil
}

func runStatus(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	flags := newFlags("status", errOut)
	asJSON := flags.Bool("json", false, "output JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	manifest, err := currentProject(ctx, store)
	if err != nil {
		if !errors.Is(err, errNotFound) {
			return err
		}
		projects, listErr := store.ListProjects(ctx)
		if listErr != nil {
			return listErr
		}
		if *asJSON {
			return writeJSON(out, map[string]any{"projects": projects})
		}
		if len(projects) == 0 {
			fmt.Fprintln(out, "No projects yet. Run tracking init --name NAME in a project directory.")
			return nil
		}
		for _, project := range projects {
			fmt.Fprintf(out, "%s  %s  %d/%d done\n", project.ID, project.Name, project.DoneCount, project.TaskCount)
		}
		return nil
	}
	detail, err := store.ProjectDetail(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(out, detail)
	}
	fmt.Fprintf(out, "%s  %d/%d done\n", detail.Project.Name, detail.Project.DoneCount, detail.Project.TaskCount)
	for _, plan := range detail.Plans {
		fmt.Fprintf(out, "\n%s  %s\n", plan.ID, plan.Title)
		for _, task := range detail.Tasks {
			if task.PlanID == plan.ID {
				fmt.Fprintf(out, "  %-7s %s  %s\n", task.Status, task.ID, task.Title)
			}
		}
	}
	return nil
}

func runNext(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	flags := newFlags("next", errOut)
	asJSON := flags.Bool("json", false, "output JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	manifest, err := currentProject(ctx, store)
	if err != nil {
		return err
	}
	tasks, err := store.NextTasks(ctx, manifest.ID)
	if err != nil {
		return err
	}
	if *asJSON {
		return writeJSON(out, map[string]any{"tasks": tasks})
	}
	if len(tasks) == 0 {
		fmt.Fprintln(out, "No ready tasks.")
	}
	for _, task := range tasks {
		fmt.Fprintf(out, "%s  %s\n", task.ID, task.Title)
	}
	return nil
}

func runContext(ctx context.Context, store *Store, args []string, out, errOut io.Writer) error {
	flags := newFlags("context", errOut)
	if err := flags.Parse(args); err != nil {
		return err
	}
	manifest, err := currentProject(ctx, store)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil
		}
		return err
	}
	detail, err := store.ProjectDetail(ctx, manifest.ID)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Tracking project: %s. %d of %d tasks done.\n", detail.Project.Name,
		detail.Project.DoneCount, detail.Project.TaskCount)
	for _, plan := range detail.Plans {
		fmt.Fprintf(out, "Plan %s: %s\n", plan.ID, plan.Title)
		for _, task := range detail.Tasks {
			if task.PlanID == plan.ID && task.Status != "done" {
				fmt.Fprintf(out, "- [%s] %s %s\n", task.Status, task.ID, task.Title)
			}
		}
	}
	return nil
}

func runIntegrate(args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("integrate requires codex or claude")
	}
	if args[0] != "codex" && args[0] != "claude" {
		return fmt.Errorf("unknown integration %q", args[0])
	}
	flags := newFlags("integrate", errOut)
	scope := flags.String("scope", "project", "project or user")
	force := flags.Bool("force", false, "replace an existing skill")
	noHook := flags.Bool("no-hook", false, "install only the skill")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	var base string
	switch *scope {
	case "project":
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		base = cwd
	case "user":
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		base = home
	default:
		return fmt.Errorf("--scope must be project or user")
	}
	var path string
	if args[0] == "codex" {
		path = filepath.Join(base, ".agents", "skills", "tracking", "SKILL.md")
	} else {
		path = filepath.Join(base, ".claude", "skills", "tracking", "SKILL.md")
	}
	if current, err := os.ReadFile(path); err == nil {
		if string(current) == string(trackingSkill) {
			fmt.Fprintf(out, "Tracking skill already installed: %s\n", path)
		} else if !*force {
			return fmt.Errorf("skill already exists at %s; use --force to replace it", path)
		} else {
			if err := os.WriteFile(path, trackingSkill, 0644); err != nil {
				return err
			}
			fmt.Fprintf(out, "Updated Tracking skill: %s\n", path)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	} else {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, trackingSkill, 0644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Installed Tracking skill: %s\n", path)
	}
	if !*noHook {
		hookPath, err := installTrackingHook(args[0], base)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Installed session hook: %s\n", hookPath)
		if args[0] == "codex" {
			fmt.Fprintln(out, "In Codex, review and trust the new hook with /hooks.")
		}
	}
	return nil
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", url)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
