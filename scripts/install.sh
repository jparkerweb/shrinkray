#!/usr/bin/env bash
# install.sh — Install shrinkray on macOS or Linux.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.sh | bash
#
# Options (via environment variables):
#   SHRINKRAY_INSTALL_DIR  — Installation directory (default: ~/.local/bin)
#   SHRINKRAY_VERSION      — Specific version to install (default: latest)

set -euo pipefail

REPO="jparkerweb/shrinkray"
BINARY="shrinkray"
INSTALL_DIR="${SHRINKRAY_INSTALL_DIR:-${HOME}/.local/bin}"

# Colors (disabled if not a terminal)
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  CYAN='\033[0;36m'
  NC='\033[0m'
else
  RED='' GREEN='' YELLOW='' CYAN='' NC=''
fi

info()  { echo -e "${CYAN}[info]${NC}  $*"; }
ok()    { echo -e "${GREEN}[ok]${NC}    $*"; }
warn()  { echo -e "${YELLOW}[warn]${NC}  $*"; }
fail()  { echo -e "${RED}[error]${NC} $*"; exit 1; }

# Detect OS
detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "${os}" in
    linux*)  echo "linux" ;;
    darwin*) echo "darwin" ;;
    *)       fail "Unsupported OS: ${os}. Use macOS or Linux." ;;
  esac
}

# Detect architecture
detect_arch() {
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             fail "Unsupported architecture: ${arch}" ;;
  esac
}

# Detect download tool
detect_downloader() {
  if command -v curl &>/dev/null; then
    echo "curl"
  elif command -v wget &>/dev/null; then
    echo "wget"
  else
    fail "Neither curl nor wget found. Please install one and retry."
  fi
}

# Download a URL to a file
download() {
  local url="$1" dest="$2"
  local dl
  dl="$(detect_downloader)"
  if [ "${dl}" = "curl" ]; then
    curl -fsSL -o "${dest}" "${url}"
  else
    wget -q -O "${dest}" "${url}"
  fi
}

# Get latest version from GitHub API
get_latest_version() {
  local url="https://api.github.com/repos/${REPO}/releases/latest"
  local dl
  dl="$(detect_downloader)"
  if [ "${dl}" = "curl" ]; then
    curl -fsSL "${url}" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'
  else
    wget -q -O - "${url}" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'
  fi
}

main() {
  local os arch version archive_name archive_url checksum_url
  local tmpdir

  info "Detecting platform..."
  os="$(detect_os)"
  arch="$(detect_arch)"
  info "Platform: ${os}/${arch}"

  # Determine version
  if [ -n "${SHRINKRAY_VERSION:-}" ]; then
    version="${SHRINKRAY_VERSION}"
    info "Installing specified version: v${version}"
  else
    info "Fetching latest version..."
    version="$(get_latest_version)"
    if [ -z "${version}" ]; then
      fail "Could not determine latest version. Set SHRINKRAY_VERSION manually."
    fi
    info "Latest version: v${version}"
  fi

  # Build URLs
  archive_name="${BINARY}_${version}_${os}_${arch}.tar.gz"
  archive_url="https://github.com/${REPO}/releases/download/v${version}/${archive_name}"
  checksum_url="https://github.com/${REPO}/releases/download/v${version}/checksums.txt"

  # Check for existing installation
  if command -v "${BINARY}" &>/dev/null; then
    local existing
    existing="$(${BINARY} version 2>/dev/null | head -1 || echo "unknown")"
    warn "shrinkray is already installed: ${existing}"
    warn "Overwriting with v${version}..."
  fi

  # Create temp directory
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  # Download archive
  info "Downloading ${archive_name}..."
  download "${archive_url}" "${tmpdir}/${archive_name}"

  # Download and verify checksum
  info "Verifying checksum..."
  download "${checksum_url}" "${tmpdir}/checksums.txt"

  local expected_sha actual_sha
  expected_sha="$(grep "${archive_name}" "${tmpdir}/checksums.txt" | awk '{print $1}')"
  if [ -n "${expected_sha}" ]; then
    if command -v sha256sum &>/dev/null; then
      actual_sha="$(sha256sum "${tmpdir}/${archive_name}" | awk '{print $1}')"
    elif command -v shasum &>/dev/null; then
      actual_sha="$(shasum -a 256 "${tmpdir}/${archive_name}" | awk '{print $1}')"
    else
      warn "No sha256sum or shasum found — skipping checksum verification"
      actual_sha="${expected_sha}"
    fi

    if [ "${expected_sha}" != "${actual_sha}" ]; then
      fail "Checksum mismatch!\n  Expected: ${expected_sha}\n  Got:      ${actual_sha}"
    fi
    ok "Checksum verified"
  else
    warn "Could not find checksum for ${archive_name} — skipping verification"
  fi

  # Extract
  info "Extracting..."
  tar -xzf "${tmpdir}/${archive_name}" -C "${tmpdir}"

  # Install
  mkdir -p "${INSTALL_DIR}"
  if [ -w "${INSTALL_DIR}" ]; then
    cp "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    chmod +x "${INSTALL_DIR}/${BINARY}"
  else
    info "Elevated permissions needed to install to ${INSTALL_DIR}"
    sudo cp "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY}"
  fi

  ok "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

  # Verify
  if command -v "${INSTALL_DIR}/${BINARY}" &>/dev/null; then
    "${INSTALL_DIR}/${BINARY}" version
  else
    echo ""
    warn "${INSTALL_DIR} is not in your PATH."
    echo ""
    echo "  Add it to your shell profile:"
    echo "    export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
  fi

  # Check for FFmpeg
  if ! command -v ffmpeg &>/dev/null; then
    echo ""
    warn "FFmpeg is not installed. shrinkray requires FFmpeg to encode video."
    echo ""
    echo "  Install FFmpeg:"
    if [ "${os}" = "darwin" ]; then
      echo "    brew install ffmpeg"
    else
      echo "    sudo apt install ffmpeg       # Debian/Ubuntu"
      echo "    sudo dnf install ffmpeg        # Fedora"
      echo "    sudo pacman -S ffmpeg          # Arch"
    fi
    echo ""
  fi

  ok "Installation complete!"
}

main "$@"
