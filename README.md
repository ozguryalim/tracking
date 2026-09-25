# Tracking

English | [Türkçe](README.tr.md)

**Keep AI-assisted plans organized by project, with dated notes for the work behind every task.** Tracking provides a CLI for Codex and Claude Code, a local SQLite database, and a simple browser dashboard. Agents can create and edit plans and tasks directly, and move tasks through `todo`, `doing`, `blocked`, and `done`. There is no separate approval gate or mandatory evidence step.

The CLI and dashboard share the same data. Each plan and task belongs to a project; task changes and notes are recorded in an event history with an actor and UTC timestamp. Commands such as `status --json` and `next --json` give agents structured access to that state.

Tracking stores data in SQLite with WAL mode and indexes for project, status, and time queries. There is no separate search index or database server to manage.

## How it works

| Action | Behavior |
| --- | --- |
| `tracking` or `tracking dashboard` | Reuses a healthy local dashboard if one is running; otherwise starts it in the background. Opens the browser and returns to the terminal. |
| `tracking serve` | Runs the dashboard in the foreground for a server or a manually managed session. |
| `tracking stop` | Stops a background dashboard process started by Tracking. It does not stop a manually started `serve` process. |
| Plan and task commands | Access SQLite directly. The dashboard server does not need to be running. |

The dashboard defaults to `http://127.0.0.1:4157`. To use another local port, run `tracking dashboard --listen 127.0.0.1:PORT`; use the same address with `tracking stop --listen 127.0.0.1:PORT` to stop that instance. The launcher accepts loopback addresses only. Web assets are embedded in the binary, so there is no separate Node or web build. In the dashboard, you can switch projects, edit plans and tasks, and inspect progress and event history.

The local HTTP process serves the editable web UI only; CLI commands work without it. The first `tracking` call starts the process, and later calls reuse it. It does not start when the computer boots. Closing the browser tab does not delete data, and the background process stays up until `tracking stop` is called. On the Mac used for measurement, the idle process used approximately **14–18 MiB of RAM** and **0% CPU at the time of measurement**. Browser memory was not included; results vary by device.

## Installation

