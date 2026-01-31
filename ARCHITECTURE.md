# KUBEGRID ARCHITECTURE

## Design Philosophy

kubegrid follows three core design principles from battle-tested terminal tools:

### 1. Layout Tree (tmux-style)
- UI is a recursive tree of panes
- Each pane can be split vertically or horizontally
- Splits are size-aware and automatically adjust
- Rendering is clipped to terminal bounds

### 2. Focused Pane Model
- Only one pane is active at a time
- Keyboard input routes only to the active pane
- Tab cycles focus like tmux
- Makes the UI predictable and scalable

### 3. Screen vs Pane Separation (k9s/vim-style)
- **Panes**: Embedded views inside layouts
- **Screens**: Fullscreen modal views (help, logs, etc.)
- Esc always backs out safely
- Avoids UI chaos

## Project Structure

```
kubegrid/
├── cmd/
│   └── kubegrid/
│       └── main.go           # Entry point
├── internal/
│   ├── cluster/
│   │   ├── collector.go      # Parallel cluster status collection
│   │   └── status.go         # Individual cluster health checks
│   ├── config/
│   │   └── loader.go         # Kubeconfig discovery and loading
│   └── tui/
│       ├── app.go            # Main application model with layout tree
│       ├── cluster_view.go   # Cluster list and inspection views
│       ├── resource_view.go  # Kubernetes resource viewing
│       ├── help_view.go      # Interactive help screen
│       ├── layout.go         # Recursive pane layout rendering
│       └── utils.go          # Helper functions
└── pkg/
    └── k8s/
        └── client.go         # Kubernetes API client wrapper

```

## Key Components

### Layout Tree (`internal/tui/layout.go`)

The layout system uses a binary tree structure:

```go
type layoutNode struct {
    kind   nodeType  // nodePane, nodeSplitV, or nodeSplitH
    pane   *pane     // Actual content (if leaf node)
    first  *layoutNode  // First child (if split)
    second *layoutNode  // Second child (if split)
}
```

**Splitting algorithm:**
1. Take current pane node
2. Transform it into a split node
3. Keep original as first child
4. Create duplicate as second child
5. Recursively render with size division

### App Model (`internal/tui/app.go`)

Manages global state:
- Root layout node
- Active pane reference
- Command mode state
- Window dimensions

**Key features:**
- Vim-style command mode (`:`)
- Focus cycling (Tab/Shift+Tab)
- Split management (Ctrl+V, Ctrl+H)
- Message display

### Cluster View (`internal/tui/cluster_view.go`)

Shows multi-cluster overview:
- Parallel health checks
- Reachability status
- API latency measurement
- Error display

### Resource View (`internal/tui/resource_view.go`)

Interactive Kubernetes resource browser:
- Pods, Deployments, Services, Namespaces
- Real-time filtering with `/`
- Log viewing with `l`
- Namespace switching with `n`
- Resource deletion with `d`

### K8s Client (`pkg/k8s/client.go`)

Wrapper around client-go with:
- Simplified API
- Namespace management
- Log streaming
- Resource operations

## Data Flow

```
main.go
  ↓
Discover kubeconfigs → Load contexts
  ↓
Initialize cluster status collector
  ↓
Create cluster view model
  ↓
Wrap in app model with layout tree
  ↓
Start bubbletea program
  ↓
User input → Update → View cycle
```

## Message Types

kubegrid uses typed messages for async operations:

- `tea.WindowSizeMsg` - Terminal resize
- `resourceLoadedMsg` - K8s resources fetched
- `logsLoadedMsg` - Pod logs retrieved
- `podDeletedMsg` - Pod deletion result

## Rendering Pipeline

```
1. app.View() called
   ↓
2. Calculate content height (screen - status line)
   ↓
3. root.render(width, height, activePane)
   ↓
4. Recursive tree traversal:
   - Pane nodes: render content with borders
   - Split nodes: divide space & join children
   ↓
5. Apply active/inactive border styles
   ↓
6. Append status line and command bar
```

## Extension Points

### Adding New Views

1. Create view model implementing `tea.Model`:
```go
type myView struct {
    // state
}

func (m myView) Init() tea.Cmd { ... }
func (m myView) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
func (m myView) View() string { ... }
```

2. Register in command handler or keybinding

### Adding New Commands

In `app.go` `executeCommand()`:
```go
case "mycmd":
    // handle command
    a.message = "Command executed"
```

### Adding New Resource Types

1. Add to `resourceType` enum in `resource_view.go`
2. Add corresponding K8s API call in `pkg/k8s/client.go`
3. Add rendering method
4. Add keybinding

## Performance Considerations

- **Parallel cluster checks**: Uses goroutines + wait groups
- **Bounded rendering**: All views clip to terminal size
- **Lazy loading**: Resources fetched on demand
- **Efficient updates**: Only active pane receives input

## Testing Strategies

### Unit Tests
- Layout tree splitting logic
- Focus cycling algorithm
- Filter functions

### Integration Tests
- Mock K8s API responses
- Test view transitions
- Validate rendering output

### Manual Testing
```bash
# Test with multiple clusters
export KUBECONFIG=~/.kube/config1:~/.kube/config2

# Test with unreachable cluster
# Modify kubeconfig with invalid endpoint

# Test splits
# Press Ctrl+V, Ctrl+H multiple times

# Test filtering
# Press /, type partial name
```

## Future Enhancements

- [ ] Persistent pane layouts (save/load sessions)
- [ ] Custom themes and color schemes
- [ ] Plugin system for custom views
- [ ] Exec into pods (interactive shell)
- [ ] Port forwarding management
- [ ] YAML editing for resources
- [ ] Real-time metrics (CPU, memory)
- [ ] Log streaming with follow mode
- [ ] Context switching from UI
- [ ] Multi-select operations
- [ ] Search across all clusters
- [ ] Alerts and notifications
