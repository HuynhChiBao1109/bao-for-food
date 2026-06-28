$ErrorActionPreference = "Stop"

Set-Location (Resolve-Path "$PSScriptRoot\..")

if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
  Write-Host "Air is not installed. Install it with:" -ForegroundColor Yellow
  Write-Host "  go install github.com/air-verse/air@latest" -ForegroundColor Cyan
  Write-Host ""
  Write-Host "After install, make sure GOPATH/bin is in PATH, then run this script again."
  exit 1
}

air -c .air.toml
