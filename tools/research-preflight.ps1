param(
    [int]$ActiveResearchLines = 80
)

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$active = Join-Path $root 'docs\research\ACTIVE_RESEARCH.md'

Write-Host '=== MOOX research preflight ==='
Write-Host ('Repo: ' + $root)
Write-Host ''

Push-Location $root
try {
    Write-Host '--- Git status ---'
    git status --short --branch
    if ($LASTEXITCODE -ne 0) {
        throw 'git status failed'
    }

    Write-Host ''
    Write-Host '--- Recent checkpoints ---'
    git log --oneline -6
    if ($LASTEXITCODE -ne 0) {
        throw 'git log failed'
    }

    Write-Host ''
    Write-Host '--- Active research ---'
    if (-not (Test-Path $active)) {
        throw "Missing $active"
    }
    Get-Content $active -TotalCount $ActiveResearchLines

    Write-Host ''
    Write-Host '--- Preflight reminder ---'
    Write-Host 'Proceed only if the planned next action directly advances ACTIVE_RESEARCH.md.'
    Write-Host 'Do not modify/stage unrelated dirty files.'
    Write-Host 'Stop an investigation path after 10 search/inspection actions without materialized progress.'
}
finally {
    Pop-Location
}
