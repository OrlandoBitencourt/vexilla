package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
	fmt.Println("📊 Vexilla Admin API - Monitoring Dashboard")
	fmt.Println("============================================\n")

	// Iniciar Vexilla
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithAdminServer(vexilla.AdminConfig{Port: 19000}),
		vexilla.WithRefreshInterval(30*time.Second),
		vexilla.WithCircuitBreaker(3, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("❌ Erro: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Vexilla iniciado - Dashboard disponível\n")
	time.Sleep(500 * time.Millisecond)

	// Gerar atividade em background
	fmt.Println("🔄 Gerando atividade contínua em background...\n")
	go generateContinuousActivity(client, ctx)

	// Aguardar um pouco para gerar dados
	time.Sleep(2 * time.Second)

	// Executar dashboard
	fmt.Println("📊 Iniciando dashboard de monitoramento...")
	fmt.Println("Pressione Ctrl+C para parar\n")
	runDashboard("http://localhost:19000", 2*time.Second)
}

// runDashboard executa dashboard em tempo real
func runDashboard(baseURL string, refreshInterval time.Duration) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	startTime := time.Now()

	for {
		select {
		case <-ticker.C:
			clearScreen()
			displayDashboard(httpClient, baseURL, startTime)
		}
	}
}

// displayDashboard exibe o dashboard completo
func displayDashboard(client *http.Client, baseURL string, startTime time.Time) {
	// Header
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║          VEXILLA MONITORING DASHBOARD                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Uptime
	uptime := time.Since(startTime)
	fmt.Printf("⏱️  Uptime: %s\n", formatDuration(uptime))
	fmt.Printf("🕐 Atualizado: %s\n", time.Now().Format("15:04:05"))
	fmt.Println()

	// Buscar métricas
	resp, err := client.Get(baseURL + "/admin/stats")
	if err != nil {
		fmt.Printf("❌ Erro ao buscar métricas: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var metrics CacheMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		fmt.Printf("❌ Erro ao decodificar métricas: %v\n", err)
		return
	}

	// Cache Performance
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 📈 CACHE PERFORMANCE                                        │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Keys Added:       %-42d │\n", metrics.Storage.KeysAdded)
	fmt.Printf("│ Keys Evicted:     %-42d │\n", metrics.Storage.KeysEvicted)
	fmt.Printf("│ Hit Ratio:        %-42s │\n", formatHitRatio(metrics.Storage.HitRatio))
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Hit Ratio Bar
	displayHitRatioBar(metrics.Storage.HitRatio)
	fmt.Println()

	// Circuit Breaker Status
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ ⚡ CIRCUIT BREAKER                                          │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ Status:           %-42s │\n", formatCircuitStatus(metrics.CircuitOpen))
	fmt.Printf("│ Consecutive Fails: %-41d │\n", metrics.ConsecutiveFails)
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Refresh Status
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│ 🔄 REFRESH STATUS                                           │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")
	if !metrics.LastRefresh.IsZero() {
		elapsed := time.Since(metrics.LastRefresh)
		fmt.Printf("│ Last Refresh:     %-42s │\n", metrics.LastRefresh.Format("15:04:05"))
		fmt.Printf("│ Time Elapsed:     %-42s │\n", formatDuration(elapsed))
	} else {
		fmt.Printf("│ Last Refresh:     %-42s │\n", "Never")
	}
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Println()

	// Alertas
	displayAlerts(metrics)

	// Footer
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Println("Admin API: http://localhost:19000")
	fmt.Println("Atualização automática a cada 2 segundos | Ctrl+C para sair")
}

// displayHitRatioBar exibe barra de progresso para hit ratio
func displayHitRatioBar(ratio float64) {
	barLength := 50
	filled := int(ratio * float64(barLength))

	fmt.Print("Hit Ratio: [")
	for i := 0; i < barLength; i++ {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}
	fmt.Printf("] %.1f%%\n", ratio*100)
}

// displayAlerts exibe alertas ativos
func displayAlerts(metrics CacheMetrics) {
	alerts := []string{}

	// Verificar condições de alerta
	if metrics.Storage.HitRatio < 0.8 {
		alerts = append(alerts,
			fmt.Sprintf("🔴 Hit ratio baixo: %.1f%% (esperado >80%%)",
				metrics.Storage.HitRatio*100))
	}

	if metrics.CircuitOpen {
		alerts = append(alerts, "🔴 Circuit breaker ABERTO - Flagr indisponível")
	}

	if metrics.ConsecutiveFails > 2 {
		alerts = append(alerts,
			fmt.Sprintf("🟡 %d falhas consecutivas de refresh", metrics.ConsecutiveFails))
	}

	if metrics.Storage.KeysEvicted > metrics.Storage.KeysAdded/2 && metrics.Storage.KeysAdded > 0 {
		evictionRate := float64(metrics.Storage.KeysEvicted) / float64(metrics.Storage.KeysAdded) * 100
		alerts = append(alerts,
			fmt.Sprintf("🟡 Alta taxa de eviction: %.1f%%", evictionRate))
	}

	// Mostrar alertas
	if len(alerts) > 0 {
		fmt.Println("┌─────────────────────────────────────────────────────────────┐")
		fmt.Println("│ ⚠️  ALERTAS ATIVOS                                          │")
		fmt.Println("├─────────────────────────────────────────────────────────────┤")
		for _, alert := range alerts {
			fmt.Printf("│ %-60s│\n", alert)
		}
		fmt.Println("└─────────────────────────────────────────────────────────────┘")
		fmt.Println()
	} else {
		fmt.Println("✅ Nenhum alerta ativo - Sistema operando normalmente")
		fmt.Println()
	}
}

// generateContinuousActivity gera atividade contínua
func generateContinuousActivity(client *vexilla.Client, ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	userID := 0
	for range ticker.C {
		userID++
		evalCtx := vexilla.NewContext(fmt.Sprintf("user-%d", userID)).
			WithAttribute("country", "BR").
			WithAttribute("tier", "premium")

		_ = client.Bool(ctx, "feature-a", evalCtx)
		_ = client.Bool(ctx, "feature-b", evalCtx)
		_ = client.String(ctx, "theme", evalCtx, "light")
		_ = client.Int(ctx, "limit", evalCtx, 100)
	}
}

// formatHitRatio formata hit ratio com cor
func formatHitRatio(ratio float64) string {
	percentage := ratio * 100
	emoji := "✅"

	if percentage < 80 {
		emoji = "⚠️"
	}
	if percentage < 60 {
		emoji = "❌"
	}

	return fmt.Sprintf("%s %.2f%%", emoji, percentage)
}

// formatCircuitStatus formata status do circuit breaker
func formatCircuitStatus(open bool) string {
	if open {
		return "🔴 ABERTO (Flagr indisponível)"
	}
	return "🟢 FECHADO (Operacional)"
}

// formatDuration formata duração de forma legível
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// clearScreen limpa a tela do terminal
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
