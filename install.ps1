$ErrorActionPreference = "Stop"

# shortk install script for Windows
# Usage: irm https://raw.githubusercontent.com/Hilaladiii/shortk/main/install.ps1 | iex

# --- CONFIGURATION ---
$Repo = "Hilaladiii/shortk"
# ---------------------

$Os = "windows"
$Arch = "amd64" # Default for Windows. Arm64 builds are not currently provided in release.yml

$BinaryName = "shortk-$Os-$Arch.exe"

Write-Host "Checking for latest release of shortk..."
$ReleaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
$LatestRelease = $null

try {
    $ReleaseInfo = Invoke-RestMethod -Uri $ReleaseUrl -Headers @{"User-Agent"="shortk-installer"}
    $LatestRelease = $ReleaseInfo.tag_name
} catch {
    Write-Host "Warning: Failed to fetch latest release version from GitHub API."
}

$Downloaded = $false
if ($null -ne $LatestRelease) {
    Write-Host "Found version $LatestRelease. Downloading $BinaryName..."
    $DownloadUrl = "https://github.com/$Repo/releases/download/$LatestRelease/$BinaryName"
    $TempFile = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), $BinaryName)

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempFile
        Write-Host "Successfully downloaded binary."
        $Downloaded = $true
    } catch {
        Write-Host "Warning: Failed to download pre-built binary. Falling back to source build..."
    }
}

# Fallback: Build from source if download failed or no release found
if (-not $Downloaded) {
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-Host "Building shortk from source..."
        $SourceDir = Get-Location
        # Check if we are in the project root
        if (-not (Test-Path "main.go")) {
            $TempSource = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "shortk-source-" + [System.Guid]::NewGuid().ToString().Substring(0,8))
            New-Item -ItemType Directory -Path $TempSource | Out-Null
            Set-Location $TempSource
            Write-Host "Cloning source code..."
            git clone --depth 1 "https://github.com/$Repo.git" .
        }
        
        go build -o shortk.exe main.go
        $TempFile = Join-Path (Get-Location) "shortk.exe"
        $Downloaded = $true
    } else {
        Write-Error "Error: Could not download pre-built binary and Go is not installed."
        exit 1
    }
}

# Determine install directory
$InstallDir = Join-Path $env:USERPROFILE ".local\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Write-Host "Installing to $InstallDir..."
$TargetPath = Join-Path $InstallDir "shortk.exe"
Move-Item -Path $TempFile -Destination $TargetPath -Force

# Add to User PATH if not already present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    Write-Host "Adding $InstallDir to User PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

# Initialize
Write-Host "Initializing shortk..."
& $TargetPath init

Write-Host "shortk installed successfully!"
Write-Host "Note: You may need to restart your terminal or run '. `$PROFILE' for PATH changes to take full effect."
