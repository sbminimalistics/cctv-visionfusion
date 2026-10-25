# dynamic root path alignment
# $PSScriptRoot is always the folder where this script lives (e.g., /scripts)
$scriptDir = $PSScriptRoot
$projectRootDir = Split-Path -Path $scriptDir -Parent

# force the active execution context back out to the true project root
Set-Location -Path $projectRootDir
Write-Host "[dev] workspace root aligned to: $projectRootDir" -ForegroundColor Gray

# load and parse environment variable file mappings natively
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

Write-Host "[dev] starting local infrastructure..." -ForegroundColor Cyan

# launch rabbitmq core broker instance using env bindings
$rmqPort = if ($env:RMQ_PORT) { $env:RMQ_PORT } else { "5672" }
$dbPort  = if ($env:DB_PORT)  { $env:DB_PORT }  else { "5432" }

Write-Host "[dev] setting up docker dev-network..." -ForegroundColor DarkCyan
if (-not (docker network ls -q -f name=dev-network)) {
    docker network create dev-network | Out-Null
}

Write-Host "[dev] spinning up rabbitmq-dev broker box..." -ForegroundColor Gray
docker run -d --name rmq-dev `
  --network dev-network `
  -p "${rmqPort}:5672" `
  -p 15672:15672 `
  -v "${PSScriptRoot}/../rabbitmq/definitions.json:/etc/rabbitmq/definitions.json:ro" `
  -v "${PSScriptRoot}/../rabbitmq/rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro" `
  -e "HOME=/tmp" `
  -e "RABBITMQ_ERLANG_COOKIE=local_secret_development_cookie_token_991" `
  rabbitmq:4-management-alpine

# launch postgresql database instance
$dbName = if (${env:DB_NAME}) { $env:DB_NAME } else {"cctv-vf_database"}
$dbPort = if (${env:DB_PORT}) { $env:DB_PORT } else {"5432"}
$backendUser = $env:BACKEND_DB_USER
$backendPass = $env:BACKEND_DB_PASS
$backendSchema = $env:BACKEND_DB_SCHEMA
$dashboardUser = $env:DASHBOARD_DB_USER
$dashboardPass = $env:DASHBOARD_DB_PASS
$dashboardSchema = $env:DASHBOARD_DB_SCHEMA

$initSqlPath = "${PSScriptRoot}/init-multi-user.sql"
$sqlContent = @"
-- Create application-specific roles/users
CREATE USER ${backendUser} WITH PASSWORD '${backendPass}';
CREATE USER ${dashboardUser} WITH PASSWORD '${dashboardPass}';

-- provision schemas
CREATE SCHEMA ${backendSchema} AUTHORIZATION ${backendUser};
CREATE SCHEMA ${dashboardSchema} AUTHORIZATION ${dashboardUser};

-- grant targeted schema permissions to maintain complete application safety boundaries
GRANT ALL ON SCHEMA ${backendSchema} TO ${backendUser};
GRANT ALL ON SCHEMA ${dashboardSchema} TO ${dashboardUser};
"@
Set-Content -Path ${initSqlPath} -Value ${sqlContent} -Encoding UTF8

# launch postgres
Write-Host "[dev] starting db-dev postgresql engine with multi-tenant workspace..." -ForegroundColor Gray
docker run -d --name db-dev `
  -p "${dbPort}:5432" `
  -e "POSTGRES_DB=${dbName}" `
  -e "POSTGRES_USER=admin" `
  -e "POSTGRES_PASSWORD=adminpass123" `
  -v "${initSqlPath}:/docker-entrypoint-initdb.d/init-multi-user.sql:ro" `
  postgres:16-alpine

Write-Host "[dev] waiting for database initial setup script execution to finish..." -ForegroundColor Yellow
while ($true) {
    $logs = docker logs db-dev 2>&1
    if ($logs -match "PostgreSQL init process complete; ready for start up") {
        break
    }
    Start-Sleep -Seconds 1
}

# remove temp sql file now that the init container sequence safely completed
Remove-Item -Path $initSqlPath -Force
Write-Host "[dev] database is completely ready, script cleaned up successfully!" 

# block until rabbitmq internal management subsystem finishes booting up
Write-Host "[dev] waiting for rabbitmq management subsystem to initialize..." -ForegroundColor Yellow
while ($true) {
    # wait for command exit code 0
    docker exec rmq-dev rabbitmqctl await_startup >$null 2>$null
    if ($LastExitCode -eq 0) {
        break
    }
    Start-Sleep -Seconds 1
}

Start-Sleep -Seconds 2

Write-Host "[dev] rabbitmq is operational. applying user accounts..." -ForegroundColor Cyan

# configure backend application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:BACKEND_RMQ_USER}" "${env:BACKEND_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:BACKEND_RMQ_USER}" ".*" ".*" ".*" | Out-Null

# configure dashboard application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:DASHBOARD_RMQ_USER}" "${env:DASHBOARD_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:DASHBOARD_RMQ_USER}" "^$" "^$" ".*" | Out-Null

# configure downloader application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:DOWNLOADER_RMQ_USER}" "${env:DOWNLOADER_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:DOWNLOADER_RMQ_USER}" "^$" ".*" ".*" | Out-Null

Write-Host "[dev] infrastructure provisioning complete. ports are stable." -ForegroundColor Green

# dynamic sleep buffer giving database and messaging daemons a window to initialize listeners
Write-Host "[dev] waiting 2 seconds for communication ports to stabilize..." -ForegroundColor Yellow
Start-Sleep -Seconds 2

# Write-Host "[dev] building python image..." -ForegroundColor Blue
# docker build -t opencv5-image "${PSScriptRoot}/../services/cv"

# Write-Host "[dev] spinning up python container..." -ForegroundColor Blue
# docker run --name opencv5-worker `
# --add-host=host.docker.internal:host-gateway `
#   --network dev-network `
#   -e RABBITMQ_HOST=${env:RMQ_HOST} `
#   -e RABBITMQ_PORT=${env:RMQ_PORT} `
#   -e RABBITMQ_VHOST=${env:RMQ_VHOST} `
#   -e RABBITMQ_QUEUE=${env:} `
#   -e DOWNLOADER_RMQ_USER=${env:DOWNLOADER_RMQ_USER} `
#   -e DOWNLOADER_RMQ_PASS=${env:DOWNLOADER_RMQ_PASS} `
#   -d `
#   opencv5-image | Out-Null

Write-Host "[dev] launching workspace components natively..." -ForegroundColor Yellow

# set execution configurations
$concurrentArgs = @(
    "concurrently",
    "--names", "BE,DASH,VITE,CV",
    "--prefix-colors", "cyan,green,magenta",
    "cd services/backend && go tool air",
    "cd services/dashboard && npm run dev",
    "cd services/dashboard/web-ui && npm run dev",
    "cd services/cv/ && watchmedo auto-restart --directory=. --pattern=*.py --recursive -- python main.py"
)

try {
    # execute the application mesh inside the active window
    & npx $concurrentArgs
}
finally {
    # this block is guaranteed to execute by the OS, even during a Ctrl+C force-abort
    Write-Host "`n[dev] cleaning up local development containers..." -ForegroundColor Yellow
    docker stop rmq-dev db-dev | Out-Null
    docker rm rmq-dev db-dev | Out-Null
    Write-Host "[dev] backend sandbox containers removed successfully." -ForegroundColor Green
}

# block script exit loop to continuously monitor the execution boundary state
while ($true) { 
    Start-Sleep -Seconds 1 
}
