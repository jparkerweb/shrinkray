#!/usr/bin/env bash
# update-homebrew.sh — Generate Homebrew formula with correct version and SHA256 hashes.
#
# Usage:
#   ./scripts/update-homebrew.sh <version>
#
# Example:
#   ./scripts/update-homebrew.sh 1.2.3
#
# This script downloads the checksums.txt from the GitHub release,
# extracts the relevant SHA256 hashes, and generates the Homebrew formula.

set -euo pipefail

VERSION="${1:?Usage: $0 <version>}"
REPO="jparkerweb/shrinkray"
FORMULA="Formula/shrinkray.rb"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/v${VERSION}/checksums.txt"

echo "Fetching checksums for v${VERSION}..."
CHECKSUMS=$(curl -fsSL "${CHECKSUMS_URL}")

get_sha() {
  local pattern="$1"
  echo "${CHECKSUMS}" | grep "${pattern}" | awk '{print $1}'
}

SHA_DARWIN_AMD64=$(get_sha "darwin_amd64.tar.gz")
SHA_DARWIN_ARM64=$(get_sha "darwin_arm64.tar.gz")
SHA_LINUX_AMD64=$(get_sha "linux_amd64.tar.gz")
SHA_LINUX_ARM64=$(get_sha "linux_arm64.tar.gz")

if [[ -z "${SHA_DARWIN_AMD64}" || -z "${SHA_DARWIN_ARM64}" || -z "${SHA_LINUX_AMD64}" || -z "${SHA_LINUX_ARM64}" ]]; then
  echo "ERROR: Could not find all required SHA256 hashes in checksums.txt"
  echo "Available checksums:"
  echo "${CHECKSUMS}"
  exit 1
fi

cat > "${FORMULA}" <<EOF
# typed: false
# frozen_string_literal: true

# Homebrew formula for shrinkray
# Install: brew tap jparkerweb/tap && brew install shrinkray
class Shrinkray < Formula
  desc "Cross-platform CLI video compression tool powered by FFmpeg"
  homepage "https://github.com/${REPO}"
  license "MIT"
  version "${VERSION}"

  on_macos do
    on_intel do
      url "https://github.com/${REPO}/releases/download/v#{version}/shrinkray_#{version}_darwin_amd64.tar.gz"
      sha256 "${SHA_DARWIN_AMD64}"
    end

    on_arm do
      url "https://github.com/${REPO}/releases/download/v#{version}/shrinkray_#{version}_darwin_arm64.tar.gz"
      sha256 "${SHA_DARWIN_ARM64}"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/${REPO}/releases/download/v#{version}/shrinkray_#{version}_linux_amd64.tar.gz"
      sha256 "${SHA_LINUX_AMD64}"
    end

    on_arm do
      url "https://github.com/${REPO}/releases/download/v#{version}/shrinkray_#{version}_linux_arm64.tar.gz"
      sha256 "${SHA_LINUX_ARM64}"
    end
  end

  depends_on "ffmpeg"

  def install
    bin.install "shrinkray"
  end

  def caveats
    <<~EOS
      shrinkray requires FFmpeg to be installed and available on your PATH.
      It has been installed as a dependency, but if you encounter issues:
        brew install ffmpeg
    EOS
  end

  test do
    assert_match "shrinkray", shell_output("#{bin}/shrinkray version")
  end
end
EOF

echo "Updated ${FORMULA} to v${VERSION}"
echo "  darwin/amd64: ${SHA_DARWIN_AMD64}"
echo "  darwin/arm64: ${SHA_DARWIN_ARM64}"
echo "  linux/amd64:  ${SHA_LINUX_AMD64}"
echo "  linux/arm64:  ${SHA_LINUX_ARM64}"
