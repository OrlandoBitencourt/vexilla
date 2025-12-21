package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

type InvalidateRequest struct {
	FlagKey string `json:"flag_key"`
}

type InvalidateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	fmt.Println("🗑️  Vexilla Admin API - Cache Invalidation")
	fmt.Println("===========================================\n")

	// Iniciar Vexilla
	client, err := vexilla.New(
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithAdminServer(vexilla.AdminConfig{Port: 19000}),
		vexilla.WithRefreshInterval(10*time.Minute),
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

	// Popular cache com algumas avaliações
	fmt.Println("📍 Populando cache...")
	fmt.Println("---------------------")
	populateCache(client, ctx)
	showCacheStats(client)
	fmt.Println()

	// Exemplo 1: Invalidar flag via SDK
	fmt.Println("📍 Exemplo 1: Invalidar Flag via SDK")
	fmt.Println("-------------------------------------")
	if err := client.InvalidateFlag(ctx, "test-flag-1"); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	} else {
		fmt.Println("✅ Flag 'test-flag-1' invalidada via SDK")
	}
	fmt.Println()

	// Exemplo 2: Invalidar flag via HTTP
	fmt.Println("📍 Exemplo 2: Invalidar Flag via HTTP")
	fmt.Println("--------------------------------------")
	if err := invalidateFlagHTTP("http://localhost:19000", "test-flag-2"); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	}
	fmt.Println()

	// Exemplo 3: Invalidar múltiplas flags
	fmt.Println("📍 Exemplo 3: Invalidar Múltiplas Flags")
	fmt.Println("----------------------------------------")
	flags := []string{"test-flag-3", "test-flag-4", "test-flag-5"}
	invalidateMultipleFlags("http://localhost:19000", flags)
	fmt.Println()

	// Exemplo 4: Invalidar todas as flags via SDK
	fmt.Println("📍 Exemplo 4: Invalidar Todas as Flags (SDK)")
	fmt.Println("--------------------------------------------")
	if err := client.InvalidateAll(ctx); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	} else {
		fmt.Println("✅ Todas as flags invalidadas via SDK")
	}
	showCacheStats(client)
	fmt.Println()

	// Popular novamente
	fmt.Println("📍 Populando cache novamente...")
	populateCache(client, ctx)
	showCacheStats(client)
	fmt.Println()

	// Exemplo 5: Invalidar todas as flags via HTTP
	fmt.Println("📍 Exemplo 5: Invalidar Todas as Flags (HTTP)")
	fmt.Println("---------------------------------------------")
	if err := invalidateAllHTTP("http://localhost:19000"); err != nil {
		log.Printf("❌ Erro: %v\n", err)
	}
	showCacheStats(client)
	fmt.Println()

	fmt.Println("✅ Exemplos de invalidação concluídos!")
}

// populateCache popula o cache com avaliações
func populateCache(client *vexilla.Client, ctx context.Context) {
	flags := []string{
		"test-flag-1",
		"test-flag-2",
		"test-flag-3",
		"test-flag-4",
		"test-flag-5",
	}

	for _, flag := range flags {
		evalCtx := vexilla.NewContext("user-test")
		_ = client.Bool(ctx, flag, evalCtx)
	}

	fmt.Println("  ✅ Cache populado com avaliações")
}

// showCacheStats mostra estatísticas atuais do cache
func showCacheStats(client *vexilla.Client) {
	metrics := client.Metrics()
	fmt.Printf("  📊 Cache: %d keys, Hit Ratio: %.1f%%\n",
		metrics.Storage.KeysAdded,
		metrics.Storage.HitRatio*100)
}

// invalidateFlagHTTP invalida uma flag via HTTP
func invalidateFlagHTTP(baseURL, flagKey string) error {
	fmt.Printf("  🌐 Invalidando '%s' via HTTP...\n", flagKey)

	reqBody := InvalidateRequest{FlagKey: flagKey}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("erro ao serializar: %w", err)
	}

	resp, err := http.Post(
		baseURL+"/admin/invalidate",
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

	fmt.Printf("  ✅ Resposta: %s - %s\n", result.Status, result.Message)
	return nil
}

// invalidateMultipleFlags invalida múltiplas flags
func invalidateMultipleFlags(baseURL string, flags []string) {
	fmt.Printf("  🌐 Invalidando %d flags...\n", len(flags))

	successCount := 0
	for i, flag := range flags {
		fmt.Printf("    [%d/%d] Invalidando '%s'... ", i+1, len(flags), flag)

		if err := invalidateFlagHTTP(baseURL, flag); err != nil {
			fmt.Printf("❌ Falhou: %v\n", err)
		} else {
			fmt.Printf("✅\n")
			successCount++
		}
	}

	fmt.Printf("\n  📊 Resultado: %d/%d flags invalidadas com sucesso\n",
		successCount, len(flags))
}

// invalidateAllHTTP invalida todas as flags via HTTP
func invalidateAllHTTP(baseURL string) error {
	fmt.Println("  🌐 Invalidando TODAS as flags via HTTP...")

	resp, err := http.Post(baseURL+"/admin/invalidate-all", "", nil)
	if err != nil {
		return fmt.Errorf("erro na requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result InvalidateResponse
	json.NewDecoder(resp.Body).Decode(&result)

	fmt.Printf("  ✅ Resposta: %s - %s\n", result.Status, result.Message)
	return nil
}
