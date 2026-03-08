# Contributing to Product MCP

Thank you for your interest in contributing to the Product MCP project!

## How to Contribute

1. **Fork the repository** on GitHub.
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/product-mcp.git
   ```
3. **Create a new branch** for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** and ensure they follow the [PRODUCT.md](https://github.com/CrymfoxLabs/product-repos/PRODUCT.md) specification.
5. **Run tests** (once added) to make sure everything still works.
6. **Commit your changes** with clear and descriptive messages.
7. **Push to your fork** and **submit a Pull Request**.

## Guidelines

- **Code Style**: Follow standard Go formatting (`go fmt`).
- **Commits**: Use the [Conventional Commits](https://www.conventionalcommits.org/) format if possible.
- **Documentation**: Update the README if you add or change tool functionality.

## Project Structure

- `main.go`: Entry point and server initialization.
- `tools/`: MCP tool implementations grouped by domain (project, goals, domains, issues).
- `product-go`: The underlying library for parsing and manipulating the specification.
