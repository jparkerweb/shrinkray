# Install the locally built shrinkray binary to the system install directory.
# Usage: .\scripts\install-local.ps1

$ErrorActionPreference = "Stop"
Push-Location (Split-Path -Parent (Split-Path -Parent $PSCommandPath))

try {
    $InstallDir = Join-Path $env:LOCALAPPDATA "shrinkray"
    $Binary = "shrinkray.exe"

    # Build with version info
    $Version = (git describe --tags --always --dirty 2>$null) ?? "dev"
    $Commit = (git rev-parse --short HEAD 2>$null) ?? "none"
    $Date = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ" -AsUTC)
    $LdFlags = "-s -w -X main.version=$Version -X main.commit=$Commit -X main.date=$Date"

    Write-Host "Building shrinkray $Version..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    go build -ldflags $LdFlags -o $Binary ./cmd/shrinkray/
    if ($LASTEXITCODE -ne 0) { throw "Build failed" }

    # Ensure install directory exists
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Copy binary
    Copy-Item $Binary (Join-Path $InstallDir $Binary) -Force
    Write-Host "Installed to $InstallDir\$Binary" -ForegroundColor Green

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($UserPath -notlike "*$InstallDir*") {
        Write-Host "Adding $InstallDir to user PATH..." -ForegroundColor Cyan
        [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Added to PATH (restart terminal for changes to take effect)" -ForegroundColor Yellow
    }

    # Verify
    & (Join-Path $InstallDir $Binary) version

    # Clean up build artifact
    Remove-Item $Binary -ErrorAction SilentlyContinue

    Write-Host "Done!" -ForegroundColor Green
} finally {
    Pop-Location
}
