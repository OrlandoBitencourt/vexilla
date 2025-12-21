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

type RefreshResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	fmt.Println("🔄 Vexilla Admin API - Manual Refresh")
	fmt.Println("======================================\n")

	// Iniciar Vexilla com refresh interval longo
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithAdminServer(vexilla.AdminConfig{Port: 19000}),
		vexilla.WithRefreshInterval(1*time.Hour), // Refresh automático muito longo
	)
	if err != nil {
		log.Fatalf("❌ Erro: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Vexilla iniciado (auto-refresh: 1 hora)")
	fmt.Println()
	time.Sleep(500 * time.Millisecond)

	// Exemplo 1: Refresh via SDK
	fmt.Println("📍 Exemplo 1: Refresh Manual via SDK")
	fmt.Println("-------------------------------------")
	showLastRefresh(client)
	fmt.Println("  🔄 Executando refresh...")
	if err := client.Sync(ctx); err != nil {
		log.Printf("  ❌ Erro: %v\n", err)
	} else {
		fmt.Println("  ✅ Refresh concluído via SDK")
	}
	showLastRefresh(client)
	fmt.Println()

	// Aguardar um pouco
	time.Sleep(2 * time.Second)

	// Exemplo 2: Refresh via HTTP
	fmt.Println("📍 Exemplo 2: Refresh Manual via HTTP")
	fmt.Println("--------------------------------------")
	showLastRefresh(client)
	if err := refreshHTTP("http://localhost:19000"); err != nil {
		log.Printf("  ❌ Erro: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)
	showLastRefresh(client)
	fmt.Println()

	// Exemplo 3: Refresh com retry
	fmt.Println("📍 Exemplo 3: Refresh com Retry Automático")
	fmt.Println("------------------------------------------")
	if err := refreshWithRetry("http://localhost:19000", 3); err != nil {
		log.Printf("  ❌ Erro: %v\n", err)
	}
	fmt.Println()

	// Exemplo 4: Refresh agendado
	fmt.Println("📍 Exemplo 4: Refresh Agendado (15 segundos)")
	fmt.Println("--------------------------------------------")
	scheduledRefresh("http://localhost:19000", 5*time.Second, 15*time.Second)

	fmt.Println("\n✅ Exemplos de refresh concluídos!")
}

// showLastRefresh mostra quando foi o último refresh
func showLastRefresh(client *vexilla.Client) {
	metrics := client.Metrics()

	if metrics.LastRefresh.IsZero() {
		fmt.Println("  ⏰ Último refresh: Nunca")
	} else {
		elapsed := time.Since(metrics.LastRefresh)
		fmt.Printf("  ⏰ Último refresh: %s (%s atrás)\n",
			metrics.LastRefresh.Format("15:04:05"),
			formatDuration(elapsed))
	}
}

// refreshHTTP executa refresh via HTTP
func refreshHTTP(baseURL string) error {
	fmt.Println("  🌐 Executando refresh via HTTP...")

	start := time.Now()
	resp, err := http.Post(baseURL+"/admin/refresh", "", nil)
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result RefreshResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("erro ao decodificar: %w", err)
	}

	fmt.Printf("  ✅ Refresh concluído em %v\n", elapsed)
	fmt.Printf("  📝 Resposta: %s - %s\n", result.Status, result.Message)

	return nil
}

// refreshWithRetry tenta refresh com retry
func refreshWithRetry(baseURL string, maxRetries int) error {
	fmt.Printf("  🔁 Tentando refresh (máx %d tentativas)...\n", maxRetries)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("    Tentativa %d/%d... ", attempt, maxRetries)

		resp, err := http.Post(baseURL+"/admin/refresh", "", nil)
		if err != nil {
			fmt.Printf("❌ Erro de conexão: %v\n", err)
			if attempt < maxRetries {
				waitTime := time.Duration(attempt) * time.Second
				fmt.Printf("    ⏳ Aguardando %v antes de tentar novamente...\n", waitTime)
				time.Sleep(waitTime)
				continue
			}
			return fmt.Errorf("falhou após %d tentativas", maxRetries)
		}

		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅")
			fmt.Printf("  ✅ Refresh bem-sucedido na tentativa %d\n", attempt)
			return nil
		}

		fmt.Printf("❌ Status %d\n", resp.StatusCode)
		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			fmt.Printf("    ⏳ Aguardando %v antes de tentar novamente...\n", waitTime)
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("falhou após %d tentativas", maxRetries)
}

// scheduledRefresh executa refresh em intervalos regulares
func scheduledRefresh(baseURL string, interval, duration time.Duration) {
	fmt.Printf("  ⏰ Executando refresh a cada %v por %v...\n\n", interval, duration)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)
	refreshCount := 0
	successCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("\n  📊 Resultado: %d/%d refreshes bem-sucedidos (%.1f%%)\n",
				successCount, refreshCount,
				float64(successCount)/float64(refreshCount)*100)
			return

		case t := <-ticker.C:
			refreshCount++
			fmt.Printf("  [%s] Executando refresh #%d... ",
				t.Format("15:04:05"), refreshCount)

			start := time.Now()
			resp, err := http.Post(baseURL+"/admin/refresh", "", nil)
			elapsed := time.Since(start)

			if err != nil {
				fmt.Printf("❌ Erro: %v\n", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				successCount++
				fmt.Printf("✅ (completou em %v)\n", elapsed)
			} else {
				fmt.Printf("⚠️  Status %d\n", resp.StatusCode)
			}
		}
	}
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
