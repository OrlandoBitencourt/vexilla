package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

type CacheMetrics struct {
	Storage struct {
		KeysAdded   uint64  `json:"keys_added"`
		KeysEvicted uint64  `json:"keys_evicted"`
		HitRatio    float64 `json:"hit_ratio"`
	} `json:"storage"`
	LastRefresh      time.Time `json:"last_refresh"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	CircuitOpen      bool      `json:"circuit_open"`
}

func main() {
	fmt.Println("📊 Vexilla Admin API - Metrics")
	fmt.Println("================================\n")

	// Iniciar Vexilla
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithAdminServer(vexilla.AdminConfig{Port: 19000}),
		vexilla.WithRefreshInterval(1*time.Minute),
	)
	if err != nil {
		log.Fatalf("❌ Erro: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Vexilla iniciado\n")
	time.Sleep(500 * time.Millisecond)

	// Exemplo 1: Obter métricas programaticamente
	fmt.Println("📍 Exemplo 1: Métricas Programáticas (via SDK)")
	fmt.Println("-----------------------------------------------")
	showProgrammaticMetrics(client)
	fmt.Println()

	// Exemplo 2: Obter métricas via HTTP
	fmt.Println("📍 Exemplo 2: Métricas via HTTP (Admin API)")
	fmt.Println("-------------------------------------------")
	if err := showHTTPMetrics("http://localhost:19000"); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	}
	fmt.Println()

	// Gerar alguma atividade
	fmt.Println("📍 Exemplo 3: Gerando Atividade")
	fmt.Println("--------------------------------")
	generateActivity(client, ctx)
	fmt.Println()

	// Mostrar métricas atualizadas
	fmt.Println("📍 Exemplo 4: Métricas Após Atividade")
	fmt.Println("-------------------------------------")
	showProgrammaticMetrics(client)
	fmt.Println()

	// Exemplo 5: Monitoramento com alertas
	fmt.Println("📍 Exemplo 5: Monitoramento com Alertas (15 segundos)")
	fmt.Println("-----------------------------------------------------")
	monitorWithAlerts("http://localhost:19000", 3*time.Second, 15*time.Second)

	fmt.Println("\n✅ Exemplos concluídos!")
}

// showProgrammaticMetrics mostra métricas via SDK
func showProgrammaticMetrics(client *vexilla.Client) {
	metrics := client.Metrics()

	fmt.Println("  📈 Cache Performance:")
	fmt.Printf("    • Keys Added:   %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("    • Keys Evicted: %d\n", metrics.Storage.KeysEvicted)
	fmt.Printf("    • Hit Ratio:    %.2f%%\n", metrics.Storage.HitRatio*100)

	fmt.Println("\n  🔄 Refresh Status:")
	if !metrics.LastRefresh.IsZero() {
		fmt.Printf("    • Last Refresh: %s\n", metrics.LastRefresh.Format("15:04:05"))
	} else {
		fmt.Printf("    • Last Refresh: Nunca\n")
	}

	fmt.Println("\n  ⚡ Circuit Breaker:")
	fmt.Printf("    • Status:           %s\n", circuitStatus(metrics.CircuitOpen))
	fmt.Printf("    • Consecutive Fails: %d\n", metrics.ConsecutiveFails)
}

// showHTTPMetrics obtém e mostra métricas via HTTP
func showHTTPMetrics(baseURL string) error {
	fmt.Println("  🌐 Consultando Admin API...")

	resp, err := http.Get(baseURL + "/admin/stats")
	if err != nil {
		return fmt.Errorf("falha ao conectar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var metrics CacheMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return fmt.Errorf("falha ao decodificar: %w", err)
	}

	// Mostrar JSON formatado
	fmt.Println("\n  📄 Resposta JSON:")
	prettyJSON, _ := json.MarshalIndent(metrics, "    ", "  ")
	fmt.Printf("    %s\n", prettyJSON)

	return nil
}

// generateActivity gera atividade para popular métricas
func generateActivity(client *vexilla.Client, ctx context.Context) {
	fmt.Println("  🔄 Gerando 100 avaliações de flags...")

	for i := 0; i < 100; i++ {
		evalCtx := vexilla.NewContext(fmt.Sprintf("user-%d", i)).
			WithAttribute("tier", "premium").
			WithAttribute("country", "BR")

		_ = client.Bool(ctx, "test-flag", evalCtx)
		_ = client.String(ctx, "theme", evalCtx, "light")
		_ = client.Int(ctx, "limit", evalCtx, 100)
	}

	fmt.Println("  ✅ Atividade gerada (300 avaliações)")
}

// monitorWithAlerts monitora métricas e gera alertas
func monitorWithAlerts(baseURL string, interval, duration time.Duration) {
	fmt.Printf("  🔍 Monitorando com alertas a cada %v...\n\n", interval)

	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)
	checkCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("\n  ✅ Monitoramento concluído (%d checks)\n", checkCount)
			return

		case <-ticker.C:
			checkCount++
			resp, err := client.Get(baseURL + "/admin/stats")
			if err != nil {
				fmt.Printf("  [%02d] ❌ Erro: %v\n", checkCount, err)
				continue
			}

			var metrics CacheMetrics
			json.NewDecoder(resp.Body).Decode(&metrics)
			resp.Body.Close()

			// Verificar alertas
			alerts := checkAlerts(metrics)

			if len(alerts) > 0 {
				fmt.Printf("  [%02d] ⚠️  ALERTAS:\n", checkCount)
				for _, alert := range alerts {
					fmt.Printf("        %s\n", alert)
				}
			} else {
				fmt.Printf("  [%02d] ✅ OK - Hit Ratio: %.1f%%, Circuit: %s\n",
					checkCount,
					metrics.Storage.HitRatio*100,
					circuitStatus(metrics.CircuitOpen))
			}
		}
	}
}

// checkAlerts verifica condições de alerta
func checkAlerts(metrics CacheMetrics) []string {
	var alerts []string

	// Alert: Low hit ratio
	if metrics.Storage.HitRatio < 0.8 {
		alerts = append(alerts,
			fmt.Sprintf("🔴 Hit ratio baixo: %.1f%% (esperado >80%%)",
				metrics.Storage.HitRatio*100))
	}

	// Alert: Circuit breaker aberto
	if metrics.CircuitOpen {
		alerts = append(alerts, "🔴 Circuit breaker ABERTO - Flagr indisponível")
	}

	// Alert: Muitas falhas consecutivas
	if metrics.ConsecutiveFails > 2 {
		alerts = append(alerts,
			fmt.Sprintf("🟡 %d falhas consecutivas de refresh", metrics.ConsecutiveFails))
	}

	// Alert: Alta taxa de eviction
	if metrics.Storage.KeysEvicted > metrics.Storage.KeysAdded/2 {
		alerts = append(alerts,
			fmt.Sprintf("🟡 Alta taxa de eviction: %d/%d (%.1f%%)",
				metrics.Storage.KeysEvicted,
				metrics.Storage.KeysAdded,
				float64(metrics.Storage.KeysEvicted)/float64(metrics.Storage.KeysAdded)*100))
	}

	return alerts
}

func circuitStatus(open bool) string {
	if open {
		return "🔴 ABERTO"
	}
	return "🟢 FECHADO"
}
