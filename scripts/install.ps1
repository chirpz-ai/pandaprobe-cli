<#
.SYNOPSIS
    pandaprobe CLI installer for Windows.

.DESCRIPTION
    Downloads the latest pandaprobe release binary and installs it to the user's
    local application data, adding it to the user PATH.

        irm https://cli.pandaprobe.com/install.ps1 | iex

.PARAMETER Version
    Install a specific version (e.g. v0.2.0). Defaults to the latest release.
    Can also be set via the PANDAPROBE_VERSION environment variable.

.PARAMETER InstallDir
    Install location. Defaults to $env:LOCALAPPDATA\pandaprobe\bin.
    Can also be set via the PANDAPROBE_INSTALL_DIR environment variable.

.PARAMETER BaseUrl
    Release download root for mirrors/testing. Defaults to the GitHub releases
    URL. Can also be set via the PANDAPROBE_BASE_URL environment variable.
#>
[CmdletBinding()]
param(
    [string]$Version = $env:PANDAPROBE_VERSION,
    [string]$InstallDir = $env:PANDAPROBE_INSTALL_DIR,
    [string]$BaseUrl = $env:PANDAPROBE_BASE_URL
)

$ErrorActionPreference = 'Stop'
$Repo = 'chirpz-ai/pandaprobe-cli'
$Project = 'pandaprobe-cli'
$Binary = 'pandaprobe.exe'

function Write-Info($msg) { Write-Host "==> $msg" }

# --- detect architecture (Windows release ships amd64 only) ---
$arch = if ([Environment]::Is64BitOperatingSystem) { 'amd64' } else {
    throw 'pandaprobe requires a 64-bit version of Windows.'
}

# --- resolve version ---
if ([string]::IsNullOrWhiteSpace($Version) -or $Version -eq 'latest') {
    Write-Info 'Resolving latest release'
    $release = Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$Repo/releases/latest"
    $tag = $release.tag_name
    if (-not $tag) { throw 'Could not determine the latest release; set -Version explicitly.' }
} else {
    $tag = $Version
}
$verNoV = $tag -replace '^v', ''

$asset = "${Project}_${verNoV}_windows_${arch}.zip"
if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
    $BaseUrl = "https://github.com/$Repo/releases/download"
}
$releaseUrl = "$BaseUrl/$tag"

# --- install dir ---
if ([string]::IsNullOrWhiteSpace($InstallDir)) {
    $InstallDir = Join-Path $env:LOCALAPPDATA 'pandaprobe\bin'
}
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# --- download ---
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("pandaprobe-" + [System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
try {
    $zipPath = Join-Path $tmp $asset
    Write-Info "Downloading $asset ($tag)"
    Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/$asset" -OutFile $zipPath

    # --- verify checksum ---
    # Downloading checksums.txt is best-effort (a network error only warns), but a
    # checksum *mismatch* is always fatal — never install a tampered binary.
    $sumsPath = Join-Path $tmp 'checksums.txt'
    $haveSums = $false
    try {
        Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/checksums.txt" -OutFile $sumsPath
        $haveSums = $true
    } catch {
        Write-Warning "Could not download checksums.txt; skipping verification: $($_.Exception.Message)"
    }
    if ($haveSums) {
        $line = Select-String -Path $sumsPath -Pattern ([regex]::Escape($asset)) | Select-Object -First 1
        if ($line) {
            $expected = ($line.Line -split '\s+')[0]
            $actual = (Get-FileHash -Algorithm SHA256 -Path $zipPath).Hash.ToLower()
            if ($expected.ToLower() -ne $actual) {
                throw "Checksum mismatch for $asset (expected $expected, got $actual)"
            }
            Write-Info 'Checksum verified'
        } else {
            Write-Warning "No checksum entry for $asset; skipping verification."
        }
    }

    # --- extract ---
    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    $src = Join-Path $tmp $Binary
    if (-not (Test-Path $src)) { throw "Binary $Binary not found in archive." }

    $dest = Join-Path $InstallDir $Binary
    Copy-Item -Path $src -Destination $dest -Force
    Write-Info "Installed pandaprobe ($tag) to $dest"
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

# --- ensure InstallDir is on the user PATH ---
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $InstallDir) {
    $newPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    $env:Path = "$env:Path;$InstallDir"
    Write-Info "Added $InstallDir to your user PATH (restart your shell to pick it up everywhere)."
}

Write-Host ''
Write-Host 'Run `pandaprobe --help` to get started.'
