#!/usr/bin/env sh
set -e

# kubegrid quick-install script
# Usage: curl -fL https://raw.githubusercontent.com/xharsh7/kubegrid/main/scripts/install.sh | sh
#   or:  curl -fL https://raw.githubusercontent.com/xharsh7/kubegrid/main/scripts/install.sh | sh -s -- --version v1.0.1

REPO="xharsh7/kubegrid"
BINARY="kubegrid"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Parse arguments
VERSION=""
for arg in "$@"; do
  case "$arg" in
    --version=*) VERSION="${arg#*=}" ;;
    --version)   VERSION="$2"; shift ;;
    -v)          VERSION="$2"; shift ;;
  esac
done

# Detect OS
detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux|darwin) echo "$os" ;;
    *) echo "unsupported OS: $os" >&2; exit 1 ;;
  esac
}

# Detect arch
detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
  esac
}

# Get latest version from GitHub API if not specified
get_latest_version() {
  if command -v curl >/dev/null 2>&1; then
    curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed 's/.*"v\([^"]*\)".*/\1/'
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed 's/.*"v\([^"]*\)".*/\1/'
  else
    echo "error: curl or wget required" >&2; exit 1
  fi
}

# Main install logic
main() {
  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ -z "$VERSION" ]; then
    echo "Fetching latest version..."
    VERSION="$(get_latest_version)"
    if [ -z "$VERSION" ]; then
      echo "error: could not determine latest version" >&2; exit 1
    fi
  fi

  FILENAME="${BINARY}_${VERSION}_${os}_${arch}.tar.gz"
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

  echo "Installing kubegrid v${VERSION} for ${os}/${arch}..."
  echo "Downloading ${DOWNLOAD_URL}..."

  TMPDIR="$(mktemp -d)"
  trap 'rm -rf "$TMPDIR"' EXIT

  if command -v curl >/dev/null 2>&1; then
    curl -fL "$DOWNLOAD_URL" -o "${TMPDIR}/${FILENAME}"
  else
    wget -q "$DOWNLOAD_URL" -O "${TMPDIR}/${FILENAME}"
  fi

  tar xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

  # Ensure install directory exists
  mkdir -p "$INSTALL_DIR"

  # Install binary
  cp "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"

  echo "kubegrid v${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
  echo "Run 'kubegrid --version' to verify"
}

main
