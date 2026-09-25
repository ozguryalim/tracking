---
name: tracking
description: Track a multi-step project plan with the Tracking CLI when work spans multiple tasks or agent sessions.
---

# Tracking

Use Tracking as the shared project plan and progress record. Keep the plan and task states current while working; the agent may edit tasks and mark them done directly.

1. Run `tracking status --json` before continuing tracked work. If the project is not initialized, use `tracking init --name "Project Name"`.
2. Record a new plan with `tracking plan add --title "Plan title" --goal "Outcome"`. Add actionable tasks with `tracking task add --title "Task title" --description "Expected work" --plan PLAN_ID`. For many tasks, import one JSON document with `tracking plan import --file plan.json`.
3. Resume from `tracking next --json`. Start a task with `tracking task start ID` and record meaningful progress with `tracking task note ID --note "What changed"`.
4. When the work is complete, run `tracking task done ID --note "What was completed"`. If work cannot proceed, run `tracking task block ID --note "What is needed"`.
5. Check `tracking status --json` before reporting overall completion. `tracking dashboard` opens the visual project view when useful.

Tracking records timestamps for task and note changes automatically. Keep task descriptions and completion notes specific enough for the next agent session to continue accurately. Do not require separate approval or evidence to update a task.
