# install.ps1 — Install shrinkray on Windows.
#
# Usage:
#   irm https://raw.githubusercontent.com/jparkerweb/shrinkray/main/scripts/install.ps1 | iex
#
# Options (via environment variables):
#   $env:SHRINKRAY_INSTALL_DIR  — Installation directory (default: $env:LOCALAPPDATA\shrinkray)
#   $env:SHRINKRAY_VERSION      — Specific version to install (default: latest)

$ErrorActionPreference = "Stop"

$Repo = "jparkerweb/shrinkray"
$Binary = "shrinkray.exe"
$InstallDir = if ($env:SHRINKRAY_INSTALL_DIR) { $env:SHRINKRAY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "shrinkray" }

function Write-Info  { param($msg) Write-Host "[info]  $msg" -ForegroundColor Cyan }
function Write-Ok    { param($msg) Write-Host "[ok]    $msg" -ForegroundColor Green }
function Write-Warn  { param($msg) Write-Host "[warn]  $msg" -ForegroundColor Yellow }
function Write-Fail  { param($msg) Write-Host "[error] $msg" -ForegroundColor Red; exit 1 }

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { Write-Fail "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    $url = "https://api.github.com/repos/$Repo/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $url -UseBasicParsing
        return $response.tag_name -replace '^v', ''
    }
    catch {
        Write-Fail "Could not fetch latest version. Set `$env:SHRINKRAY_VERSION manually."
    }
}

function Main {
    $arch = Get-Arch
    Write-Info "Platform: windows/$arch"

    # Determine version
    if ($env:SHRINKRAY_VERSION) {
        $version = $env:SHRINKRAY_VERSION
        Write-Info "Installing specified version: v$version"
    }
    else {
        Write-Info "Fetching latest version..."
        $version = Get-LatestVersion
        Write-Info "Latest version: v$version"
    }

    # Build URLs
    $archiveName = "shrinkray_${version}_windows_${arch}.zip"
    $archiveUrl = "https://github.com/$Repo/releases/download/v$version/$archiveName"
    $checksumUrl = "https://github.com/$Repo/releases/download/v$version/checksums.txt"

    # Check for existing installation
    $existingPath = Get-Command "shrinkray" -ErrorAction SilentlyContinue
    if ($existingPath) {
        $existingVersion = & shrinkray version 2>&1 | Select-Object -First 1
        Write-Warn "shrinkray is already installed: $existingVersion"
        Write-Warn "Overwriting with v$version..."
    }

    # Create temp directory
    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) "shrinkray-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download archive
        Write-Info "Downloading $archiveName..."
        $archivePath = Join-Path $tmpDir $archiveName
        Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing

        # Download and verify checksum
        Write-Info "Verifying checksum..."
        $checksumPath = Join-Path $tmpDir "checksums.txt"
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing

        $checksumContent = Get-Content $checksumPath
        $expectedLine = $checksumContent | Where-Object { $_ -match $archiveName }
        if ($expectedLine) {
            $expectedSha = ($expectedLine -split '\s+')[0]
            $actualSha = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()
            if ($expectedSha -ne $actualSha) {
                Write-Fail "Checksum mismatch!`n  Expected: $expectedSha`n  Got:      $actualSha"
            }
            Write-Ok "Checksum verified"
        }
        else {
            Write-Warn "Could not find checksum for $archiveName - skipping verification"
        }

        # Extract
        Write-Info "Extracting..."
        Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

        # Install
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        Copy-Item (Join-Path $tmpDir $Binary) (Join-Path $InstallDir $Binary) -Force

        Write-Ok "Installed shrinkray to $InstallDir\$Binary"

        # Add to PATH if not already there
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            Write-Info "Adding $InstallDir to user PATH..."
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
            Write-Ok "Added to PATH (restart your terminal for changes to take effect)"
        }

        # Verify
        $installed = Join-Path $InstallDir $Binary
        & $installed version

        # Check for FFmpeg
        $ffmpeg = Get-Command "ffmpeg" -ErrorAction SilentlyContinue
        if (-not $ffmpeg) {
            Write-Host ""
            Write-Warn "FFmpeg is not installed. shrinkray requires FFmpeg to encode video."
            Write-Host ""
            Write-Host "  Install FFmpeg:"
            Write-Host "    scoop install ffmpeg"
            Write-Host "    choco install ffmpeg"
            Write-Host "    winget install ffmpeg"
            Write-Host ""
        }

        Write-Ok "Installation complete!"
    }
    finally {
        # Cleanup
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Main
