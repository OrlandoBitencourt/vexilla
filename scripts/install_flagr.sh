
Write-Host "=== Instalador Flagr (GoFlagr) para Windows ===`n"

# 1. Verificar se Docker está instalado
Write-Host "Verificando Docker..."
if (!(Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Docker Desktop não encontrado."
    Write-Host "Instale o Docker Desktop antes de continuar: https://www.docker.com/products/docker-desktop/"
    exit 1
}

Write-Host "✔ Docker encontrado."

# 2. Criar pasta do projeto
$flagrPath = "$env:USERPROFILE\flagr"
if (!(Test-Path $flagrPath)) {
    New-Item -ItemType Directory -Path $flagrPath | Out-Null
    Write-Host "Criada pasta: $flagrPath"
} else {
    Write-Host "Pasta já existe: $flagrPath"
}

# 3. Gerar arquivo docker-compose.yml
$composeContent = @"
version: '3.7'

services:
  flagr:
    image: openflagr/flagr:latest
    container_name: flagr
    ports:
      - "18000:18000"
    environment:
      - FLAGR_DB_DRIVER=sqlite3
      - FLAGR_DB_DSN=/data/flagr.sqlite3
    volumes:
      - ./data:/data
"@

$composeFile = "$flagrPath\docker-compose.yml"
$composeContent | Out-File -Encoding UTF8 $composeFile

Write-Host "✔ Arquivo docker-compose.yml criado."

# 4. Subir Flagr
Write-Host "Iniciando Flagr com Docker..."
docker compose -f "$composeFile" up -d

Write-Host "`n✔ Flagr iniciado!"
Write-Host "Acesse o painel em: http://localhost:18000"
Write-Host "Para parar: docker compose -f $composeFile down"
