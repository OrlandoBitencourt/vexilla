package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/OrlandoBitencourt/vexilla"
)

func main() {
	fmt.Println("🚀 Vexilla Admin API - Basic Setup")
	fmt.Println("====================================\n")

	// Criar client com Admin API habilitada
	client, err := vexilla.New(
		// Conexão com Flagr
		vexilla.WithFlagrEndpoint("http://localhost:18000"),
		vexilla.WithRefreshInterval(5*time.Minute),

		// IMPORTANTE: Habilitar Admin API
		vexilla.WithAdminServer(vexilla.AdminConfig{
			Port: 19000, // Admin API na porta 19000
		}),

		// Configurações adicionais
		vexilla.WithOnlyEnabled(true),
		vexilla.WithCircuitBreaker(3, 30*time.Second),
	)
	if err != nil {
		log.Fatalf("❌ Erro ao criar client: %v", err)
	}

	// Iniciar client
	ctx := context.Background()
	if err := client.Start(ctx); err != nil {
		log.Fatalf("❌ Erro ao iniciar client: %v", err)
	}
	defer client.Stop()

	fmt.Println("✅ Vexilla iniciado com sucesso!")
	fmt.Println()
	fmt.Println("📡 Admin API disponível em: http://localhost:19000")
	fmt.Println()
	fmt.Println("Endpoints disponíveis:")
	fmt.Println("  - GET  http://localhost:19000/health           (Health check)")
	fmt.Println("  - GET  http://localhost:19000/admin/stats      (Métricas)")
	fmt.Println("  - POST http://localhost:19000/admin/invalidate (Invalidar flag)")
	fmt.Println("  - POST http://localhost:19000/admin/refresh    (Refresh)")
	fmt.Println()
	fmt.Println("📝 Teste com curl:")
	fmt.Println("  curl http://localhost:19000/health")
	fmt.Println("  curl http://localhost:19000/admin/stats | jq")
	fmt.Println()

	// Simular algumas avaliações de flags para popular métricas
	fmt.Println("🔄 Simulando avaliações de flags...")
	for i := 0; i < 10; i++ {
		evalCtx := vexilla.NewContext(fmt.Sprintf("user-%d", i))
		_ = client.Bool(ctx, "test-flag", evalCtx)
	}
	fmt.Println("✅ Avaliações concluídas\n")

	// Mostrar métricas programaticamente
	metrics := client.Metrics()
	fmt.Println("📊 Métricas atuais:")
	fmt.Printf("  - Keys Added: %d\n", metrics.Storage.KeysAdded)
	fmt.Printf("  - Hit Ratio: %.2f%%\n", metrics.Storage.HitRatio*100)
	fmt.Printf("  - Circuit Open: %v\n", metrics.CircuitOpen)
	fmt.Println()

	// Aguardar sinal de interrupção
	fmt.Println("⏳ Pressione Ctrl+C para parar...")
	waitForShutdown()

	fmt.Println("\n👋 Encerrando...")
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
}
