param (
    [Alias("backend-config")]
    [string]$BackendConfig
)

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

# translate regular env values to Vite specific
$env:VITE_BFF_HOST = $env:BFF_HOST
$env:VITE_BFF_PORT = $env:BFF_PORT
$env:VITE_BFF_API_PREFIX = $env:BFF_API_PREFIX

Write-Host "[dev] starting local infrastructure..." -ForegroundColor Cyan

Write-Host "[dev] setting up docker dev-network..." -ForegroundColor DarkCyan
if (-not (docker network ls -q -f name=dev-network)) {
    docker network create dev-network | Out-Null
}

Write-Host "[dev] spinning up SeaweedFS box..." -ForegroundColor Gray

if (-not (Test-Path -Path "${env:SEAWEED_HOST_PATH}")) {
    New-Item -ItemType Directory -Path "s${env:SEAWEED_HOST_PATH}"
}

# SeaweedFS includes a native, lightweight administration tool called 'wed' or 's3.config' 
# We can create a config file directly on Windows host path that SeaweedFS reads!
$s3ConfigPath = "${env:SEAWEED_HOST_PATH}\s3.json"
$s3ConfigContent = @{
    "identities" = @(
        # downloader
        @{
            "name"        = "downloader_service"
            "credentials" = @(
                @{
                    "accessKey" = "${env:DOWNLOADER_SEAWEED_USER}"
                    "secretKey" = "${env:DOWNLOADER_SEAWEED_PASS}"
                }
            )
            "actions"     = @("Write", "List", "Tagging")
            "buckets"     = @("${env:SEAWEED_BUCKET}")
        },
        # cv
        @{
            "name"        = "cv_service"
            "credentials" = @(
                @{
                    "accessKey" = "${env:CV_SEAWEED_USER}"
                    "secretKey" = "${env:CV_SEAWEED_PASS}"
                }
            )
            "actions"     = @("Read", "Tagging")
            "buckets"     = @("${env:SEAWEED_BUCKET}")
        },
        # dashboard
        @{
            "name"        = "dashboard_ui"
            "credentials" = @(
                @{
                    "accessKey" = "${env:DASHBOARD_SEAWEED_USER}"
                    "secretKey" = "${env:DASHBOARD_SEAWEED_PASS}"
                }
            )
            "actions"     = @("Read", "List") # No write or delete permissions allowed
            "buckets"     = @("${env:SEAWEED_BUCKET}")
        },
        # admin
        @{
            "name"        = "admin_user"
            "credentials" = @(
                @{
                    "accessKey" = "${env:SEAWEED_ADMIN_USER}"
                    "secretKey" = "${env:SEAWEED_ADMIN_PASS}"
                }
            )
            "actions"     = @("Admin", "Read", "Write", "List")
            # Leaving "buckets" out completely gives this user access to ALL buckets
        }
    )
} | ConvertTo-Json -Depth 5

Set-Content -Path $s3ConfigPath -Value $s3ConfigContent

docker run -d --name seaweedfs-dev `
  --network dev-network `
  -p ${env:SEAWEED_MASTER_PORT}:9333 `
  -p ${env:SEAWEED_VOLUME_PORT}:8080 `
  -p ${env:SEAWEED_API_PORT}:8333 `
  -v ${env:SEAWEED_HOST_PATH}:/data `
  -v "${projectRootDir}/seaweed:/conf" `
  chrislusf/seaweedfs:latest `
  server -dir=/data -s3 -s3.config=/data/s3.json -s3.allowEmptyFolder=true --lifecycle-configuration /conf/lifecycle.xml

# Wait a brief moment for SeaweedFS to initialize
Write-Host "[dev] Waiting for SeaweedFS to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 3

Write-Host "[dev] creating SeaweedFS credentials and bucket..." -ForegroundColor Gray

# Restart the container quickly just to pick the brand new credentials file
docker restart seaweedfs-dev | Out-Null

# create the bucket
Start-Sleep -Seconds 2

"s3.bucket.create -name=${env:SEAWEED_BUCKET}" | docker exec -i seaweedfs-dev weed shell

Write-Host "[dev] SeaweedFS setup complete! Port ${env:SEAWEED_API_PORT} for S3 API endpoint." -ForegroundColor Green

