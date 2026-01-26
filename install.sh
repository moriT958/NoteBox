#!/bin/sh
set -e

REPO="moriT958/NoteBox"
BINARY="notebox"
INSTALL_DIR="${HOME}/.local/bin"

OS=$(uname -s)
ARCH=$(uname -m)

# Adjust for goreleaser naming rules
case "$ARCH" in
  x86_64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  i386|i686) ARCH="i386" ;;
esac

# Get latest version
VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

# Create download url for latest version
URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}_${OS}_${ARCH}.tar.gz"

# Make install dir
echo "Make install dir: ${INSTALL_DIR}"
mkdir -p "$INSTALL_DIR"

# Download
echo "Downloading ${BINARY} ${VERSION}..."
curl -sSfL "$URL" | tar xz -C /tmp

# Install
echo "Installing to ${INSTALL_DIR}..."
mv "/tmp/${BINARY}" "${INSTALL_DIR}/"
chmod +x "${INSTALL_DIR}/${BINARY}"

# Setup PATH
add_to_path() {
  SHELL_NAME=$(basename "$SHELL")
  case "$SHELL_NAME" in
    bash)
      RC_FILE="$HOME/.bashrc"
      ;;
    zsh)
      RC_FILE="$HOME/.zshrc"
      ;;
    *)
      RC_FILE=""
      ;;
  esac

  if [ -n "$RC_FILE" ] && [ -f "$RC_FILE" ]; then
    if ! grep -q "${INSTALL_DIR}" "$RC_FILE" 2>/dev/null; then
      echo "" >> "$RC_FILE"
      echo "# notebox" >> "$RC_FILE"
      echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "$RC_FILE"
      echo "Added ${INSTALL_DIR} to PATH in ${RC_FILE}"
      echo "Run 'source ${RC_FILE}' or restart your terminal"
    fi
  fi
}

case ":$PATH:" in
  *":${INSTALL_DIR}:"*)
    # already in PATH
    ;;
  *)
    add_to_path
    ;;
esac

echo "Install completed! Run 'notebox' to get started."
