---
schema: product/1.0
name: taskr
description: A minimal CLI task manager built in Go.
version: 0.3.0
last_updated: "2025-01-15"
---

## Vision

A fast, keyboard-driven task manager that lives entirely in the terminal. No cloud, no subscriptions — just files.

## Goals

- [ ] add-task: Users can add tasks with a title and optional due date
- [ ] list-tasks: Users can list tasks filtered by status or tag
- [x] persistence: Tasks persist to disk across sessions

## Tech Stack

| Layer | Technology |
|-------|------------|
| CLI | cobra |
| Storage | JSON flat file |

## Architecture

The CLI layer delegates to a core engine. The engine reads and writes a JSON store.

```
┌─────────┐     ┌────────┐     ┌──────────┐
│  cobra  │────▶│ engine │────▶│ JSON     │
│  (CLI)  │     │        │     │ store    │
└─────────┘     └────────┘     └──────────┘
```

## Scopes

| name | path | type | state |
|------|------|------|-------|
| cli | cmd/ | package | active |
| storage | internal/store/ | package | active |

## Domains

### CLI
Handles all user-facing commands and flag parsing.
**files**: `cmd/root.go`, `cmd/add.go`, `cmd/list.go`

#### add-task
- **state**: building
- **why**: Core user action — without this nothing else matters.
- **acceptance**:
  - [ ] when user runs `taskr add "buy milk"`, then a new task appears in list
  - [ ] when due date flag is provided, then task stores the date
- **files**: `cmd/add.go:15`
- **issues**: [ISSUE-001], [ISSUE-002]

#### list-tasks
- **state**: specced
- **why**: Users need to see their tasks to act on them.
- **acceptance**:
  - [ ] when user runs `taskr list`, then all open tasks are shown
  - [ ] when `--tag` flag is supplied, then only matching tasks are shown
- **depends-on**: add-task
- **files**: `cmd/list.go:10`

### memory

See domains/memory.md.

## Open Questions

1. Should we support multiple task files (one per project)?
2. What is the migration path if the schema changes?

## References

- domains/storage.md