# launch rabbitmq core broker instance using env bindings
$rmqPort = if ($env:RMQ_PORT) { $env:RMQ_PORT } else { "5672" }
$dbPort  = if ($env:DB_PORT)  { $env:DB_PORT }  else { "5432" }

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
ALTER DEFAULT PRIVILEGES FOR ROLE ${updaterUser} IN SCHEMA ${backendSchema}
    GRANT USAGE, SELECT ON SEQUENCES TO ${backendUser};

ALTER DEFAULT PRIVILEGES FOR ROLE ${updaterUser} IN SCHEMA ${dashboardSchema} 
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${dashboardUser};
ALTER DEFAULT PRIVILEGES FOR ROLE ${updaterUser} IN SCHEMA ${dashboardSchema}
    GRANT USAGE, SELECT ON SEQUENCES TO ${dashboardUser};

-- set user searrch_path
ALTER ROLE ${backendUser} SET search_path TO ${backendSchema};
ALTER ROLE ${dashboardUser} SET search_path TO ${dashboardSchema};
"@
Set-Content -Path ${initSqlPath} -Value ${sqlContent} -Encoding UTF8

# launch postgres
Write-Host "[dev] starting db-dev postgresql engine with multi-tenant workspace..." -ForegroundColor Gray
docker run -d --name db-dev `
  -p "${dbPort}:5432" `
  -e "POSTGRES_DB=${dbName}" `
  -e "POSTGRES_USER=${env:DB_ADMIN_USER}" `
  -e "POSTGRES_PASSWORD=${env:DB_ADMIN_PASS}" `
  -v cctv-vf-dev-pgdata:/var/lib/postgresql/data `
  -v "${initSqlPath}:/docker-entrypoint-initdb.d/init-multi-user.sql:ro" `
  postgres:16-alpine

Write-Host "[dev] waiting for database initial setup script execution to finish..." -ForegroundColor Yellow
$ready = $false
for ($i = 0; $i -lt 30; $i++) {
    docker exec db-dev pg_isready -U $env:DB_ADMIN_USER -d $env:DB_NAME 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        break
    }
    Start-Sleep -Seconds 2
}

if (-not $ready) {
    throw "Timed out waiting for db-dev to become ready."
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

Write-Host "[dev] rabbitmq is operational. applying user accounts..." -ForegroundColor Yellow

# configure backend application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:BACKEND_RMQ_USER}" "${env:BACKEND_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:BACKEND_RMQ_USER}" ".*" ".*" ".*" | Out-Null

# configure dashboard application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:DASHBOARD_RMQ_USER}" "${env:DASHBOARD_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:DASHBOARD_RMQ_USER}" "^$" "^$" ".*" | Out-Null

# configure downloader application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:DOWNLOADER_RMQ_USER}" "${env:DOWNLOADER_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:DOWNLOADER_RMQ_USER}" "^$" ".*" ".*" | Out-Null

# configure cv application rmq user
docker exec rmq-dev rabbitmqctl add_user "${env:CV_RMQ_USER}" "${env:CV_RMQ_PASS}" | Out-Null
docker exec rmq-dev rabbitmqctl set_permissions -p "${env:RMQ_VHOST}" "${env:CV_RMQ_USER}" "^$" ".*" ".*" | Out-Null

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

# add BACKEND_CONFIG_PATH environment variable
if (-not [string]::IsNullOrWhiteSpace($BackendConfig)) {
    $env:BACKEND_CONFIG_PATH = $BackendConfig
}

$ErrorActionPreference = "Stop"
$RepoRoot = $projectRootDir

$Services = @(
    @{
        Name  = "BE"
        Color = "Cyan"
        Dir   = Join-Path $RepoRoot "services/backend"
        File  = "go"
        Args  = @("tool", "air")
    },
    @{
        Name  = "DWNLDR"
        Color = "Green"
        Dir   = Join-Path $RepoRoot "services/downloader"
        File  = "go"
        Args  = @("tool", "air")
    },
    @{
        Name  = "DASH"
        Color = "Magenta"
        Dir   = Join-Path $RepoRoot "services/dashboard"
        File  = Join-Path $RepoRoot "services/dashboard/node_modules/.bin/tsx.cmd"
        Args  = @("watch", "--clear-screen=false", "src/app.ts")
    },
    @{
        Name  = "VITE"
        Color = "DarkGray"
        Dir   = Join-Path $RepoRoot "services/dashboard/web-ui"
        File  = "node.exe"
        Args  = @(".\node_modules\vite\bin\vite.js")
    },
    @{
        Name  = "CV"
        Color = "Yellow"
        Dir   = Join-Path $RepoRoot "services/cv"
        File  = Join-Path $RepoRoot "services/cv/.venv/Scripts/watchmedo.exe"
        Args  = @(
            "auto-restart",
            "--directory=.",
            "--pattern=*.py",
            "--ignore-directories",
            "--recursive",
            "--",
            ".\.venv\Scripts\python.exe",
            "-u",
            "main.py"
        )
    }
)

$Jobs = @()
$Pids = @{}

function Stop-ProcessTree {
    param([int]$ParentPid)

    $ErrorActionPreference = "SilentlyContinue"
    $children = Get-CimInstance Win32_Process -Filter "ParentProcessId = $ParentPid"
    foreach ($child in $children) {
        if ($child.ProcessId) {
            Stop-ProcessTree -ParentPid ([int]$child.ProcessId)
        }
    }
    if (Get-Process -Id $ParentPid) {
        Stop-Process -Id $ParentPid -Force
    }
}

function Stop-AllServices {
    # Never let a single missing PID abort the rest of shutdown or docker cleanup
    $ErrorActionPreference = "Continue"
    Write-Host "`n[DEV] Shutting down all services and child process trees..." -ForegroundColor Red

    # 1. Kill only tracked PIDs that are still alive
    foreach ($name in @($Pids.Keys)) {
        $procId = $Pids[$name]
        if (-not $procId) { continue }

        $alive = Get-Process -Id $procId -ErrorAction SilentlyContinue
        if (-not $alive) {
            Write-Host "  Skipped $name (PID $procId already exited)" -ForegroundColor DarkGray
            continue
        }

        Stop-ProcessTree -ParentPid $procId
        Write-Host "  Stopped $name (PID $procId)" -ForegroundColor DarkYellow
    }

    # 2. Safety-net sweep for any detached child processes inside $RepoRoot\services\
    $resolvedRoot = Resolve-Path $RepoRoot -ErrorAction SilentlyContinue
    if ($resolvedRoot) {
        $escapedRoot = [Regex]::Escape($resolvedRoot.Path)
        $strays = Get-CimInstance Win32_Process -ErrorAction SilentlyContinue | Where-Object {
            $_.ProcessId -ne $PID -and
            $_.CommandLine -and
            $_.CommandLine -match "$escapedRoot[\\/]services[\\/]"
        }
        foreach ($stray in $strays) {
            Stop-Process -Id ([int]$stray.ProcessId) -Force -ErrorAction SilentlyContinue
            Write-Host "  Cleaned detached child $($stray.Name) (PID $($stray.ProcessId))" -ForegroundColor DarkYellow
        }
    }

    # 3. Terminate background Start-Job host processes directly so pending ReadLineAsync()
    #    tasks don't hang Stop-Job, then remove the job objects cleanly
    $jobHostPids = Get-CimInstance Win32_Process -Filter "ParentProcessId = $PID" -ErrorAction SilentlyContinue |
        Select-Object -ExpandProperty ProcessId
    foreach ($jpid in $jobHostPids) {
        Stop-Process -Id ([int]$jpid) -Force -ErrorAction SilentlyContinue
    }

    foreach ($item in $Jobs) {
        Remove-Job -Job $item.Job -Force -ErrorAction SilentlyContinue
    }
}

