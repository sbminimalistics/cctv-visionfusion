# dynamic root path alignment
# $PSScriptRoot is always the folder where this script lives (e.g., /scripts)
$scriptDir = $PSScriptRoot
$projectRootDir = Split-Path -Path $scriptDir -Parent

Write-Host "[install-dev] workspace root aligned to: $projectRootDir" -ForegroundColor Gray

Write-Host "[install-dev] installing python dependencies in /services/cv" -ForegroundColor Yellow
Set-Location -Path ${projectRootDir}\services\cv
python -m venv .venv
.venv\Scripts\activate
py -m pip install -r requirements.txt -r requirements-dev.txt

Write-Host "[install-dev] installing node dependencies in /services/dashboard" -ForegroundColor Yellow
Set-Location -Path ${projectRootDir}\services\dashboard
npm ci
Write-Host "[install-dev] installing node dependencies in /services/dashboard/web-ui" -ForegroundColor Yellow
Set-Location -Path ${projectRootDir}\services\dashboard\web-ui
npm ci

Write-Host "[install-dev] dev dependencies successfully installed" -ForegroundColor Yellow
