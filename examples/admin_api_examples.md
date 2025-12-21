# Admin API - Exemplos Práticos

Este guia mostra como usar a API Admin do Vexilla para monitoramento, gerenciamento e operações em produção.

## Índice

- [Configuração Inicial](#configuração-inicial)
- [Health Check](#health-check)
- [Métricas e Monitoramento](#métricas-e-monitoramento)
- [Invalidação de Cache](#invalidação-de-cache)
- [Refresh Manual](#refresh-manual)
- [Integração com CI/CD](#integração-com-cicd)
- [Kubernetes Health Probes](#kubernetes-health-probes)
- [Dashboard de Monitoramento](#dashboard-de-monitoramento)

---

## Configuração Inicial

### Setup Básico

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/OrlandoBitencourt/vexilla"
)

func main() {
    client, err := vexilla.New(
        // Conexão com Flagr
        vexilla.WithFlagrEndpoint("http://localhost:18000"),
        vexilla.WithRefreshInterval(5 * time.Minute),

        // Habilitar Admin API
        vexilla.WithAdminServer(vexilla.AdminConfig{
            Port: 19000,
        }),

        // Opcional: Webhook para invalidação em tempo real
        vexilla.WithWebhookInvalidation(vexilla.WebhookConfig{
            Port:   18001,
            Secret: "your-webhook-secret",
        }),
    )
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := client.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Stop()

    log.Println("Admin API disponível em http://localhost:19000")
    log.Println("Webhook disponível em http://localhost:18001/webhook")

    // Sua aplicação continua rodando...
    select {}
}
```

### Setup em Produção com Docker

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o vexilla-app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/vexilla-app .

# Expor porta da Admin API
EXPOSE 19000
# Expor porta do Webhook
EXPOSE 18001

CMD ["./vexilla-app"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  vexilla-app:
    build: .
    ports:
      - "8080:8080"   # Sua aplicação
      - "19000:19000" # Admin API
      - "18001:18001" # Webhook
    environment:
      - FLAGR_ENDPOINT=http://flagr:18000
      - WEBHOOK_SECRET=${WEBHOOK_SECRET}
    depends_on:
      - flagr
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:19000/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  flagr:
    image: checkr/flagr:latest
    ports:
      - "18000:18000"
```

---

## Health Check

### Verificar Saúde do Serviço

```bash
# Health check básico
curl http://localhost:19000/health

# Resposta:
{
  "status": "healthy",
  "timestamp": "2025-12-21T10:30:00Z"
}
```

### Script de Health Check

```bash
#!/bin/bash
# check_health.sh

ADMIN_URL="http://localhost:19000"

response=$(curl -s -w "\n%{http_code}" "${ADMIN_URL}/health")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
    echo "✅ Vexilla está saudável"
    echo "$body" | jq '.'
    exit 0
else
    echo "❌ Vexilla não está respondendo"
    echo "HTTP Status: $http_code"
    exit 1
fi
```

### Uso em Go

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type HealthResponse struct {
    Status    string `json:"status"`
    Timestamp string `json:"timestamp"`
}

func checkHealth(adminURL string) error {
    client := &http.Client{Timeout: 5 * time.Second}

    resp, err := client.Get(adminURL + "/health")
    if err != nil {
        return fmt.Errorf("falha ao conectar: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status não-saudável: %d", resp.StatusCode)
    }

    var health HealthResponse
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        return fmt.Errorf("falha ao decodificar resposta: %w", err)
    }

    fmt.Printf("✅ Status: %s (às %s)\n", health.Status, health.Timestamp)
    return nil
}

func main() {
    if err := checkHealth("http://localhost:19000"); err != nil {
        fmt.Printf("❌ Health check falhou: %v\n", err)
    }
}
```

---

## Métricas e Monitoramento

### Consultar Estatísticas do Cache

```bash
# Obter todas as métricas
curl http://localhost:19000/admin/stats | jq

# Resposta:
{
  "storage": {
    "keys_added": 150,
    "keys_evicted": 12,
    "hit_ratio": 0.98
  },
  "last_refresh": "2025-12-21T10:25:00Z",
  "consecutive_fails": 0,
  "circuit_open": false
}
```

### Script de Monitoramento Contínuo

```bash
#!/bin/bash
# monitor_stats.sh - Monitora métricas a cada 30 segundos

ADMIN_URL="http://localhost:19000"
INTERVAL=30

echo "🔍 Monitorando Vexilla a cada ${INTERVAL}s (Ctrl+C para parar)"
echo "─────────────────────────────────────────────────────"

while true; do
    clear
    date
    echo ""

    # Obter métricas
    stats=$(curl -s "${ADMIN_URL}/admin/stats")

    # Extrair valores importantes
    hit_ratio=$(echo "$stats" | jq -r '.storage.hit_ratio * 100')
    keys_added=$(echo "$stats" | jq -r '.storage.keys_added')
    keys_evicted=$(echo "$stats" | jq -r '.storage.keys_evicted')
    last_refresh=$(echo "$stats" | jq -r '.last_refresh')
    circuit_open=$(echo "$stats" | jq -r '.circuit_open')

    echo "📊 Cache Performance:"
    echo "   Hit Ratio: ${hit_ratio}%"
    echo "   Keys Added: ${keys_added}"
    echo "   Keys Evicted: ${keys_evicted}"
    echo ""
    echo "🔄 Refresh:"
    echo "   Last Refresh: ${last_refresh}"
    echo ""
    echo "⚡ Circuit Breaker:"
    if [ "$circuit_open" = "true" ]; then
        echo "   Status: 🔴 ABERTO (Flagr indisponível)"
    else
        echo "   Status: 🟢 FECHADO (Funcionando)"
    fi

    sleep $INTERVAL
done
```

### Alertas Baseados em Métricas

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type CacheMetrics struct {
    Storage struct {
        KeysAdded   int     `json:"keys_added"`
        KeysEvicted int     `json:"keys_evicted"`
        HitRatio    float64 `json:"hit_ratio"`
    } `json:"storage"`
    LastRefresh      string `json:"last_refresh"`
    ConsecutiveFails int    `json:"consecutive_fails"`
    CircuitOpen      bool   `json:"circuit_open"`
}

func monitorMetrics(adminURL string) {
    client := &http.Client{Timeout: 5 * time.Second}

    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        resp, err := client.Get(adminURL + "/admin/stats")
        if err != nil {
            fmt.Printf("❌ Erro ao buscar métricas: %v\n", err)
            continue
        }

        var metrics CacheMetrics
        if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
            resp.Body.Close()
            fmt.Printf("❌ Erro ao decodificar métricas: %v\n", err)
            continue
        }
        resp.Body.Close()

        // Alertas
        if metrics.HitRatio < 0.8 {
            fmt.Printf("⚠️  ALERTA: Hit ratio baixo: %.2f%% (esperado >80%%)\n",
                metrics.HitRatio*100)
        }

        if metrics.CircuitOpen {
            fmt.Printf("🔴 ALERTA CRÍTICO: Circuit breaker aberto! Flagr indisponível\n")
        }

        if metrics.ConsecutiveFails > 3 {
            fmt.Printf("⚠️  ALERTA: %d falhas consecutivas de refresh\n",
                metrics.ConsecutiveFails)
        }

        // Log normal
        fmt.Printf("✅ Métricas OK - Hit Ratio: %.2f%%, Circuit: %v\n",
            metrics.HitRatio*100, !metrics.CircuitOpen)
    }
}

func main() {
    monitorMetrics("http://localhost:19000")
}
```

---

## Invalidação de Cache

### Invalidar Flag Específica

```bash
# Invalidar uma flag específica
curl -X POST http://localhost:19000/admin/invalidate \
  -H "Content-Type: application/json" \
  -d '{"flag_key": "new-feature"}'

# Resposta:
{
  "status": "success",
  "message": "Flag 'new-feature' invalidated"
}
```

### Invalidar Todas as Flags

```bash
# Limpar todo o cache
curl -X POST http://localhost:19000/admin/invalidate-all

# Resposta:
{
  "status": "success",
  "message": "All flags invalidated"
}
```

### Script de Invalidação Massiva

```bash
#!/bin/bash
# invalidate_flags.sh - Invalida múltiplas flags

ADMIN_URL="http://localhost:19000"
FLAGS=("new-feature" "beta-access" "premium-tier")

echo "🔄 Invalidando ${#FLAGS[@]} flags..."

for flag in "${FLAGS[@]}"; do
    echo -n "  - Invalidando '$flag'... "

    response=$(curl -s -X POST "${ADMIN_URL}/admin/invalidate" \
        -H "Content-Type: application/json" \
        -d "{\"flag_key\": \"$flag\"}")

    if echo "$response" | grep -q "success"; then
        echo "✅"
    else
        echo "❌ Falhou"
        echo "    Resposta: $response"
    fi
done

echo ""
echo "✅ Invalidação concluída!"
```

### Invalidação Programática em Go

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type InvalidateRequest struct {
    FlagKey string `json:"flag_key"`
}

type InvalidateResponse struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}

func invalidateFlag(adminURL, flagKey string) error {
    reqBody := InvalidateRequest{FlagKey: flagKey}
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return fmt.Errorf("erro ao serializar: %w", err)
    }

    resp, err := http.Post(
        adminURL+"/admin/invalidate",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return fmt.Errorf("erro na requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status code: %d", resp.StatusCode)
    }

    var result InvalidateResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("erro ao decodificar: %w", err)
    }

    fmt.Printf("✅ %s\n", result.Message)
    return nil
}

func invalidateAll(adminURL string) error {
    resp, err := http.Post(adminURL+"/admin/invalidate-all", "", nil)
    if err != nil {
        return fmt.Errorf("erro na requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status code: %d", resp.StatusCode)
    }

    fmt.Println("✅ Todas as flags foram invalidadas")
    return nil
}

func main() {
    adminURL := "http://localhost:19000"

    // Invalidar flags específicas
    flags := []string{"new-feature", "beta-access"}
    for _, flag := range flags {
        if err := invalidateFlag(adminURL, flag); err != nil {
            fmt.Printf("❌ Erro ao invalidar '%s': %v\n", flag, err)
        }
    }

    // Ou invalidar todas de uma vez
    // if err := invalidateAll(adminURL); err != nil {
    //     fmt.Printf("❌ Erro: %v\n", err)
    // }
}
```

---

## Refresh Manual

### Forçar Atualização de Flags

```bash
# Forçar refresh imediato de todas as flags
curl -X POST http://localhost:19000/admin/refresh

# Resposta:
{
  "status": "success",
  "message": "Flags refreshed from Flagr"
}
```

### Refresh Agendado

```bash
#!/bin/bash
# scheduled_refresh.sh - Executa refresh a cada 10 minutos

ADMIN_URL="http://localhost:19000"

while true; do
    echo "[$(date)] Executando refresh..."

    response=$(curl -s -X POST "${ADMIN_URL}/admin/refresh")

    if echo "$response" | grep -q "success"; then
        echo "[$(date)] ✅ Refresh bem-sucedido"
    else
        echo "[$(date)] ❌ Refresh falhou: $response"
    fi

    sleep 600  # 10 minutos
done
```

### Refresh com Verificação

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

func refreshFlags(adminURL string) error {
    resp, err := http.Post(adminURL+"/admin/refresh", "", nil)
    if err != nil {
        return fmt.Errorf("erro na requisição: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("status code: %d", resp.StatusCode)
    }

    var result map[string]string
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("erro ao decodificar: %w", err)
    }

    fmt.Printf("✅ %s\n", result["message"])
    return nil
}

func refreshWithRetry(adminURL string, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        if err := refreshFlags(adminURL); err == nil {
            return nil
        } else if i < maxRetries-1 {
            fmt.Printf("⚠️  Tentativa %d falhou, tentando novamente...\n", i+1)
            time.Sleep(time.Duration(i+1) * time.Second)
        } else {
            return fmt.Errorf("falhou após %d tentativas", maxRetries)
        }
    }
    return nil
}

func main() {
    adminURL := "http://localhost:19000"

    if err := refreshWithRetry(adminURL, 3); err != nil {
        fmt.Printf("❌ Erro: %v\n", err)
    }
}
```

---

## Integração com CI/CD

### GitHub Actions - Invalidar após Deploy

```yaml
# .github/workflows/deploy.yml
name: Deploy and Invalidate Cache

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Deploy Application
        run: |
          # Seus passos de deploy aqui
          echo "Deploying..."

      - name: Invalidate Vexilla Cache
        run: |
          curl -X POST https://vexilla.example.com/admin/invalidate-all \
            -H "Authorization: Bearer ${{ secrets.ADMIN_API_TOKEN }}"

      - name: Wait for Cache Refresh
        run: sleep 10

      - name: Verify Health
        run: |
          curl -f https://vexilla.example.com/health || exit 1
```

### GitLab CI - Invalidação Específica

```yaml
# .gitlab-ci.yml
stages:
  - deploy
  - cache-invalidate

deploy:
  stage: deploy
  script:
    - echo "Deploying application..."
    # Seus passos de deploy

invalidate-cache:
  stage: cache-invalidate
  script:
    - |
      # Invalidar flags relacionadas ao deploy
      for flag in new-ui beta-features premium-tier; do
        curl -X POST $VEXILLA_ADMIN_URL/admin/invalidate \
          -H "Content-Type: application/json" \
          -d "{\"flag_key\": \"$flag\"}"
      done
    - curl -X POST $VEXILLA_ADMIN_URL/admin/refresh
  only:
    - main
```

---

## Kubernetes Health Probes

### Deployment com Probes

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vexilla-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: vexilla-app
  template:
    metadata:
      labels:
        app: vexilla-app
    spec:
      containers:
      - name: app
        image: your-registry/vexilla-app:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 19000
          name: admin

        # Liveness Probe - Verifica se está vivo
        livenessProbe:
          httpGet:
            path: /health
            port: 19000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3

        # Readiness Probe - Verifica se está pronto
        readinessProbe:
          httpGet:
            path: /health
            port: 19000
          initialDelaySeconds: 10
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 2

        # Startup Probe - Para inicialização lenta
        startupProbe:
          httpGet:
            path: /health
            port: 19000
          initialDelaySeconds: 0
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 30  # 30 * 10s = 5min para iniciar

        env:
        - name: FLAGR_ENDPOINT
          value: "http://flagr-service:18000"
        - name: ADMIN_PORT
          value: "19000"
```

### Service para Admin API

```yaml
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: vexilla-admin
spec:
  selector:
    app: vexilla-app
  ports:
  - name: admin
    port: 19000
    targetPort: 19000
  type: ClusterIP
```

### Job para Invalidação Pós-Deploy

```yaml
# k8s/invalidate-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: vexilla-invalidate-cache
  annotations:
    argocd.argoproj.io/hook: PostSync
    argocd.argoproj.io/hook-delete-policy: HookSucceeded
spec:
  template:
    spec:
      containers:
      - name: invalidate
        image: curlimages/curl:latest
        command:
        - sh
        - -c
        - |
          echo "Invalidando cache do Vexilla..."
          curl -X POST http://vexilla-admin:19000/admin/invalidate-all
          echo "Cache invalidado com sucesso!"
      restartPolicy: Never
  backoffLimit: 3
```

---

## Dashboard de Monitoramento

### Prometheus Metrics Exporter

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    cacheHitRatio = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vexilla_cache_hit_ratio",
        Help: "Cache hit ratio (0-1)",
    })

    keysAdded = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vexilla_keys_added_total",
        Help: "Total keys added to cache",
    })

    circuitOpen = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "vexilla_circuit_open",
        Help: "Circuit breaker status (1=open, 0=closed)",
    })
)

func init() {
    prometheus.MustRegister(cacheHitRatio)
    prometheus.MustRegister(keysAdded)
    prometheus.MustRegister(circuitOpen)
}

type CacheMetrics struct {
    Storage struct {
        HitRatio float64 `json:"hit_ratio"`
        KeysAdded int    `json:"keys_added"`
    } `json:"storage"`
    CircuitOpen bool `json:"circuit_open"`
}

func collectMetrics(adminURL string) {
    ticker := time.NewTicker(15 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        resp, err := http.Get(adminURL + "/admin/stats")
        if err != nil {
            continue
        }

        var metrics CacheMetrics
        json.NewDecoder(resp.Body).Decode(&metrics)
        resp.Body.Close()

        cacheHitRatio.Set(metrics.Storage.HitRatio)
        keysAdded.Set(float64(metrics.Storage.KeysAdded))
        if metrics.CircuitOpen {
            circuitOpen.Set(1)
        } else {
            circuitOpen.Set(0)
        }
    }
}

func main() {
    // Coletar métricas do Vexilla
    go collectMetrics("http://localhost:19000")

    // Expor métricas para Prometheus
    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":9090", nil)
}
```

### Grafana Dashboard JSON

Salve como `grafana-dashboard.json`:

```json
{
  "dashboard": {
    "title": "Vexilla Monitoring",
    "panels": [
      {
        "title": "Cache Hit Ratio",
        "targets": [
          {
            "expr": "vexilla_cache_hit_ratio * 100"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Circuit Breaker Status",
        "targets": [
          {
            "expr": "vexilla_circuit_open"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Keys Added",
        "targets": [
          {
            "expr": "rate(vexilla_keys_added_total[5m])"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

---

## Conclusão

A Admin API do Vexilla fornece endpoints HTTP simples e poderosos para:

- ✅ **Monitoramento** - Health checks e métricas em tempo real
- ✅ **Operações** - Invalidação e refresh sob demanda
- ✅ **Integração** - Fácil integração com CI/CD e Kubernetes
- ✅ **Observabilidade** - Exportação para Prometheus/Grafana

Para mais informações:
- [Server Features Guide](../SERVER_FEATURES.md)
- [API Reference](https://pkg.go.dev/github.com/OrlandoBitencourt/vexilla)
- [Examples](../examples/)
