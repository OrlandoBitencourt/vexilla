.PHONY: help start stop setup clean dev backend frontend

help: ## Show this help message
	@echo "Vexilla Complete Demo - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

start: ## Start all services (Flagr, Backend, Frontend)
	@echo "🚀 Starting Vexilla Demo..."
	@echo ""
	@echo "1️⃣  Starting Flagr..."
	@docker-compose up -d flagr
	@echo "⏳ Waiting for Flagr to be ready..."
	@sleep 5
	@echo ""
	@echo "2️⃣  Setting up flags..."
	@cd scripts && chmod +x setup-flags.sh && ./setup-flags.sh
	@echo ""
	@echo "3️⃣  Starting backend..."
	@cd backend && go run main.go &
	@echo ""
	@echo "4️⃣  Starting frontend..."
	@cd frontend && npm run dev
	@echo ""
	@echo "✅ All services started!"
	@echo "📊 Flagr UI:    http://localhost:18000"
	@echo "🔌 Backend API: http://localhost:8080"
	@echo "🎨 Frontend:    http://localhost:3000"

stop: ## Stop all services
	@echo "🛑 Stopping all services..."
	@docker-compose down
	@pkill -f "go run main.go" || true
	@pkill -f "next dev" || true
	@echo "✅ All services stopped"

setup: ## Setup flags in Flagr
	@echo "🏗️  Setting up flags..."
	@cd scripts && chmod +x setup-flags.sh && ./setup-flags.sh

clean: ## Clean up Docker volumes and build artifacts
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@rm -rf backend/tmp
	@rm -rf frontend/.next
	@rm -rf frontend/node_modules
	@echo "✅ Cleanup complete"

flagr: ## Start only Flagr
	@echo "🏴 Starting Flagr..."
	@docker-compose up -d flagr
	@echo "✅ Flagr started on http://localhost:18000"

backend: ## Start only backend
	@echo "🔌 Starting backend..."
	@cd backend && go run main.go

frontend: ## Start only frontend
	@echo "🎨 Starting frontend..."
	@cd frontend && npm run dev

dev: ## Development mode (start Flagr + setup flags)
	@echo "🛠️  Development mode"
	@docker-compose up -d flagr
	@sleep 5
	@cd scripts && chmod +x setup-flags.sh && ./setup-flags.sh
	@echo ""
	@echo "✅ Ready for development!"
	@echo ""
	@echo "Run these in separate terminals:"
	@echo "  Terminal 1: make backend"
	@echo "  Terminal 2: make frontend"

logs: ## Show Flagr logs
	@docker-compose logs -f flagr

test-checkout: ## Test checkout endpoint with different CPFs
	@echo "Testing checkout with different CPFs..."
	@echo ""
	@echo "CPF 1: 12345678909"
	@curl -X POST http://localhost:8080/checkout -H "X-User-ID: user-1" -H "X-CPF: 12345678909" | jq
	@echo ""
	@echo "CPF 2: 98765432100"
	@curl -X POST http://localhost:8080/checkout -H "X-User-ID: user-2" -H "X-CPF: 98765432100" | jq
	@echo ""
	@echo "CPF 3: 11122233344"
	@curl -X POST http://localhost:8080/checkout -H "X-User-ID: user-3" -H "X-CPF: 11122233344" | jq

health: ## Check health of all services
	@echo "🏥 Health Check"
	@echo ""
	@echo "Flagr:"
	@curl -s http://localhost:18000/api/v1/health | jq || echo "❌ Flagr not running"
	@echo ""
	@echo "Backend:"
	@curl -s http://localhost:8080/health | jq || echo "❌ Backend not running"
	@echo ""
	@echo "Frontend:"
	@curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 && echo " ✅ Frontend running" || echo "❌ Frontend not running"
