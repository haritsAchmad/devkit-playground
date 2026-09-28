# Build devkittool.exe lalu salin ke tiap folder manifest di agent-tools/.
#
# Kenapa disalin, bukan dipakai bersama lewat path relatif: local-agent-playground
# mengekang executable manifest ke folder manifest itu sendiri, jadi tiap folder
# butuh salinannya sendiri (pola yang sama dengan database-tool/redis-tool).
# Salinan .exe diabaikan Git (*.exe di .gitignore).

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot\..

go build -trimpath -o devkittool.exe ./cmd/devkittool

$targets = Get-ChildItem agent-tools -Directory | ForEach-Object { Join-Path $_.FullName "devkittool.exe" }
foreach ($target in $targets) {
    Copy-Item devkittool.exe $target -Force
}

Write-Host "Build selesai, disalin ke:"
$targets | ForEach-Object { Write-Host "  $_" }
