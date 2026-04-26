$ErrorActionPreference = "Stop"

# shortk uninstall script for Windows
# Usage: irm https://raw.githubusercontent.com/Hilaladiii/shortk/main/uninstall.ps1 | iex

$InstallDir = Join-Path $env:USERPROFILE ".local\bin"
$BinaryPath = Join-Path $InstallDir "shortk.exe"

if (Test-Path $BinaryPath) {
    Write-Host "Removing shortk binary..."
    Remove-Item $BinaryPath -Force
}

# Remove from User PATH
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -match [regex]::Escape($InstallDir)) {
    Write-Host "Removing $InstallDir from User PATH..."
    $NewPath = ($UserPath -split ";" | Where-Object { $_ -ne $InstallDir }) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
}

# Clean up profile
$Profiles = @(
    Join-Path $env:USERPROFILE "Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
    Join-Path $env:USERPROFILE "Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
)

$StartMarker = "# <<< shortk initialize <<<"
$EndMarker = "# >>> shortk initialize >>>"

foreach ($Profile in $Profiles) {
    if (Test-Path $Profile) {
        Write-Host "Cleaning up profile: $Profile"
        $Content = Get-Content $Profile
        $NewContent = @()
        $Skip = $false
        foreach ($Line in $Content) {
            if ($Line -eq $StartMarker) { $Skip = $true; continue }
            if ($Line -eq $EndMarker) { $Skip = $false; continue }
            if (-not $Skip) { $NewContent += $Line }
        }
        $NewContent | Set-Content $Profile
    }
}

Write-Host "shortk has been successfully uninstalled."
