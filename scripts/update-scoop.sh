#!/usr/bin/env bash
# update-scoop.sh — Generate Scoop manifest with correct version and SHA256 hashes.
#
# Usage:
#   ./scripts/update-scoop.sh <version>
#
# Example:
#   ./scripts/update-scoop.sh 1.2.3
#
# This script downloads the checksums.txt from the GitHub release,
# extracts the relevant SHA256 hashes, and generates the Scoop manifest.

set -euo pipefail

VERSION="${1:?Usage: $0 <version>}"
REPO="jparkerweb/shrinkray"
MANIFEST="shrinkray.json"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt"

echo "Fetching checksums for v${VERSION}..."
CHECKSUMS=$(curl -fsSL "${CHECKSUMS_URL}")

get_sha() {
  local pattern="$1"
  echo "${CHECKSUMS}" | grep "${pattern}" | awk '{print $1}'
}

SHA_WINDOWS_AMD64=$(get_sha "windows_amd64.zip")
SHA_WINDOWS_ARM64=$(get_sha "windows_arm64.zip")

if [[ -z "${SHA_WINDOWS_AMD64}" || -z "${SHA_WINDOWS_ARM64}" ]]; then
  echo "ERROR: Could not find all required SHA256 hashes in checksums.txt"
  echo "Available checksums:"
  echo "${CHECKSUMS}"
  exit 1
fi

cat > "${MANIFEST}" <<EOF
{
  "version": "${VERSION}",
  "description": "Cross-platform CLI video compression tool powered by FFmpeg. Less bytes, same vibes.",
  "homepage": "https://github.com/${REPO}",
  "license": "MIT",
  "architecture": {
    "64bit": {
      "url": "https://github.com/${REPO}/releases/download/v${VERSION}/shrinkray_${VERSION}_windows_amd64.zip",
      "hash": "${SHA_WINDOWS_AMD64}"
    },
    "arm64": {
      "url": "https://github.com/${REPO}/releases/download/v${VERSION}/shrinkray_${VERSION}_windows_arm64.zip",
      "hash": "${SHA_WINDOWS_ARM64}"
    }
  },
  "bin": "shrinkray.exe",
  "checkver": {
    "github": "https://github.com/${REPO}"
  },
  "autoupdate": {
    "architecture": {
      "64bit": {
        "url": "https://github.com/${REPO}/releases/download/v\$version/shrinkray_\$version_windows_amd64.zip"
      },
      "arm64": {
        "url": "https://github.com/${REPO}/releases/download/v\$version/shrinkray_\$version_windows_arm64.zip"
      }
    },
    "hash": {
      "url": "https://github.com/${REPO}/releases/download/v\$version/checksums.txt"
    }
  },
  "suggest": {
    "ffmpeg": {
      "source": "main"
    }
  }
}
EOF

echo "Updated ${MANIFEST} to v${VERSION}"
echo "  windows/amd64: ${SHA_WINDOWS_AMD64}"
echo "  windows/arm64: ${SHA_WINDOWS_ARM64}"
