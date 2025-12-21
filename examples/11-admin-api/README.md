# Admin API Examples

Exemplos práticos e executáveis da Admin API do Vexilla.

## Estrutura

```
11-admin-api/
├── 01-basic-setup/          # Setup básico com Admin API
├── 02-health-check/         # Health checks e monitoramento
├── 03-metrics/              # Consulta e exportação de métricas
├── 04-cache-invalidation/   # Invalidação de cache
├── 05-manual-refresh/       # Refresh manual de flags
├── 06-monitoring-dashboard/ # Dashboard de monitoramento
└── README.md
```

## Como Executar

Cada exemplo é independente e pode ser executado diretamente:

```bash
# Exemplo 1: Setup básico
cd 01-basic-setup
go run main.go

# Exemplo 2: Health check
cd 02-health-check
go run main.go

# E assim por diante...
```

## Pré-requisitos

1. **Flagr rodando localmente**:
   ```bash
   docker run -p 18000:18000 checkr/flagr:latest
   ```

2. **Go 1.22+** instalado

## Ordem Recomendada

1. `01-basic-setup` - Entender como habilitar a Admin API
2. `02-health-check` - Verificar saúde do serviço
3. `03-metrics` - Monitorar métricas de performance
4. `04-cache-invalidation` - Gerenciar cache
5. `05-manual-refresh` - Atualizar flags manualmente
6. `06-monitoring-dashboard` - Visualizar métricas em tempo real

## Endpoints Disponíveis

Todos os exemplos usam os seguintes endpoints da Admin API:

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/health` | GET | Health check |
| `/admin/stats` | GET | Métricas do cache |
| `/admin/invalidate` | POST | Invalidar flag específica |
| `/admin/invalidate-all` | POST | Invalidar todas as flags |
| `/admin/refresh` | POST | Forçar refresh |

## Testando com curl

```bash
# Health check
curl http://localhost:19000/health

# Métricas
curl http://localhost:19000/admin/stats | jq

# Invalidar flag
curl -X POST http://localhost:19000/admin/invalidate \
  -H "Content-Type: application/json" \
  -d '{"flag_key": "my-flag"}'

# Refresh
curl -X POST http://localhost:19000/admin/refresh
```
