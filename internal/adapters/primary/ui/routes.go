package ui

import (
	"net/http"

	"marketflow/internal/adapters/primary/ui/handlers"
)

func RegisterRoutes(handler *handlers.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /prices/latest/{symbol}", handler.LatestBySymbol)
	mux.HandleFunc("GET /prices/latest/{exchange}/{symbol}", handler.LatestBySymbolAndExchange)

	mux.HandleFunc("GET /prices/highest/{symbol}", handler.HighestBySymbol)
	mux.HandleFunc("GET /prices/highest/{exchange}/{symbol}", handler.HighestBySymbolAndExchange)

	mux.HandleFunc("GET /prices/lowest/{symbol}", handler.LowestBySymbol)
	mux.HandleFunc("GET /prices/lowest/{exchange}/{symbol}", handler.LowestBySymbolAndExchange)

	mux.HandleFunc("GET /prices/average/{symbol}", handler.AverageBySymbol)
	mux.HandleFunc("GET /prices/average/{exchange}/{symbol}", handler.AverageBySymbolAndExchange)

	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /mode/test", handler.SwitchToTestMode)
	mux.HandleFunc("POST /mode/live", handler.SwitchToLiveMode)

	return mux
}