Building from source requires [Git](https://git-scm.com/downloads) and [Go](https://go.dev/dl/) **1.27.1 or newer**. Clone the repository first, or skip this step if you already have the source:

```sh
git clone https://github.com/ozguryalim/tracking.git
cd tracking
```

Run the build commands below from the Tracking source directory. Installing the binary in a user-owned directory does not require administrator privileges.

### macOS and Linux

```sh
mkdir -p "$HOME/.local/bin"
go build -o "$HOME/.local/bin/tracking" .
"$HOME/.local/bin/tracking" help
```

To run `tracking` from any directory, add `~/.local/bin` to your `PATH`. Check your current value with `echo "$PATH"` first. If the directory is missing, add the following line **once** to your shell profile (usually `~/.zshrc` for zsh or `~/.bashrc` for bash), then open a new terminal:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### Windows PowerShell

```powershell
$bin = Join-Path $HOME '.local\bin'
New-Item -ItemType Directory -Force -Path $bin | Out-Null
go build -o "$bin\tracking.exe" .
& "$bin\tracking.exe" help
```

To run the command from any directory, add `%USERPROFILE%\.local\bin` to **User environment variables → Path** in Windows; leave the system `Path` unchanged. Open a new PowerShell window and verify with `tracking help`. You can also call the binary by its full path without changing `PATH`.

The same Go source builds for macOS, Linux, and Windows. To target another architecture, set `GOOS` and `GOARCH`; for example, `GOOS=linux GOARCH=arm64 go build -o tracking-linux-arm64 .`.

## First project and everyday use

Run these commands **in the project directory you want to track**:

```sh
tracking init --name "My App"
tracking plan add --title "First release" --goal "Prepare a usable release"
tracking task add --title "Settings screen" --description "Implement and check the screen at runtime"
tracking status
tracking next --json
```

`plan add` prints the plan ID, and `task add` prints the task ID. Unless you provide `--plan PLAN_ID`, `task add` uses the latest plan. Task commands accept either a full ID or a unique ID prefix.

```sh
tracking task start TASK_ID --note "Started work on the screen"
tracking task note TASK_ID --note "Added the layout and state handling"
tracking task done TASK_ID --note "Checked the screen at runtime"
tracking status --json
```

Use `tracking task block TASK_ID --note "Reason"` when work is blocked, and `tracking task reopen TASK_ID --note "Reason"` to return it to `todo`. `tracking task edit` and `tracking plan edit` update existing records. `tracking context` prints a short summary of open work. Run `tracking help` for the full command list.

### How are projects separated?

`tracking init` writes `.tracking/project.json` in the current directory. This small file contains the project ID and name; plans and task notes live in the user-owned SQLite database. When the CLI runs in a subdirectory, it looks upward for the project identity. Run `tracking init` at the root of each separate project. `tracking projects` lists known projects.

To link a project created in the dashboard to a local directory, run `tracking attach PROJECT_ID` in that directory. If the directory already belongs to a different project ID, the command does not silently replace the existing link. Projects in the same local database remain separate by ID.

### Import a plan from one file

Example `plan.json`:

```json
{
  "title": "First release",
  "goal": "Prepare a usable release",
  "tasks": [
    {
      "title": "Implement the settings screen",
      "description": "Build the UI and check it at runtime"
    },
    {
      "title": "Check the release",
      "description": "Run relevant tests and record the result"
    }
  ]
}
```

```sh
tracking plan import --file plan.json
```

Import creates a new plan and its tasks together. It requires a title and **1–500 tasks**, each with a title. If `--file` is omitted, JSON is read from standard input; `--file -` does the same.

## Codex and Claude Code integration

In the project you want to track, install the integration for the agents you use:

```sh
tracking integrate codex
tracking integrate claude
```

Each command installs the Tracking skill and a session-start hook that runs `tracking context`. The skill tells the agent to read plans, update tasks, and note completed work. The hook reminds the agent of project status when a session starts, resumes, or refreshes its context; it does not change tasks by itself.

| Agent | Project skill | Project hook |
| --- | --- | --- |
| Codex | `.agents/skills/tracking/SKILL.md` | `.codex/hooks.json` |
| Claude Code | `.claude/skills/tracking/SKILL.md` | `.claude/settings.json` |

`--scope user` installs the corresponding files in your home directory. A user-scoped hook does nothing in a directory that is not linked to a Tracking project. `--no-hook` installs only the skill. An existing skill with different content is preserved unless you pass `--force`. The `tracking` binary must be on the agent's `PATH` for the hook to run. In Codex, you may need to review and trust a new hook through `/hooks`. See the [integration guide](integrations/README.md), [Codex skill docs](https://learn.chatgpt.com/docs/build-skills), [Codex hook docs](https://learn.chatgpt.com/docs/hooks), [Claude Code skill docs](https://code.claude.com/docs/en/skills), and [Claude Code hook docs](https://code.claude.com/docs/en/hooks).

The skill guides the agent's workflow; it does not guarantee that every response will be written to Tracking automatically. Check the record with `tracking status --json` when work ends.

## Running on a server and access boundaries

The same binary runs on a server:

```sh
tracking serve --listen 127.0.0.1:4157
```

`serve` stays in the foreground, so a service manager can supervise it. The HTTP dashboard and API do not have built-in authentication yet; the default listener therefore binds only to `127.0.0.1`. For remote browser access, use an authenticated reverse proxy or an SSH tunnel. For example, with the server running as above, open `ssh -L 44157:127.0.0.1:4157 user@server` on your computer and visit `http://127.0.0.1:44157`. Do not expose the server port directly to the public internet.

The CLI currently uses a local SQLite file; it does not connect to a remote Tracking server as a client or synchronize automatically. To change server-side projects through the CLI, run commands in the server environment using the same `TRACKING_DB` value. Use the HTTP server for remote browser access rather than opening one SQLite file from two machines over a network share.

## Data and privacy

Tracking does not send data to a cloud account or external service on its own. Plan titles, task descriptions, notes, and timestamped events stay in the SQLite file on your machine. When an agent reads `tracking context` or `status` output, that content also enters the AI tool's session context.

By default, the database is at `tracking/tracking.db` under Go's [user configuration directory](https://pkg.go.dev/os#UserConfigDir):

| System | Default location |
| --- | --- |
| macOS | `~/Library/Application Support/tracking/tracking.db` |
| Linux | `$XDG_CONFIG_HOME/tracking/tracking.db`, or `~/.config/tracking/tracking.db` if unset |
| Windows | `%AppData%\tracking\tracking.db` |

Set `TRACKING_DB` to choose another database file. `TRACKING_ACTOR` sets the actor name on CLI events; otherwise the system user name is used. Start the CLI and dashboard with the same `TRACKING_DB` value if you want them to show the same data. `.tracking/project.json` is only the project link, not a backup of the database.

## FAQ

**Can an agent update tasks while the dashboard is closed?** Yes. Plan and task commands write directly to SQLite. Open the browser dashboard only when you need it.

**Does closing the browser tab delete data?** No. Data remains in SQLite. Use `tracking stop` if you also want to stop the background dashboard process.

**`tracking` is not found.** Verify the binary using its full path, then check that its directory is on your `PATH`. Open a new terminal after changing a shell profile or the Windows user `Path`.

**I see the wrong project or an empty plan.** Run the command from the intended project directory and check `.tracking/project.json`. The CLI and dashboard must use the same `TRACKING_DB`. Compare `tracking projects` with `tracking status --json`.

**The skill or hook is not working.** Check the path printed by `tracking integrate codex` or `tracking integrate claude`. Restart the agent session, make sure it can find `tracking` on its `PATH`, and inspect the new hook's trust status with `/hooks` in Codex.

**Port 4157 is in use.** `tracking` reuses a healthy Tracking dashboard connected to the same database. If another program or a Tracking instance using a different database holds the port, choose a free local port with `tracking dashboard --listen 127.0.0.1:PORT`.

**I rebuilt Tracking, but the dashboard has not changed.** The existing background process may still be running the earlier binary. Run `tracking stop`, then `tracking`.

## Copy-paste setup prompt for an AI agent

Send the following prompt to Codex or Claude Code **with the project you want to track open**. It asks which agents you use, then installs only those integrations:

```text
I want to use Tracking in the current project. First ask whether I use Codex,
Claude Code, or both; install integrations only for the agents I choose.

Check whether the tracking command is available. If it is not, look for the
Tracking source in the open workspace. If it is not there, clone
https://github.com/ozguryalim/tracking.git into a suitable user-owned source
directory. Check for Git and the required Go version; if either is missing,
use a trustworthy installation method appropriate for my operating system.
Build the binary from the source directory and install it in a user-owned
directory without administrator privileges. If a binary already exists,
check its path and identity; do not overwrite an unrelated program.

Add the installation directory to PATH safely: preserve the existing value,
avoid duplicate entries, and back up any shell profile you change. On
Windows, change only the user Path. Verify that tracking help works in a new
terminal.

Return to the project directory where this session started. Preserve any
existing .tracking/project.json link while initializing the project with
tracking init --name. Run tracking integrate for the agents I selected and
verify the skill and hook files. Tell me explicitly that I need to review
and trust a new Codex hook through /hooks.

Finally, run tracking to open the dashboard. Briefly report the installed
binary path, project ID, database location, and the first three commands
I should use. If a step fails, explain why. Leave unrelated configuration
alone.
```
