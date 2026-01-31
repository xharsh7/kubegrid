# kubegrid Usage Examples

## Basic Usage

### Start kubegrid
```bash
./kubegrid
```

### With specific kubeconfig
```bash
export KUBECONFIG=~/.kube/prod.config
./kubegrid
```

### With multiple kubeconfigs
```bash
export KUBECONFIG=~/.kube/prod.config:~/.kube/staging.config
./kubegrid
```

## Common Workflows

### 1. Monitoring Multiple Clusters

Monitor production and staging clusters side by side:

```bash
# Start kubegrid
./kubegrid

# You'll see cluster list:
# > prod-us-east
#   prod-eu-west
#   staging

# Press Ctrl+V to split vertically
# Navigate to first cluster, press Enter

# Press Tab to switch to right pane
# Navigate to second cluster, press Enter

# Now both clusters are visible!
```

**Keybindings:**
- `Ctrl+V` - Split vertically
- `Tab` - Switch panes
- `Enter` - Inspect cluster
- `r` - Refresh

---

### 2. Debugging a Failing Pod

Find and debug a pod with issues:

```bash
# In kubegrid:

# 1. Press '1' to view pods
# 2. Press '/' to search
# 3. Type part of pod name: "web-app"
# 4. Navigate to the pod
# 5. Press 'l' to view logs
# 6. Scroll logs with ↑↓
# 7. Press 'Esc' to go back

# If you need to delete it:
# 8. Press 'd' to delete
# 9. Confirm if prompted
```

**Keybindings:**
- `1` - View pods
- `/` - Filter
- `l` - View logs
- `d` - Delete pod
- `Esc` - Go back

---

### 3. Comparing Resources Across Namespaces

Check deployment status in multiple namespaces:

```bash
# In kubegrid:

# Pane 1: production namespace
# 1. Press '4' to view namespaces
# 2. Navigate to 'production'
# 3. Press 'n' to switch to it
# 4. Press '2' to view deployments

# Pane 2: staging namespace  
# 5. Press Ctrl+V to split
# 6. Press Tab to switch panes
# 7. Press '4' then navigate to 'staging'
# 8. Press 'n' then '2'

# Now compare deployments side by side!
```

---

### 4. Quick Cluster Health Check

Check all clusters are healthy:

```bash
./kubegrid

# Look at the STATE column:
# ✓ UP   = cluster healthy
# ✗ DOWN = cluster unreachable

# Check LATENCY column for performance
# < 50ms   = excellent
# 50-200ms = good
# > 200ms  = slow, investigate
```

---

### 5. Finding a Specific Service

Locate a service across all resources:

```bash
# In kubegrid:

# 1. Press '3' to view services
# 2. Press '/' to filter
# 3. Type service name: "api-gateway"
# 4. Press Enter to see details
```

---

### 6. Working with Command Mode

Use vim-style commands:

```bash
# In kubegrid, press ':' then type:

:vsplit    # Split vertically
:hsplit    # Split horizontally
:help      # Show help
:q         # Quit
```

---

### 7. Exploring a New Cluster

Get familiar with a cluster:

```bash
# In kubegrid:

# 1. Select cluster from list
# 2. Press Enter to inspect
# 3. Press '4' to view namespaces
# 4. Note the important namespaces
# 5. Press '1' to see all pods
# 6. Press '2' for deployments
# 7. Press '3' for services
# 8. Use 'r' frequently to refresh
```

---

## Advanced Usage

### Split Layout Patterns

#### Triple Split (monitoring 3 clusters)
```
1. Ctrl+V - split vertically
2. Tab then Ctrl+H - split horizontally
Result: 3 panes (left, top-right, bottom-right)
```

#### Quad Split (monitoring 4 things)
```
1. Ctrl+V - split vertically
2. Ctrl+H - split horizontally
3. Tab then Ctrl+V - split again
4. Tab then Ctrl+H - split again
Result: 4 panes in a grid
```

### Filtering Techniques

