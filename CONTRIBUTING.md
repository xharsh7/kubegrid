# Contributing to kubegrid

Thank you for your interest in contributing to kubegrid! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Making Changes](#making-changes)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Review Process](#review-process)

## Code of Conduct

This project follows a code of conduct. Be respectful, professional, and inclusive in all interactions.

## Getting Started

1. **Fork the repository**
   ```bash
   # Click the "Fork" button on GitHub
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/kubegrid.git
   cd kubegrid
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/xharsh7/kubegrid.git
   ```

4. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.25 or higher
- kubectl configured with at least one cluster
- Make (optional but recommended)
- A terminal with Unicode support

### Build and Run

```bash
# Using Make
make build
make run

# Or directly with Go
go build -o kubegrid ./cmd/kubegrid
./kubegrid
```

### Development Mode

```bash
# Run with race detector
make dev

# Or
go run -race ./cmd/kubegrid
```

## Project Structure

```
kubegrid/
├── cmd/kubegrid/          # Application entry point
│   └── main.go
├── internal/              # Internal packages
│   ├── cluster/           # Cluster status collection
│   │   ├── collector.go
│   │   └── status.go
│   ├── config/            # Configuration loading
│   │   └── loader.go
│   └── tui/               # Terminal UI
│       ├── app.go         # Main app model
│       ├── cluster_view.go
│       ├── resource_view.go
│       ├── help_view.go
│       ├── layout.go      # Pane layout tree
│       └── utils.go
├── pkg/                   # Public packages
│   └── k8s/
│       └── client.go      # Kubernetes client wrapper
├── ARCHITECTURE.md        # Architecture documentation
├── README.md
└── Makefile
```

### Package Guidelines

- `cmd/`: Application entrypoints
- `internal/`: Private application code
- `pkg/`: Public, reusable packages
- Keep packages focused and single-purpose
- Minimize dependencies between packages

## Making Changes

### Adding a New Feature

1. **Check existing issues** - See if someone is already working on it
2. **Create an issue** - Discuss the feature before implementing
3. **Design first** - Consider the tmux/vim/k9s design principles
4. **Implement** - Follow the coding guidelines
5. **Test** - Write tests for new functionality
6. **Document** - Update README and ARCHITECTURE.md

### Adding a New View

To add a new TUI view:

1. Create a new file in `internal/tui/`, e.g., `my_view.go`

2. Implement the `tea.Model` interface:
```go
package tui

import tea "github.com/charmbracelet/bubbletea"

type myViewModel struct {
    // Your state here
}

func NewMyView() myViewModel {
    return myViewModel{}
}

func (m myViewModel) Init() tea.Cmd {
    return nil
}

func (m myViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keys
    }
    return m, nil
}

func (m myViewModel) View() string {
    // Render your view
    return "Your view here"
}
```

3. Register in the app model or add keybinding to open it

### Adding Kubernetes Resources

To add support for a new K8s resource type:

1. Add the API call to `pkg/k8s/client.go`:
```go
func (c *Client) ListMyResource(ctx context.Context) ([]MyResource, error) {
    // Implementation
}
```

2. Add to `resourceType` enum in `internal/tui/resource_view.go`
3. Add rendering method `renderMyResources()`
4. Add keybinding to switch to it

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting: `make fmt`
- Run `go vet`: `make lint`
- Keep functions small and focused
- Prefer clarity over cleverness

### Naming Conventions

- Use descriptive names: `clusterStatus` not `cs`
- Interfaces: `Reader`, `Writer` (noun or verb+er)
- Packages: lowercase, single word when possible
- Files: snake_case

### Error Handling

```go
// Good
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Bad - loses context
if err != nil {
    return err
}
```

### Logging

- Use fmt for user-facing messages
- Minimal logging in the TUI
- Error messages should be actionable

### TUI Guidelines

- **Keyboard-first**: All functionality accessible via keyboard
- **Responsive**: Handle window resize gracefully
- **Consistent**: Use familiar keybindings (vim/tmux style)
- **Visual feedback**: Show what state the UI is in
- **Error handling**: Display errors clearly to users

### Performance

- Use goroutines for parallel operations
- Avoid blocking the UI thread
- Cache expensive operations when appropriate
- Profile before optimizing

## Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/tui/...
```

### Writing Tests

```go
func TestClusterStatusCollection(t *testing.T) {
    // Arrange
    contexts := []config.KubeContext{
        {Name: "test", Source: "/path/to/config"},
    }

    // Act
    statuses := cluster.CollectStatuses(contexts)

    // Assert
    if len(statuses) != 1 {
        t.Errorf("expected 1 status, got %d", len(statuses))
    }
}
```

### Test Coverage

- Aim for >80% coverage for business logic
- Don't test framework code (tea.Model implementations)
- Test error conditions
- Use table-driven tests for multiple scenarios

## Submitting Changes

### Commit Messages

Follow conventional commits:

```
feat: add pod deletion functionality
fix: correct namespace switching bug
docs: update installation instructions
refactor: simplify layout rendering
test: add tests for cluster collector
```

### Pull Request Process

1. **Update your branch**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests**
   ```bash
   make test
   make lint
   ```

3. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create Pull Request**
   - Use a clear title
   - Describe what and why
   - Reference related issues
   - Include screenshots for UI changes
   - List breaking changes if any

### PR Template

```markdown
## Description
Brief description of changes

## Motivation
Why is this change needed?

## Changes
- Change 1
- Change 2

## Testing
How was this tested?

## Screenshots (if applicable)
[Add screenshots here]

## Checklist
- [ ] Tests pass
- [ ] Code formatted
- [ ] Documentation updated
- [ ] ARCHITECTURE.md updated (if needed)
```

## Review Process

1. **Automated checks** - CI must pass
2. **Code review** - At least one approval required
3. **Discussion** - Address reviewer feedback
4. **Approval** - Maintainer approves
5. **Merge** - Squash and merge to main

### Review Guidelines

When reviewing PRs:
- Be constructive and kind
- Focus on code, not the person
- Explain your reasoning
- Suggest alternatives
- Approve when satisfied

## Development Tips

### Hot Reload

For rapid development:
```bash
# In one terminal
make build && ./kubegrid

# Make changes, then Ctrl+C and re-run
```

### Debugging

```go
// Add debug output (remove before committing)
fmt.Fprintf(os.Stderr, "DEBUG: value=%v\n", value)
```

### Testing with Mock Clusters

Create test kubeconfig:
```yaml
# ~/.kube/test.config
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test
current-context: test
users:
- name: test-user
  user: {}
```

### Terminal Testing

Test in different terminals:
- iTerm2 (macOS)
- Terminal.app (macOS)
- gnome-terminal (Linux)
- Windows Terminal
- tmux inside terminal

## Questions?

- **Issues**: [GitHub Issues](https://github.com/xharsh7/kubegrid/issues)
- **Discussions**: [GitHub Discussions](https://github.com/xharsh7/kubegrid/discussions)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to kubegrid! 🎉
