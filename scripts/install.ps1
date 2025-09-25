param (
    [string]$Repo = "JuanBrotenelle/eel-cli",
    [string]$ExeName = "eel.exe"
)

$installDir = "$env:LOCALAPPDATA\Programs\EelCli"

Write-Host "🔍 Fetching latest release from $Repo..."

$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers @{ "User-Agent" = "PowerShell" }

$asset = $release.assets | Where-Object { $_.name -eq $ExeName }
if (-not $asset) {
    Write-Error "❌ Not found $ExeName."
    exit 1
}

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

$exePath = Join-Path $installDir $ExeName

Write-Host "⬇️  Downloading $ExeName..."
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $exePath

$oldPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($oldPath -notlike "*$installDir*") {
    Write-Host "⚙️  Adding $installDir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$oldPath;$installDir", "User")
}

Write-Host "✅ Installation completed. Run '$ExeName'"
