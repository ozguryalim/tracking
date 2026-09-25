# Tracking agent integrations

English | [Türkçe](README.tr.md)

[`tracking/SKILL.md`](tracking/SKILL.md) is the canonical Tracking CLI skill in this repository. From a project directory, install the integration for Codex, Claude Code, or both:

```sh
tracking integrate codex
tracking integrate claude
```

Each command installs a separate copy of the skill and adds `tracking context` to a `SessionStart` hook. When a session starts, resumes, or refreshes its context, the agent receives a short summary of open work. The hook only reads project status; it does not mark tasks complete. The agent updates tasks through the CLI commands described in the skill.

| Agent | Project skill | Project hook | User skill | User hook |
| --- | --- | --- | --- | --- |
| Codex | `.agents/skills/tracking/SKILL.md` | `.codex/hooks.json` | `~/.agents/skills/tracking/SKILL.md` | `~/.codex/hooks.json` |
| Claude Code | `.claude/skills/tracking/SKILL.md` | `.claude/settings.json` | `~/.claude/skills/tracking/SKILL.md` | `~/.claude/settings.json` |

The default is `--scope project`, which installs into the current directory. Use `--scope user` to install in your home directory; a user-scoped hook does nothing in directories that are not linked to a Tracking project. Add `--no-hook` to install only the skill. A skill with different existing content is preserved unless you explicitly pass `--force`. Tracking reads existing hook settings and adds its entry without replacing unrelated hooks. Codex may require you to review and trust the new hook through `/hooks` before it runs.

The hook runs `tracking context`, so the `tracking` binary must be on the agent session's `PATH`. Installing the skill does not install the binary. Decide whether to commit project-scoped skill and hook files according to your team's repository policy; user-scoped files apply only on this machine. When the canonical skill changes, update an installed copy with `tracking integrate codex --force` or `tracking integrate claude --force`.

For location and hook behavior, see the [Codex skill](https://learn.chatgpt.com/docs/build-skills), [Codex hook](https://learn.chatgpt.com/docs/hooks), [Claude Code skill](https://code.claude.com/docs/en/skills), and [Claude Code hook](https://code.claude.com/docs/en/hooks) documentation. For installation and everyday use, return to the [main README](../README.md).