#### Exact match
```
/ + exact-pod-name
```

#### Partial match
```
/ + web    # matches web-app-123, web-api, etc.
```

#### Multiple filters (successive)
```
/ + prod     # filter for prod
Esc          # apply
/ + web      # further filter
```

### Log Analysis

#### View recent logs
```
l           # opens last 100 lines
↓↓↓         # scroll down
g           # jump to top
G           # jump to bottom
```

#### Search in logs
Currently logs are static. To search:
1. View logs with `l`
2. Visually scan
3. Use terminal's search (Ctrl+F in most terminals)

---

## Troubleshooting

### Cluster shows as DOWN

```bash
# Check connectivity
kubectl --context=problem-cluster cluster-info

# Check kubeconfig
cat ~/.kube/config | grep -A5 "name: problem-cluster"

# Test from kubegrid
# Select cluster, press Enter
# Check error message
```

### Can't see all clusters

```bash
# List your kubeconfigs
ls -la ~/.kube/

# kubegrid discovers files matching:
# - config
# - *.config

# Rename if needed:
mv ~/.kube/prod.yaml ~/.kube/prod.config
```

### Layout looks broken

```bash
# Resize terminal to be larger
# Minimum recommended: 80x24

# Or restart kubegrid
Ctrl+C
./kubegrid
```

### Performance is slow

```bash
# Check cluster latency in list view
# High latency clusters slow down the UI

# Options:
# 1. Remove slow clusters from kubeconfig
# 2. Increase timeout in code
# 3. Use splits to isolate slow clusters
```

---

## Tips & Tricks

### Tip 1: Muscle Memory
Learn these core keys:
- `j/k` for navigation (like vim)
- `r` to refresh anything
- `Esc` to go back
- `?` when stuck

### Tip 2: Keep Focus
Use splits to:
- Monitor logs while checking resources
- Compare two clusters
- Watch a deployment while checking pods

### Tip 3: Quick Actions
Common sequences:
- `1` + `/` + name + `l` = Quick log view
- `4` + navigate + `n` + `1` = Quick namespace switch
- `Ctrl+V` + `Tab` + `Enter` = Quick split compare

### Tip 4: Namespace Workflow
Create a namespace pane:
- Split vertically
- Left pane: keep on namespace list (press 4)
- Right pane: work in selected namespace
- Press `n` in left pane to switch right pane's namespace

### Tip 5: Emergency Actions
If something's wrong:
- `Ctrl+C` - force quit
- `q` - quit current view
- `Esc` - safe return
- `r` - refresh if data looks stale

---

## Keyboard Shortcuts Cheatsheet

```
NAVIGATION
  ↑/k         Move up
  ↓/j         Move down
  g           Jump to top
  G           Jump to bottom
  Tab         Next pane
  Shift+Tab   Previous pane

VIEWS
  1           Pods
  2           Deployments
  3           Services
  4           Namespaces
  Enter       Inspect
  Esc         Back

ACTIONS
  r           Refresh
  l           Logs (pods)
  d           Delete (pods)
  n           Switch namespace
  /           Filter
  ?           Help
  
WINDOWS
  Ctrl+V      Split vertical
  Ctrl+H      Split horizontal
  
COMMAND
  :           Enter command mode
  :q          Quit
  :vsplit     Split vertical
  :hsplit     Split horizontal
  :help       Show help

SYSTEM
  q           Quit view
  Ctrl+C      Force quit
```

---

## Environment Setup

### For Development
```bash
# Use local minikube
minikube start
export KUBECONFIG=~/.kube/config

# Run kubegrid
./kubegrid
```

### For Production Monitoring
```bash
# Set up dedicated kubeconfigs
export KUBECONFIG=~/.kube/prod-us.config:~/.kube/prod-eu.config

# Run kubegrid
./kubegrid
```

### For Testing
```bash
# Use test configs
export KUBECONFIG=./examples/kubeconfig-example.yaml

# Run
./kubegrid
```

---

Happy cluster managing! 🚀
