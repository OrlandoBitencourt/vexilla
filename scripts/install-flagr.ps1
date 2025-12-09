# Instalador Flagr para Windows (sem Docker)

Write-Host "=== Instalando Flagr no Windows (sem Docker) ==="

$flagrVersion = "1.1.18"
$installPath = "$env:USERPROFILE\flagr"
$downloadUrl = "https://github.com/openflagr/flagr/releases/download/v$flagrVersion/flagr_${flagrVersion}_windows_amd64.zip"
$zipFile = "$installPath\flagr.zip"

# Criar pasta
if (!(Test-Path $installPath)) {
    New-Item -ItemType Directory -Path $installPath | Out-Null
}

# Baixar
Write-Host "Baixando Flagr v$flagrVersion..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile -UseBasicParsing

# Extrair
Write-Host "Extraindo..."
Expand-Archive -Path $zipFile -DestinationPath $installPath -Force
Remove-Item $zipFile

# Criar config.yml
$configFile = "$installPath\config.yml"

$configContent = @'
db:
  driver: sqlite3
  dsn: "flagr.sqlite3"

log:
  level: info

ui:
  appName: "Flagr"
'@

Set-Content -Path $configFile -Value $configContent -Encoding UTF8

# Criar start-flagr.bat (sem Here-String!)
$batFile = "$installPath\start-flagr.bat"

$batLines = @(
    "@echo off"
    "cd $installPath"
    "echo Iniciando Flagr..."
    "flagr.exe -config config.yml"
    "pause"
)

Set-Content -Path $batFile -Value $batLines -Encoding ASCII

Write-Host "Instalação concluída!"
Write-Host "Execute: $batFile"
Write-Host "Acesse: http://localhost:18000"
