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

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

func main() {
	fmt.Println("🏥 Vexilla Admin API - Health Check")
	fmt.Println("=====================================\n")

	// Iniciar Vexilla com Admin API
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithAdminServer(vexilla.AdminConfig{Port: 19000}),
	)
	if err != nil {
		log.Fatalf("❌ Erro: %v", err)
	}

	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Vexilla iniciado - Admin API: http://localhost:19000\n")

	// Aguardar um pouco para garantir que o servidor está pronto
	time.Sleep(500 * time.Millisecond)

	// Exemplo 1: Health check básico
	fmt.Println("📍 Exemplo 1: Health Check Básico")
	fmt.Println("-----------------------------------")
	if err := checkHealth("http://localhost:19000"); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	}
	fmt.Println()

	// Exemplo 2: Health check com timeout
	fmt.Println("📍 Exemplo 2: Health Check com Timeout")
	fmt.Println("---------------------------------------")
	if err := checkHealthWithTimeout("http://localhost:19000", 3*time.Second); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	}
	fmt.Println()

	// Exemplo 3: Loop de monitoramento
	fmt.Println("📍 Exemplo 3: Monitoramento Contínuo (10 segundos)")
	fmt.Println("--------------------------------------------------")
	monitorHealth("http://localhost:19000", 2*time.Second, 10*time.Second)

	fmt.Println("\n✅ Exemplos concluídos!")
}

// checkHealth faz um health check básico
func checkHealth(baseURL string) error {
	fmt.Println("  🔍 Verificando saúde do serviço...")

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		return fmt.Errorf("falha ao conectar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status não-saudável: %d", resp.StatusCode)
	}

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return fmt.Errorf("falha ao decodificar: %w", err)
	}

	fmt.Printf("  ✅ Status: %s\n", health.Status)
	fmt.Printf("  🕐 Timestamp: %s\n", health.Timestamp)

	return nil
}

// checkHealthWithTimeout faz health check com timeout customizado
func checkHealthWithTimeout(baseURL string, timeout time.Duration) error {
	fmt.Printf("  🔍 Health check com timeout de %v...\n", timeout)

	client := &http.Client{Timeout: timeout}

	start := time.Now()
	resp, err := client.Get(baseURL + "/health")
	elapsed := time.Since(start)

	if err != nil {
		return fmt.Errorf("falha após %v: %w", elapsed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status não-saudável: %d", resp.StatusCode)
	}

	var health HealthResponse
	json.NewDecoder(resp.Body).Decode(&health)

	fmt.Printf("  ✅ Status: %s (respondeu em %v)\n", health.Status, elapsed)

	return nil
}

// monitorHealth monitora a saúde por um período determinado
func monitorHealth(baseURL string, interval, duration time.Duration) {
	fmt.Printf("  🔄 Monitorando a cada %v por %v...\n\n", interval, duration)

	client := &http.Client{Timeout: 5 * time.Second}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)
	checkCount := 0
	successCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("\n  📊 Resultado: %d/%d checks bem-sucedidos (%.1f%%)\n",
				successCount, checkCount, float64(successCount)/float64(checkCount)*100)
			return

		case <-ticker.C:
			checkCount++
			resp, err := client.Get(baseURL + "/health")

			if err != nil {
				fmt.Printf("  [%02d] ❌ Falha ao conectar: %v\n", checkCount, err)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				successCount++
				var health HealthResponse
				json.NewDecoder(resp.Body).Decode(&health)
				fmt.Printf("  [%02d] ✅ Healthy - %s\n", checkCount, health.Timestamp)
			} else {
				fmt.Printf("  [%02d] ⚠️  Status %d\n", checkCount, resp.StatusCode)
			}
			resp.Body.Close()
		}
	}
}
