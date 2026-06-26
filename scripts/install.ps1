# git-lan installer (Windows, PowerShell).
# Builds from source and installs git-lan.exe so that `git lan` works (git
# resolves `git lan` to a `git-lan` executable on PATH).

#Requires -Version 5.0
$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$Binary   = "git-lan.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\git-lan"

Write-Host "==> git-lan installer"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go toolchain not found. Install Go 1.26+ from https://go.dev/dl/"
    exit 1
}
if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Error "git not found. git-lan shells out to the git binary."
    exit 1
}

$Version = (git -C $RepoRoot describe --tags --always --dirty 2>$null)
if (-not $Version) { $Version = "dev" }
$LdFlags = "-s -w -X github.com/matixandr/git-lan/cmd.Version=$Version"

Write-Host "==> building $Binary ($Version)"
Push-Location $RepoRoot
try {
    go build -ldflags $LdFlags -o $Binary .
} finally {
    Pop-Location
}

Write-Host "==> installing to $InstallDir"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Move-Item -Force (Join-Path $RepoRoot $Binary) (Join-Path $InstallDir $Binary)

# Add to the user PATH if missing.
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    Write-Host "note: added $InstallDir to your user PATH (restart your shell)."
}

Write-Host "==> done. $Binary -> $InstallDir\$Binary"
Write-Host "Try: git lan list"
Write-Host "If 'git lan' is not found, ensure Bonjour/mDNS is available on this network."
