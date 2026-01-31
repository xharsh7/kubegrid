# 🚀 kubegrid Quick Start Guide

Get up and running with kubegrid in 2 minutes!

## Installation

```bash
# Clone the repo
git clone https://github.com/xharsh7/kubegrid.git
cd kubegrid

# Build
make build

# Or without make
go build -o kubegrid ./cmd/kubegrid
```

## First Run

```bash
# Make sure you have at least one Kubernetes cluster configured
kubectl config get-contexts

# Run kubegrid
./kubegrid
```

## What You'll See

```
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  KUBEGRID :: MULTI-CLUSTER OPERATIONS CONSOLE                              ┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃   CLUSTER        CONTEXT                                   STATE    LATENCY┃
┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫
┃ > minikube       minikube                                  UP       15ms   ┃
┃   docker-desktop docker-desktop                           UP       8ms    ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
```

## Essential Keybindings

### Must Know (First 5 minutes)
```
?          Show help screen (learn more keybindings)
↑/↓ or j/k Navigate up/down
Enter      Inspect selected cluster
1          View pods
r          Refresh
q          Quit current view
Ctrl+C     Exit application
```

### Navigation
```
g          Jump to top
G          Jump to bottom
Tab        Cycle between panes
Esc        Go back
```

### Views
```
1          Pods
2          Deployments  
3          Services
4          Namespaces
```

### Actions
```
/          Filter/search
l          View logs (when pod selected)
d          Delete pod
n          Switch namespace (from namespace list)
```

### Splits
```
Ctrl+V     Split vertically
Ctrl+H     Split horizontally
```

## 5-Minute Tutorial

### Step 1: View Pods (30 seconds)
```
1. Launch: ./kubegrid
2. Press '1' to view pods
3. Use ↑↓ or j/k to navigate
4. Press 'r' to refresh
```

### Step 2: Filter Resources (30 seconds)
```
1. Press '/' to start filtering
2. Type part of a pod name
3. Press Enter to apply filter
4. Press Esc to clear filter
```

### Step 3: View Logs (1 minute)
```
1. Navigate to a pod
2. Press 'l' to view logs
3. Use ↑↓ to scroll logs
4. Press 'q' or Esc to close logs
```

### Step 4: Split Panes (1 minute)
```
1. Press Ctrl+V to split vertically
2. Press Tab to switch to the new pane
3. Press '2' to view deployments in the right pane
4. Press Tab again to switch panes
5. Now you're monitoring pods and deployments simultaneously!
```

### Step 5: Switch Namespaces (1 minute)
```
1. Press '4' to view namespaces
2. Navigate to a namespace (e.g., 'kube-system')
3. Press 'n' to switch to it
4. Press '1' to view pods in that namespace
```

### Step 6: Multiple Clusters (1 minute)
```
1. From cluster list, press Enter on first cluster
2. Press Ctrl+V to split
3. Press Tab to switch panes
4. Press Esc to return to cluster list
5. Navigate to second cluster, press Enter
6. Now monitoring two clusters at once!
```

## Common Workflows

### Debug a Failing Pod
```
1. Press '1' (view pods)
2. Press '/' (filter)
3. Type pod name
4. Press 'l' (view logs)
5. Find the error
6. Press Esc (go back)
7. Press 'd' (delete if needed)
```

### Compare Two Clusters
```
1. Press Ctrl+V (split)
2. Select first cluster, press Enter
3. Press Tab
4. Select second cluster, press Enter
5. Compare side by side!
```

### Monitor Deployments and Pods
```
1. Press '2' (deployments)
2. Press Ctrl+H (split horizontal)
3. Press Tab
4. Press '1' (pods)
5. See deployments above, pods below!
```

## Tips for New Users

### Tip 1: Use Help
**Press `?` anytime** to see the complete help screen with all keybindings.

### Tip 2: Vim Users
If you know vim, you'll feel at home:
- `j/k` for navigation
- `g/G` for top/bottom
- `/` for search
- `:` for commands
- `Esc` to escape

### Tip 3: Stuck?
- **Esc** - Go back to previous view
- **q** - Quit current view
- **Ctrl+C** - Force quit application
- **?** - Show help

### Tip 4: Performance
If a cluster is slow:
- It will show in the LATENCY column
- You can still use kubegrid
- Just that cluster's operations will be slower

### Tip 5: Learning Path
1. Day 1: Learn basic navigation (↑↓, Enter, Esc)
2. Day 2: Learn views (1,2,3,4) and refresh (r)
3. Day 3: Learn filtering (/) and logs (l)
4. Day 4: Learn splits (Ctrl+V, Ctrl+H)
5. Day 5: Learn command mode (:vsplit, :help)

## Troubleshooting

### "No kubeconfigs found"
```bash
# Check your kubeconfig
ls ~/.kube/

# kubegrid looks for:
# - ~/.kube/config
# - ~/.kube/*.config

# Make sure you have at least one
kubectl config view
```

### "All clusters show DOWN"
```bash
# Test cluster connectivity
kubectl cluster-info

# Check kubeconfig is valid
kubectl get nodes

# Try with specific context
kubectl --context=your-context get nodes
```

### "Layout looks broken"
```bash
# Make sure terminal is at least 80x24
resize terminal

# Or restart kubegrid
Ctrl+C
./kubegrid
```

### "Can't split panes"
```bash
# Splits only work when inside a pane
# From cluster list:
1. Press Enter to inspect a cluster
2. Now press Ctrl+V to split
```

## Next Steps

Once comfortable with basics:

1. Read the full **README.md** for all features
2. Check **examples/USAGE_EXAMPLES.md** for workflows
3. Read **ARCHITECTURE.md** to understand design
4. Contribute! See **CONTRIBUTING.md**

## Getting Help

- **In-app**: Press `?` for help
- **Documentation**: Check README.md
- **Examples**: See examples/USAGE_EXAMPLES.md
- **Issues**: GitHub Issues for bugs
- **Discussions**: GitHub Discussions for questions

## Quick Reference Card

```
┌─────────────────────────────────────────────────────────┐
│                    KUBEGRID CHEATSHEET                  │
├─────────────────────────────────────────────────────────┤
│ HELP            ?     Show help screen                  │
│                                                         │
│ NAVIGATION      ↑↓    Move cursor                       │
│                 jk    Vim-style navigation              │
│                 g     Top                               │
│                 G     Bottom                            │
│                 Tab   Next pane                         │
│                                                         │
│ VIEWS           1     Pods                              │
│                 2     Deployments                       │
│                 3     Services                          │
│                 4     Namespaces                        │
│                 Enter Inspect                           │
│                                                         │
│ ACTIONS         r     Refresh                           │
│                 /     Filter                            │
│                 l     Logs (pods)                       │
│                 d     Delete (pods)                     │
│                 n     Switch namespace                  │
│                                                         │
│ SPLITS          ^V    Vertical split                    │
│                 ^H    Horizontal split                  │
│                                                         │
│ EXIT            Esc   Back                              │
│                 q     Quit view                         │
│                 ^C    Force quit                        │
└─────────────────────────────────────────────────────────┘
```

---

**You're ready to go! Launch kubegrid and press `?` if you get stuck. Happy cluster managing! 🚀**
