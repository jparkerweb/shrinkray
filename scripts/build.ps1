# Build shrinkray for the current platform (Windows PowerShell).
# Usage: .\scripts\build.ps1 [build|run|test|lint|ci|clean]

param(
    [ValidateSet("build", "run", "test", "lint", "ci", "clean")]
    [string]$Command = "build"
)

$ErrorActionPreference = "Stop"
Push-Location (Split-Path -Parent (Split-Path -Parent $PSCommandPath))

try {
    $BinaryName = "shrinkray.exe"
    $Version = (git describe --tags --always --dirty 2>$null) ?? "dev"
    $Commit = (git rev-parse --short HEAD 2>$null) ?? "none"
    $Date = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ" -AsUTC)
    $LdFlags = "-s -w -X main.version=$Version -X main.commit=$Commit -X main.date=$Date"

    switch ($Command) {
        "build" {
            Write-Host "Building $BinaryName $Version..."
            $env:CGO_ENABLED = "0"
            go build -ldflags $LdFlags -o $BinaryName ./cmd/shrinkray/
            if ($LASTEXITCODE -ne 0) { throw "Build failed" }
            Write-Host "Built .\$BinaryName"
        }
        "run" {
            Write-Host "Building and running $BinaryName..."
            $env:CGO_ENABLED = "0"
            go build -ldflags $LdFlags -o $BinaryName ./cmd/shrinkray/
            if ($LASTEXITCODE -ne 0) { throw "Build failed" }
            & ".\$BinaryName"
        }
        "test" {
            Write-Host "Running tests..."
            go test ./...
            if ($LASTEXITCODE -ne 0) { throw "Tests failed" }
        }
        "lint" {
            Write-Host "Running linter..."
            golangci-lint run ./...
            if ($LASTEXITCODE -ne 0) { throw "Lint failed" }
        }
        "ci" {
            Write-Host "Running CI pipeline (lint -> test -> build)..."
            golangci-lint run ./...
            if ($LASTEXITCODE -ne 0) { throw "Lint failed" }
            go test ./...
            if ($LASTEXITCODE -ne 0) { throw "Tests failed" }
            $env:CGO_ENABLED = "0"
            go build -ldflags $LdFlags -o $BinaryName ./cmd/shrinkray/
            if ($LASTEXITCODE -ne 0) { throw "Build failed" }
            Write-Host "CI passed."
        }
        "clean" {
            Write-Host "Cleaning..."
            Remove-Item -Force -ErrorAction SilentlyContinue shrinkray, shrinkray.exe
            Remove-Item -Recurse -Force -ErrorAction SilentlyContinue dist/
            Write-Host "Clean."
        }
    }
} finally {
    Pop-Location
}
