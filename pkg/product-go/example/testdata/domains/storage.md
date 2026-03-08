---
schema: domain/1.0
name: storage
product: taskr
---

## Storage
Manages reading and writing of the task JSON store to disk.
**files**: `internal/store/file.go`, `internal/store/model.go`

### load-store
- **state**: ready
- **why**: Every command that reads tasks needs the store loaded first.
- **acceptance**:
  - [ ] when store file does not exist, then return empty task list
  - [ ] when store file is corrupt, then return a descriptive error
- **files**: `internal/store/file.go:10`
- **notes**: Uses atomic rename-on-write to prevent corruption.

### save-store
- **state**: ready
- **why**: Changes must be persisted atomically to prevent data loss.
- **acceptance**:
  - [ ] when save is called, then file is written with 0600 permissions
  - [ ] when write fails, then original file is preserved
- **depends-on**: load-store
- **files**: `internal/store/file.go:45`
- **issues**: [ISSUE-003]
