# dynamic root path alignment
# $PSScriptRoot is always the folder where this script lives (e.g., /scripts)
$scriptDir = $PSScriptRoot
$projectRootDir = Split-Path -Path $scriptDir -Parent
$projectRootDirForward = $projectRootDir.Replace('\', '/').Replace("C:", "c:")

# force the active execution context back out to the true project root
Set-Location -Path $projectRootDir
Write-Host "[dev] workspace root aligned to: $projectRootDir" -ForegroundColor Gray

# Create the output directory folders if they do not exist yet
if (-not (Test-Path -Path "services/vms-backend-service/proto")) {
    New-Item -ItemType Directory -Path "services/vms-backend-service/proto"
}
if (-not (Test-Path -Path "services/vms-dashboard-service/proto")) {
    New-Item -ItemType Directory -Path "services/vms-dashboard-service/proto"
}

$IncludeCacheDir = Join-Path WorkspaceRoot ".proto_include\google\protobuf"
if (-not (Test-Path -Path \$IncludeCacheDir)) {
    New-Item -ItemType Directory -Force -Path \$IncludeCacheDir | Out-Null
    Write-Host "[+] Fetching missing Google system timestamp dependency components..." -ForegroundColor Yellow
    $BaseUrl = "https://githubusercontent.com"
    Invoke-WebRequest -Uri "BaseUrl/timestamp.proto" -OutFile (Join-Path IncludeCacheDir "timestamp.proto")
}

# Execute unified compilation
protoc `
--proto_path=proto `
--go_out=services/vms-backend-service/proto `
--plugin=protoc-gen-ts_proto="${projectRootDirForward}/services/vms-dashboard-service/node_modules/.bin/protoc-gen-ts_proto.cmd" `
--ts_proto_out="${projectRootDirForward}/services/vms-dashboard-service/proto" `
--ts_proto_opt=env=node,esModuleInterop=true `
proto/events.proto
