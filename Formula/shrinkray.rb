# typed: false
# frozen_string_literal: true

# Homebrew formula for shrinkray
# Install: brew tap jparkerweb/tap && brew install shrinkray
class Shrinkray < Formula
  desc "Cross-platform CLI video compression tool powered by FFmpeg"
  homepage "https://github.com/jparkerweb/shrinkray"
  license "MIT"
  version "0.0.0"

  on_macos do
    on_intel do
      url "https://github.com/jparkerweb/shrinkray/releases/download/v#{version}/shrinkray_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_AMD64_SHA256"
    end

    on_arm do
      url "https://github.com/jparkerweb/shrinkray/releases/download/v#{version}/shrinkray_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_DARWIN_ARM64_SHA256"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/jparkerweb/shrinkray/releases/download/v#{version}/shrinkray_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64_SHA256"
    end

    on_arm do
      url "https://github.com/jparkerweb/shrinkray/releases/download/v#{version}/shrinkray_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64_SHA256"
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
