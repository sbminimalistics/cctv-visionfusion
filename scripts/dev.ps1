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

Write-Host "[dev] spinning up rabbitmq-dev broker box..." -ForegroundColor Gray
docker run -d --name rmq-dev `
  -p "${rmqPort}:5672" `
  -p 15672:15672 `
  -v "${PSScriptRoot}/../rabbitmq/definitions.json:/etc/rabbitmq/definitions.json:ro" `
  -v "${PSScriptRoot}/../rabbitmq/rabbitmq.conf:/etc/rabbitmq/rabbitmq.conf:ro" `
  -e "HOME=/tmp" `
  -e "RABBITMQ_ERLANG_COOKIE=local_secret_development_cookie_token_991" `
  rabbitmq:4-management-alpine



# launch postgresql database instance
$dbName = if (${env:DB_NAME}) { $env:DB_NAME } else {"vms_database"}
$dbPort = if (${env:DB_PORT}) { $env:DB_PORT } else {"5432"}
$producerUser = $env:DB_PRODUCER_USER
$producerPass = $env:DB_PRODUCER_PASSWORD
$producerSchema = $env:DB_PRODUCER_SCHEMA
$consumerUser = $env:DB_CONSUMER_USER
$consumerPass = $env:DB_CONSUMER_PASSWORD
$consumerSchema = $env:DB_CONSUMER_SCHEMA

$initSqlPath = "${PSScriptRoot}/init-multi-user.sql"
$sqlContent = @"
-- Create application-specific roles/users
CREATE USER ${producerUser} WITH PASSWORD '${producerPass}';
CREATE USER ${consumerUser} WITH PASSWORD '${consumerPass}';

-- provision schemas
CREATE SCHEMA ${producerSchema} AUTHORIZATION ${producerUser};
CREATE SCHEMA ${consumerSchema} AUTHORIZATION ${consumerUser};

-- grant targeted schema permissions to maintain complete application safety boundaries
GRANT ALL ON SCHEMA ${producerSchema} TO ${producerUser};
GRANT ALL ON SCHEMA ${consumerSchema} TO ${consumerUser};
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

# configure the go application profile boundary rules
docker exec rmq-dev rabbitmqctl add_user "${env:RMQ_PRODUCER_USER}" "${env:RMQ_PRODUCER_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:RMQ_PRODUCER_USER}" ".*" ".*" ".*" | Out-Null

# configure the node application profile boundary rules
docker exec rmq-dev rabbitmqctl add_user "${env:RMQ_CONSUMER_USER}" "${env:RMQ_CONSUMER_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:RMQ_CONSUMER_USER}" "^$" "^$" ".*" | Out-Null

Write-Host "[dev] infrastructure provisioning complete. ports are stable." -ForegroundColor Green

# dynamic sleep buffer giving database and messaging daemons a window to initialize listeners
Write-Host "[dev] waiting 5 seconds for communication ports to stabilize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

Write-Host "[dev] launching workspace components natively..." -ForegroundColor Green

# set execution configurations
$concurrentArgs = @(
    "concurrently",
    "--names", "GO,NODE,VITE",
    "--prefix-colors", "cyan,green,magenta",
    "cd vms-backend-service && go tool air",
    "cd vms-ui-service && npm run dev",
    "cd vms-ui-service/web-ui && npm run dev"
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
