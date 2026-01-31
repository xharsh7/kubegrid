# 🎯 kubegrid

**A terminal-based Kubernetes operations dashboard inspired by tmux, k9s, and Neovim.**

Manage multiple Kubernetes clusters from a single, powerful terminal interface with tmux-style splits, vim-style navigation, and k9s-style resource visibility.

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃   CLUSTER        CONTEXT                                   STATE    LATENCY┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ > prod-us-east   eks-prod-us-east-1                        UP       45ms   ┃
┃   prod-eu-west   eks-prod-eu-west-1                        UP       123ms  ┃
┃   staging        gke-staging-central                       UP       28ms   ┃
┃   dev            minikube                                  DOWN     0s     ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## ✨ Features

### 🏗️ **Tmux-Style Layout Management**
- **Recursive pane splits** - Split any pane vertically or horizontally
- **Focus cycling** - Tab through panes like tmux
- **Size-aware rendering** - Automatic layout adjustments on terminal resize
- **Visual focus indicators** - Clear borders show active pane

### 🎮 **Vim-Style Interaction**
- **Command mode** - Press `:` to enter command mode
- **Modal interface** - Esc always returns to safe state
- **Keyboard-first** - No mouse required
- **Familiar keybindings** - j/k, g/G, /, etc.

### ☸️ **Multi-Cluster Kubernetes Management**
- **Automatic discovery** - Finds all kubeconfigs in ~/.kube/
- **Parallel health checks** - Fast status collection across clusters
- **Real-time metrics** - API latency and reachability
- **Error visualization** - Clear display of cluster issues

### 📦 **Resource Operations**
- **Browse resources** - Pods, Deployments, Services, Namespaces
- **Live filtering** - Instant search with `/`
- **Log viewing** - Stream pod logs with `l`
- **Quick actions** - Delete pods, switch namespaces
- **Namespace switching** - Navigate to any namespace

### 🔍 **Advanced Features**
- **Split-pane monitoring** - Watch multiple clusters simultaneously
- **Context preservation** - No more lost terminal state
- **Fullscreen modes** - Focus on what matters
- **Help system** - Press `?` for complete keybinding reference

---

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/xharsh7/kubegrid.git
cd kubegrid

# Build
go build -o kubegrid ./cmd/kubegrid

# Run
./kubegrid
```

### Prerequisites

- Go 1.25+
- kubectl configured with at least one cluster
- Terminal with good Unicode support

---

## 📖 Usage Guide

### Navigation

| Key | Action |
|-----|--------|
| `↑`/`k` | Move cursor up |
| `↓`/`j` | Move cursor down |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Tab` | Cycle focus between panes |
| `Shift+Tab` | Cycle focus backwards |

### Resource Views

| Key | Action |
|-----|--------|
| `1` | View Pods |
| `2` | View Deployments |
| `3` | View Services |
| `4` | View Namespaces |
| `Enter` | Inspect selected item |
| `Esc` | Go back |

### Actions

| Key | Action |
|-----|--------|
| `r` | Refresh current view |
| `l` | View logs (pods only) |
| `d` | Delete selected pod |
| `n` | Switch namespace (from namespace list) |
| `/` | Filter/search |
| `?` | Toggle help |
| `q` | Quit current view |
| `Ctrl+C` | Force quit |

### Window Management

| Key | Action |
|-----|--------|
| `Ctrl+V` | Split pane vertically |
| `Ctrl+H` | Split pane horizontally |

### Command Mode

| Command | Action |
|---------|--------|
| `:q` or `:quit` | Quit application |
| `:vsplit` | Split vertically |
| `:hsplit` | Split horizontally |
| `:help` | Show help |

---

## 🎬 Example Workflows

### Monitor Multiple Clusters

1. Launch kubegrid: `./kubegrid`
2. Press `Ctrl+V` to split vertically
3. Press `Tab` to switch panes
4. Press `Enter` on different clusters in each pane
5. Monitor both simultaneously

### Debug a Pod

1. Press `1` to view pods
2. Press `/` and type pod name to filter
3. Navigate to pod and press `l` to view logs
4. Press `Esc` to go back
5. Press `d` to delete if needed

### Switch Namespaces

1. Press `4` to view namespaces
2. Navigate to desired namespace
3. Press `n` to switch to it
4. Press `1` to see pods in new namespace

### Explore a Cluster

1. Select a cluster from the list
2. Press `Enter` to inspect
3. Press `1`, `2`, `3` to browse different resources
4. Use `/` to filter resources
5. Press `r` to refresh data

---

## 🧠 Design Principles

### Why kubegrid?

Managing multiple Kubernetes clusters typically means:
- ❌ Juggling many kubeconfig files
- ❌ Constantly running `kubectl config use-context`
- ❌ Opening/closing multiple terminals
- ❌ Losing context while debugging

kubegrid solves this by:
- ✅ Loading all clusters at once
- ✅ Single interactive TUI
- ✅ Context preservation
- ✅ Side-by-side monitoring

### Architecture Highlights

**1. Layout Tree** (tmux-inspired)
- Binary tree structure for panes
- Recursive splitting algorithm
- Automatic size calculation
- Efficient rendering

**2. Focused Pane Model**
- Single active pane
- Input routing to focus
- Tab-based cycling
- Predictable behavior

**3. Screen Separation** (k9s/vim-inspired)
- Embedded panes for persistent views
- Fullscreen screens for modals
- Clean responsibility separation
- Esc-friendly navigation

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed design documentation.

---

## 🛠️ Development

### Project Structure

```
kubegrid/
├── cmd/kubegrid/        # Application entry point
├── internal/
│   ├── cluster/         # Cluster status collection
│   ├── config/          # Kubeconfig management
│   └── tui/             # Terminal UI components
└── pkg/k8s/            # Kubernetes client wrapper
```

### Running Tests

```bash
go test ./...
```

### Building from Source

```bash
go build -o kubegrid ./cmd/kubegrid
```

### Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## 🔮 Roadmap

- [ ] Save/load pane layouts
- [ ] Custom themes and colors
- [ ] Plugin system
- [ ] Exec into pods
- [ ] Port forwarding UI
- [ ] YAML editing
- [ ] Real-time metrics (CPU/memory)
- [ ] Log streaming with follow
- [ ] Context switching from UI
- [ ] Multi-select operations
- [ ] Cross-cluster search
- [ ] Alerts and notifications

---

## 📝 License

MIT License - see LICENSE file for details

---

## 🙏 Acknowledgments

Inspired by:
- **tmux** - Terminal multiplexer and pane management
- **k9s** - Kubernetes CLI management tool
- **Neovim** - Modal editing and command patterns
- **kubectl** - Kubernetes command-line tool

Built with:
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style rendering
- [client-go](https://github.com/kubernetes/client-go) - Kubernetes Go client

---

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/xharsh7/kubegrid/issues)
- **Discussions**: [GitHub Discussions](https://github.com/xharsh7/kubegrid/discussions)

---

Made with ❤️ for the Kubernetes community