try {
    foreach ($svc in $Services) {
        $job = Start-Job -Name $svc.Name -ScriptBlock {
            param($Dir, $File, $ProcArgs)

            $psi = New-Object System.Diagnostics.ProcessStartInfo
            $psi.FileName = $File
            $psi.WorkingDirectory = $Dir
            $psi.UseShellExecute = $false
            $psi.RedirectStandardOutput = $true
            $psi.RedirectStandardError = $true
            $psi.CreateNoWindow = $true
            $psi.StandardOutputEncoding = [System.Text.Encoding]::UTF8
            $psi.StandardErrorEncoding = [System.Text.Encoding]::UTF8

            # copy all parent script environment variables into the child process
            foreach ($key in $EnvMap.Keys) {
                $psi.EnvironmentVariables[$key] = $EnvMap[$key]
            }

            $psi.Arguments = ($ProcArgs | ForEach-Object {
                if ($_ -match '\s') { '"' + $_ + '"' } else { $_ }
            }) -join ' '

            $psi.RedirectStandardInput = $true
            $psi.EnvironmentVariables["PYTHONUNBUFFERED"] = "1"
            $psi.EnvironmentVariables["FORCE_COLOR"] = "1"

            $proc = New-Object System.Diagnostics.Process
            $proc.StartInfo = $psi
            [void]$proc.Start()

            $proc.StandardInput.Close()

            Start-Sleep -Milliseconds 50
            if (-not $proc.HasExited) {
                Write-Output "__PID__:$($proc.Id)"
            }

            $outTask = $proc.StandardOutput.ReadLineAsync()
            $errTask = $proc.StandardError.ReadLineAsync()

            while (-not $proc.HasExited -or -not $outTask.IsCompleted -or -not $errTask.IsCompleted) {
                if ($outTask.IsCompleted) {
                    $line = $outTask.Result
                    if ($null -ne $line) {
                        Write-Output $line
                        $outTask = $proc.StandardOutput.ReadLineAsync()
                    }
                }
                if ($errTask.IsCompleted) {
                    $errLine = $errTask.Result
                    if ($null -ne $errLine) {
                        Write-Output " $errLine"
                        $errTask = $proc.StandardError.ReadLineAsync()
                    }
                }
                Start-Sleep -Milliseconds 15
            }
        } -ArgumentList $svc.Dir, $svc.File, $svc.Args

        $Jobs += @{ Job = $job; Name = $svc.Name; Color = $svc.Color }
    }

    # Intercept Ctrl+C cleanly so PowerShell doesn't abort the finally block mid-cleanup
    [Console]::TreatControlCAsInput = $true
    Write-Host "[DEV] All services started. Press Ctrl+C to stop.`n" -ForegroundColor White

    while ($true) {
        # Check for Ctrl+C keypress
        if ([Console]::KeyAvailable) {
            $key = [Console]::ReadKey($true)
            if (($key.Modifiers -band [ConsoleModifiers]::Control) -and ($key.Key -eq [ConsoleKey]::C)) {
                break
            }
        }

        $anyRunning = $false
        foreach ($item in $Jobs) {
            if ($item.Job.State -eq "Running") { $anyRunning = $true }

            $lines = Receive-Job -Job $item.Job -ErrorVariable jobErr -ErrorAction SilentlyContinue
            foreach ($err in $jobErr) {
                Write-Host "[$($item.Name) FAIL] " -ForegroundColor Red -NoNewline
                Write-Host $err.Exception.Message
            }

            foreach ($line in $lines) {
                if ($line -is [string] -and $line.StartsWith("__PID__:")) {
                    $Pids[$item.Name] = [int]$line.Substring(8)
                    Write-Host [int]$line.Substring(8)
                    continue
                }
                Write-Host "[$($item.Name)] " -ForegroundColor $item.Color -NoNewline
                Write-Host $line
            }
        }

        if (-not $anyRunning) { break }
        Start-Sleep -Milliseconds 50
    }
}
finally {
    [Console]::TreatControlCAsInput = $false

    Stop-AllServices

    Write-Host "`n[dev] cleaning up local development containers..." -ForegroundColor Yellow
    docker stop rmq-dev db-dev seaweedfs-dev 2>$null | Out-Null
    docker rm rmq-dev db-dev seaweedfs-dev 2>$null | Out-Null
    Remove-Item Env:BACKEND_CONFIG_PATH -ErrorAction SilentlyContinue
    Write-Host "[dev] backend sandbox containers removed successfully." -ForegroundColor Green

    exit 0
}



# block script exit loop to continuously monitor the execution boundary state
while ($true) { 
    Start-Sleep -Seconds 1 
}
