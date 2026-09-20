# Reset Neon dev database and re-seed dev admin (API boot).
# Usage (from backend/): powershell -ExecutionPolicy Bypass -File scripts/reset-dev-db.ps1

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

function Import-DotEnv {
    param([string]$Path = ".env")
    foreach ($line in Get-Content $Path) {
        if ($line -match '^\s*#' -or $line -match '^\s*$') { continue }
        if ($line -match '^([^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim().Trim('"')
            Set-Item -Path "env:$name" -Value $value
        }
    }
}

Import-DotEnv

if ([string]::IsNullOrWhiteSpace($env:POSTGRES_URL)) {
    Write-Error "POSTGRES_URL is not set in backend/.env"
}

Write-Host "==> Atlas schema clean (drops all objects)"
& atlas schema clean --url $env:POSTGRES_URL --auto-approve

Write-Host "==> Atlas migrate apply"
& atlas migrate apply --env local

Write-Host "==> Migration status"
& atlas migrate status --env local

Write-Host "==> Redis FLUSHALL"
docker exec boms-redis redis-cli FLUSHALL | Out-Null

Write-Host "==> Build API binary"
& go build -o api.exe ./cmd/api

Write-Host "==> Boot API to seed dev admin ($env:SEED_DEV_ADMIN_EMAIL)"
$api = Start-Process -FilePath ".\api.exe" -WorkingDirectory (Get-Location) -PassThru -WindowStyle Hidden

$healthy = $false
for ($i = 0; $i -lt 60; $i++) {
    Start-Sleep -Seconds 1
    try {
        $resp = Invoke-WebRequest -Uri "http://127.0.0.1:$($env:HTTP_PORT)/health" -UseBasicParsing -TimeoutSec 2
        if ($resp.StatusCode -eq 200) {
            $healthy = $true
            break
        }
    } catch {
        # API still starting
    }
}

if (-not $healthy) {
    Stop-Process -Id $api.Id -Force -ErrorAction SilentlyContinue
    Write-Error "API did not become healthy within 30s - check Redis and .env"
}

Stop-Process -Id $api.Id -Force -ErrorAction SilentlyContinue
Write-Host "==> Seed development catalog"
& go run ./cmd/seed

Write-Host '==> Done. Login with SEED_DEV_ADMIN_EMAIL and SEED_DEV_ADMIN_PASSWORD from .env'
