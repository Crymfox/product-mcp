# Product MCP Server

An MCP (Model Context Protocol) server for managing projects that follow the [PRODUCT.md](https://github.com/kidkuddy/PRODUCT.md) specification.

This repository is a monorepo that provides a bridge between AI models and structured project documentation.

## Acknowledgements

This project is built upon the foundational work of **[kidkuddy](https://github.com/kidkuddy)**:
- **[PRODUCT.md](https://github.com/kidkuddy/PRODUCT.md)**: The original specification for machine-readable project documentation.
- **[product-go](https://github.com/kidkuddy/product-go)**: The reference Go implementation for parsing and querying the spec (included here in `pkg/product-go`).

## Monorepo Structure

- **`product-mcp`**: The MCP server implementation.
- **`pkg/product-go`**: Forked and integrated reference Go library for parsing and manipulating `PRODUCT.md` files.

This server provides a set of tools to programmatically read, update, and validate your project documentation, ensuring adherence to the spec while enabling AI agents to act as effective project managers.

## Features

- **Integrated Library**: Includes the `product-go` engine for robust Markdown parsing.
- **Project Management**: Initialize new projects, update vision, manage tech stack, and define scopes.
- **Zero-Config Discovery**: Automatically searches upwards from the current directory to find the project root.
- **Goal Tracking**: Add, list, and toggle completion status of project goals.
- **Domain & Feature Management**: Manage domains and features, supporting both inline and split-file (`domains/*.md`) architectures automatically.
- **Issue Tracking**: Create, update, and list issues in `ISSUES.md` with auto-incrementing IDs.
- **Smart Save**: Intelligently updates the correct files (main `PRODUCT.md` vs. individual domain files) without data loss or duplication.

## Installation

```bash
cd /home/crymfox/repos/CrymfoxLabs/product-repos/product-mcp
go build -o product-mcp
```

## Agent Configuration

To get the best results, add the following instructions to your coding agent (e.g., `.cursorrules`, `.agent`, or system prompt).

### Project Management Protocol

As an agent working on this project, you MUST adhere to the following workflow to ensure `PRODUCT.md` and `ISSUES.md` remain the "Single Source of Truth."

#### 1. Discovery & Context
- **The "First Look" Rule**: BEFORE writing any code or proposing a plan, run `project_summary`.
- **Domain Audit**: Read `list_domains` and subsequent domain details to understand the "Why" and "Acceptance Criteria" for existing features.
- **Backlog Check**: Run `list_issues status:open` to see if your task is already tracked.

#### 2. Operational Discipline (The "Sync" Rule)
- **Start of Task**: 
    - Transition the relevant issue to `status: in-progress` using `update_issue`.
    - Set associated features to `state: building` using `update_feature_state`.
- **Implementation**:
    - If you add a library, run `manage_tech_stack`.
    - If you add a new sub-folder or module, run `manage_scopes`.
- **Completion**:
    - Set issue to `status: closed` and provide a concise `fix` description.
    - Set features to `state: ready`.
    - If a Goal is fully satisfied, use `toggle_goal done:true`.

#### 3. Documentation Standards
- **Domains**: ALWAYS use `as_file: true` when adding domains to keep `PRODUCT.md` readable.
- **Traceability**: Every Feature MUST belong to a Domain. Every Issue SHOULD relate to a Feature.

---

## Tools

### Project
- `project_init`: Initialize a new project structure.
- `project_summary`: Get high-level stats and progress.
- `update_vision`: Update the central vision statement.
- `manage_tech_stack`: Add/update layers and technologies.
- `manage_scopes`: Manage project directories and their implementation state.

### Goals
- `list_goals`: List all goals and their status.
- `add_goal`: Add a new project goal.
- `toggle_goal`: Mark a goal as completed/pending.

### Domains
- `list_domains`: List all domains and their locations.
- `add_domain`: Add a new domain (`as_file: true` recommended).
- `add_feature`: Add a feature with state, why, and acceptance criteria.
- `update_feature_state`: Transition a feature (vision -> building -> ready).
- `set_feature_acceptance`: Overwrite the acceptance criteria list for a feature.
- `toggle_feature_acceptance`: Mark a specific acceptance criterion as done or pending.

### Issues
- `list_issues`: Filter and list issues from `ISSUES.md`.
- `create_issue`: Create a new issue (Auto-generates `ISSUE-NNN`).
- `update_issue`: Update status or record a fix.

## Project Structure

- `main.go`: Entry point for the MCP server.
- `tools/`: MCP tool definitions and handlers.
- `pkg/product-go/`: Reference Go library for the PRODUCT.md specification.

## License

MIT - See [LICENSE](LICENSE) for details. Contains work originally by [kidkuddy](https://github.com/kidkuddy).

---

<p align="center">Built with ❤️ by Crymfox Labs</p>
