#  kubegrid

**A terminal-based Kubernetes multi-cluster operations dashboard inspired by tmux, k9s, and Neovim.**

Manage multiple Kubernetes clusters from a single, powerful terminal interface with tmux-style splits, vim-style navigation, and k9s-style resource visibility.

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃   CLUSTER        CONTEXT                                   STATE    LATENCY┃
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ > prod-us-east   eks-prod-us-east-1                        UP       45ms   ┃
┃   prod-eu-west   eks-prod-eu-west-1                        UP       123ms  ┃
┃   staging        gke-staging-central                       UP       28ms   ┃
┃   dev            minikube                                  DOWN     0s     ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## ✨ Features

### 🏗️ **Tmux-Style Layout Management**
- **Recursive pane splits** — Split any pane vertically or horizontally
- **Focus cycling** — Tab through panes like tmux
- **Size-aware rendering** — Automatic layout adjustments on terminal resize
- **Visual focus indicators** — Clear borders show active pane

###  **Vim-Style Interaction**
- **Command mode** — Press `:` to enter command mode
- **Modal interface** — Esc always returns to safe state
- **Keyboard-first** — No mouse required
- **Familiar keybindings** — j/k, g/G, /, etc.

### ☸️ **Multi-Cluster Kubernetes Management**
- **Automatic discovery** — Finds all kubeconfigs in `~/.kube/`
- **Parallel health checks** — Fast status collection across clusters
- **API latency** — Integer millisecond display per cluster
- **Error visualization** — Clear display of cluster issues

### 📦 **Resource Operations**
- **Browse resources** — Pods, Deployments, Services, Namespaces, Events, CRDs
- **Live filtering** — Instant search with `/`
- **Pod logs** — View logs with container picker for multi-container pods
- **Log follow mode** — Stream logs in real-time with `f`
- **Resource describe** — View YAML for any resource with `y`
- **Quick actions** — Delete pods, switch namespaces

### 🔍 **Advanced Features**
- **Split-pane monitoring** — Watch multiple clusters simultaneously
- **Custom Resource support** — Discover and browse CRD instances
- **Events view** — Namespace event timeline
- **Help system** — Press `?` for complete keybinding reference

---

## 🚀 Setup & Installation

### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| **Go** | 1.24+ | Required to build from source |
| **kubectl** | any | Must be configured with at least one cluster |
| **Terminal** | Unicode-capable | iTerm2, Alacritty, Kitty, GNOME Terminal, etc. |

### Step 1: Verify kubectl

```bash
kubectl config get-contexts
```

Make sure you have at least one cluster configured. kubegrid auto-discovers all
kubeconfig files in `~/.kube/` (files named `config` or `*.config`).

### Step 2: Install

#### Option A: Clone & Build (recommended)

```bash
git clone https://github.com/xharsh7/kubegrid.git
cd kubegrid
make build
./kubegrid
```

#### Option B: Go Install

```bash
go install github.com/xharsh7/kubegrid/cmd/kubegrid@latest
kubegrid
```

#### Option C: Download Binary

Grab a pre-built binary from the [Releases](https://github.com/xharsh7/kubegrid/releases) page:

```bash
# Linux x86_64
curl -LO https://github.com/xharsh7/kubegrid/releases/download/v1.0.1/kubegrid_1.0.1_linux_amd64.tar.gz
tar xzf kubegrid_1.0.1_linux_amd64.tar.gz
chmod +x kubegrid
./kubegrid
```

### Step 3: Verify

```bash
kubegrid --version
```

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
| `5` | View Events |
| `6` | View Custom Resources (CRDs) |
| `Enter` | Inspect cluster / CRD instances |
| `Esc` | Go back |

### Actions

| Key | Action |
|-----|--------|
| `r` | Refresh current view |
| `l` | View pod logs (container picker for multi-container pods) |
| `f` | Toggle log follow/streaming mode (while viewing logs) |
| `t` | Toggle timestamps in log output |
| `y` | View resource YAML / describe |
| `d` | Delete selected pod |
| `n` | Switch namespace (from namespace list) |
| `/` | Filter/search resources |
| `?` | Toggle help |
| `q` | Quit current view |
| `Ctrl+C` | Force quit |

### Window Management

| Key | Action |
|-----|--------|
| `Ctrl+B` or `Ctrl+V` | Split pane vertically (side by side) |
| `Ctrl+G` or `Ctrl+H` | Split pane horizontally (stacked) |
| `Ctrl+X` | Close active pane (when more than one) |

### Command Mode

| Command | Action |
|---------|--------|
| `:q` or `:quit` | Quit application |
| `:vsplit` | Split vertically |
| `:hsplit` | Split horizontally |
| `:close` | Close active pane |
| `:help` | Show available commands |

---

## 🎬 Example Workflows

### Monitor Multiple Clusters

1. Launch kubegrid: `./kubegrid`
2. Press `Ctrl+B` or `Ctrl+V` to split vertically
3. Press `Tab` to switch panes
4. Press `Enter` on different clusters in each pane
5. Monitor both simultaneously

### Debug a Pod

1. Press `1` to view pods
2. Press `/` and type pod name to filter
3. Navigate to pod and press `l` to view logs
4. Press `f` to follow logs in real-time
5. Press `Esc` to go back
6. Press `d` to delete if needed

### View Custom Resources

1. Press `6` to list available CRDs
2. Navigate to a CRD and press `Enter`
3. Browse instances of that CRD

### Describe a Resource

1. Navigate to any pod, deployment, or service
2. Press `y` to view its full YAML description

### Switch Namespaces

1. Press `4` to view namespaces
2. Navigate to desired namespace
3. Press `n` to switch to it
4. Press `1` to see pods in new namespace

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

### Make Targets

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make build-prod` | Build with production optimizations (stripped) |
| `make build-all` | Cross-compile for all platforms |
| `make run` | Build and run |
| `make dev` | Run with race detector |
| `make test` | Run tests |
| `make test-coverage` | Run tests with coverage report |
| `make lint` | Run linters (go vet + gofmt) |
| `make fmt` | Format code |
| `make tidy` | Tidy dependencies |
| `make install` | Install to `~/.local/bin` |
| `make uninstall` | Remove from `~/.local/bin` |
| `make clean` | Remove build artifacts |

### Running Tests

```bash
make test
# or
go test -v -race ./...
```

### Building from Source

```bash
make build
# or
go build -o kubegrid ./cmd/kubegrid
```

### Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Run tests (`make test`) and lint (`make lint`)
5. Submit a pull request

---

## 🔮 Roadmap

- [ ] Save/load pane layouts (sessions)
- [ ] Custom themes and colors
- [ ] Plugin system
- [ ] Exec into pods (interactive shell)
- [ ] Port forwarding UI
- [ ] YAML editing
- [ ] Real-time metrics (CPU/memory)
- [ ] Context switching from UI
- [ ] Multi-select operations
- [ ] Cross-cluster search
- [ ] Alerts and notifications

---

## 📝 License

MIT License — see [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Inspired by:
- **tmux** — Terminal multiplexer and pane management
- **k9s** — Kubernetes CLI management tool
- **Neovim** — Modal editing and command patterns
- **kubectl** — Kubernetes command-line tool

Built with:
- [bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) — Style rendering
- [client-go](https://github.com/kubernetes/client-go) — Kubernetes Go client

---

## 💬 Support

- **Issues**: [GitHub Issues](https://github.com/xharsh7/kubegrid/issues)
- **Discussions**: [GitHub Discussions](https://github.com/xharsh7/kubegrid/discussions)

---

Made with ❤️ for the Kubernetes community
