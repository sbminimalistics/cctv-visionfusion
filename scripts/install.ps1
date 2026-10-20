# dynamic root path alignment
# $PSScriptRoot is always the folder where this script lives (e.g., /scripts)
$scriptDir = $PSScriptRoot
$projectRootDir = Split-Path -Path $scriptDir -Parent

# force the active execution context back out to the true project root
Set-Location -Path $projectRootDir
Write-Host "[install] workspace root aligned to: $projectRootDir" -ForegroundColor Gray

Write-Host "[install] TO-DO: implement dependency installation calls for each application in ${projectRootDir}\services" -ForegroundColor Gray
