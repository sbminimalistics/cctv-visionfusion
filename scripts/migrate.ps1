# scripts/migrate.ps1
param(
    [string]$Direction = "up" # "up" or "down"
)

# dynamic root path alignment
# $PSScriptRoot is always the folder where this script lives (e.g., /scripts)
$scriptDir = $PSScriptRoot
$projectRootDir = Split-Path -Path $scriptDir -Parent

# force the active execution context back out to the true project root
Set-Location -Path $projectRootDir
Write-Host "[dev] workspace root aligned to: $projectRootDir" -ForegroundColor Gray

$ErrorActionPreference = "Stop"

$envFilePath = "./.env_dev"
if (Test-Path $envFilePath) {
    Get-Content $envFilePath | ForEach-Object {
        $line = $_.Trim()
        # skip comments or empty lines, but make sure the line contains an equals sign
        if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
            # use native powershell -split which gives a valid array structure
            $parts = $line -split '=', 2
            
            $key   = $parts[0].Trim()
            $value = $parts[1].Trim()
            
            Set-Item -Path "env:$key" -Value $value
        }
    }
    Write-Host "[dev] successfully ingested environment configurations out of .env_dev" -ForegroundColor Gray
} else {
    Write-Host "[dev] warning: .env_dev file not found at project layer base" -ForegroundColor Red
}

$startedTemporaryDb = $false
$runningContainer = docker ps -q -f "name=^db-dev$"

try {
    if (-not $runningContainer) {
        Write-Host "[0/2] Starting temporary db-dev container..." -ForegroundColor Yellow
        docker rm -f db-dev 2>$null | Out-Null

        $dbName = if (${env:DB_NAME}) { $env:DB_NAME } else {"cctv-vf_database"}
        $dbPort = if (${env:DB_PORT}) { $env:DB_PORT } else {"5432"}
        $backendUser = $env:BACKEND_DB_USER
        $backendPass = $env:BACKEND_DB_PASS
        $backendSchema = $env:BACKEND_DB_SCHEMA
        $dashboardUser = $env:DASHBOARD_DB_USER
        $dashboardPass = $env:DASHBOARD_DB_PASS
        $dashboardSchema = $env:DASHBOARD_DB_SCHEMA
        $updaterUser = $env:MIGRATION_DB_USER
        $updaterPass = $env:MIGRATION_DB_PASS

        $initSqlPath = "${PSScriptRoot}/init-multi-user.sql"
        $sqlContent = @"
-- Create application-specific roles/users
CREATE USER ${updaterUser} WITH PASSWORD '${updaterPass}';
CREATE USER ${backendUser} WITH PASSWORD '${backendPass}';
CREATE USER ${dashboardUser} WITH PASSWORD '${dashboardPass}';

-- provision schemas
CREATE SCHEMA ${backendSchema} AUTHORIZATION ${backendUser};
CREATE SCHEMA ${dashboardSchema} AUTHORIZATION ${dashboardUser};

-- grant the migrator permissions to create things inside both schemas
GRANT CREATE, USAGE ON SCHEMA ${backendSchema} TO ${updaterUser};
GRANT CREATE, USAGE ON SCHEMA ${dashboardSchema} TO ${updaterUser};

-- grant targeted schema permissions to maintain complete application safety boundaries
GRANT ALL ON SCHEMA ${backendSchema} TO ${backendUser};
GRANT ALL ON SCHEMA ${dashboardSchema} TO ${dashboardUser};

-- default privileges explicitly by SCHEMA
ALTER DEFAULT PRIVILEGES FOR ROLE ${updaterUser} IN SCHEMA ${backendSchema} 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${backendUser};

ALTER DEFAULT PRIVILEGES FOR ROLE ${updaterUser} IN SCHEMA ${dashboardSchema} 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${dashboardUser};
"@
        Set-Content -Path ${initSqlPath} -Value ${sqlContent} -Encoding UTF8

        docker run -d --name db-dev `
            -e "POSTGRES_USER=${env:DB_ADMIN_USER}" `
            -e "POSTGRES_PASSWORD=${env:DB_ADMIN_PASS}" `
            -e POSTGRES_DB=$env:DB_NAME `
            -p "${env:DB_PORT}:5432" `
            -v "${initSqlPath}:/docker-entrypoint-initdb.d/init-multi-user.sql:ro" `
            postgres:16-alpine | Out-Null
        if ($LASTEXITCODE -ne 0) { throw "failed to start temporary db-dev container" }

        $startedTemporaryDb = $true

        # Wait until Postgres is genuinely accepting TCP connections (max 15s)
        $ready = $false
        for ($i = 0; $i -lt 30; $i++) {
            docker exec db-dev pg_isready -U $env:DB_ADMIN_USER -d $env:DB_NAME 2>$null | Out-Null
            if ($LASTEXITCODE -eq 0) {
                $ready = $true
                break
            }
            Start-Sleep -Seconds 1
        }
        if (-not $ready) { throw "timed out waiting for temporary db-dev to become ready" }

        Remove-Item -Path $initSqlPath -Force
    }

    Write-Host "[1/2] running DB migration ($Direction)..." -ForegroundColor Cyan
    Push-Location services/migration
    try {
        go run ./cmd/app $Direction
        if ($LASTEXITCODE -ne 0) { throw "Database migration ($Direction) failed." }
    }
    finally {
        Pop-Location
    }

    Write-Host "[2/2] regenerating GORM models..." -ForegroundColor Cyan
    Push-Location services/backend
    try {
        go run ./cmd/gormgen
        if ($LASTEXITCODE -ne 0) { throw "GORM model generation failed" }
    }
    finally {
        Pop-Location
    }

    Write-Host "schema migrated ($Direction) and GORM models updated" -ForegroundColor Green
}
finally {
    if ($startedTemporaryDb) {
        Write-Host "removing temporary db-dev container..." -ForegroundColor Yellow
        docker stop db-dev | Out-Null
        docker rm db-dev | Out-Null
    }
}