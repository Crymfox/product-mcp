---
schema: issues/1.0
project: taskr
---

## [ISSUE-001] Add command panics on empty title

- **type**: bug
- **severity**: high
- **status**: open
- **source**: manual
- **effort**: S
- **domain**: cli
- **feature**: add-task

The `add` command panics with a nil pointer dereference when the user provides an empty string as the task title.

**fix**: Add a validation check before accessing the title field; return a user-friendly error.

---
## [ISSUE-002] Support due-date flag on add command

- **type**: feature
- **severity**: medium
- **status**: open
- **source**: manual
- **effort**: M
- **domain**: cli
- **feature**: add-task

Users cannot currently specify a due date when adding a task. The `--due` flag is parsed but ignored.

**fix**: Wire the due-date flag value through to the store layer and persist it alongside the task.

---
## [ISSUE-003] Task file world-readable permissions

- **type**: security
- **severity**: high
- **status**: closed
- **source**: audit
- **effort**: S
- **location**: ``internal/store/file.go:34``
- **audit-ref**: SEC-001
- **domain**: storage

The task JSON file is created with 0644 permissions, making it readable by all local users.

**fix**: Change os.WriteFile call to use 0600 permissions.

---
## [ISSUE-004] Add shell completion support

- **type**: task
- **severity**: low
- **status**: open
- **source**: manual
- **effort**: L

Cobra supports generating shell completion scripts but we have not wired this up. Users must type full command names.

**fix**: Call `rootCmd.GenBashCompletionFile` and `rootCmd.GenZshCompletionFile` and document in README.

---
