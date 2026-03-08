---
schema: domain/1.0
name: memory
product: taskr
---

# Memory

Durable storage for tasks: SQL access and file read/write.
**files**: `tools/memory/main.go`, `tools/memory/handlers.go`

## sql-query

- **state**: ready
- **why**: Tasks need structured storage for queries.
- **acceptance**:
  - [ ] when query is called, the named SQLite file is opened
  - [ ] results are returned as JSON array
- **files**: `tools/memory/handlers.go:10`
- **notes**: No read-only enforcement yet.

## file-read

- **state**: ready
- **why**: Need to read task files from disk.
- **acceptance**:
  - [ ] when file_read is called, the file content is returned
  - [ ] when file does not exist, an error is returned
- **depends-on**: sql-query
- **files**: `tools/memory/handlers.go:50`
- **issues**: [ISSUE-002]
