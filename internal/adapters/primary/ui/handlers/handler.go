package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"marketflow/internal/core/service"
)

type Handler struct {
	service          *service.Stats
	switchToTestMode func() error
	switchToLiveMode func() error
	healthCheck      func() []byte
}

func WithTestModeSwitch(f func() error, h *Handler) {
	h.switchToTestMode = f
}

func WithLiveModeSwitch(f func() error, h *Handler) {
	h.switchToLiveMode = f
}

func WithHealthCheck(f func() []byte, h *Handler) {
	h.healthCheck = f
}

type PriceResponse struct {
	PairName     string    `json:"pair_name"`
	Exchange     string    `json:"exchange"`
	Price        *float64  `json:"price,omitempty"`
	AveragePrice *float64  `json:"average_price,omitempty"`
	MinPrice     *float64  `json:"min_price,omitempty"`
	MaxPrice     *float64  `json:"max_price,omitempty"`
	Period       string    `json:"period,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

type SystemResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	setCORSHeaders(w)
	writeJSONResponse(w, ErrorResponse{Error: message}, statusCode)
}

func NewHandler(service *service.Stats) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) LatestBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")

	price, err := h.service.GetLatestPrice(r.Context(), "global", symbol)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  "global",
		Price:     &price,
		Timestamp: time.Now(),
	}

	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) LatestBySymbolAndExchange(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	exchange := r.PathValue("exchange")

	price, err := h.service.GetLatestPrice(r.Context(), exchange, symbol)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  exchange,
		Price:     &price,
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) HighestBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetHighestPrice(r.Context(), "global", symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  "global",
		MaxPrice:  &price,
		Period:    period,
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) HighestBySymbolAndExchange(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	exchange := r.PathValue("exchange")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetHighestPrice(r.Context(), exchange, symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  exchange,
		MaxPrice:  &price,
		Period:    period,
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) LowestBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetLowestPrice(r.Context(), "global", symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  "global",
		MinPrice:  &price,
		Period:    period,
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) LowestBySymbolAndExchange(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	exchange := r.PathValue("exchange")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetLowestPrice(r.Context(), exchange, symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:  symbol,
		Exchange:  exchange,
		MinPrice:  &price,
		Period:    period,
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) AverageBySymbol(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetAveragePrice(r.Context(), "global", symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:     symbol,
		Exchange:     "global",
		AveragePrice: &price,
		Period:       period,
		Timestamp:    time.Now(),
	}

	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) AverageBySymbolAndExchange(w http.ResponseWriter, r *http.Request) {
	symbol := r.PathValue("symbol")
	exchange := r.PathValue("exchange")
	period := r.URL.Query().Get("period")

	price, err := h.service.GetAveragePrice(r.Context(), exchange, symbol, period)
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := PriceResponse{
		PairName:     symbol,
		Exchange:     exchange,
		AveragePrice: &price,
		Period:       period,
		Timestamp:    time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := h.healthCheck()
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SwitchToTestMode(w http.ResponseWriter, r *http.Request) {
	err := h.switchToTestMode()
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := SystemResponse{
		Status:    "test",
		Message:   "Switched to test mode",
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}

func (h *Handler) SwitchToLiveMode(w http.ResponseWriter, r *http.Request) {
	err := h.switchToLiveMode()
	if err != nil {
		writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	response := SystemResponse{
		Status:    "live",
		Message:   "Switched to live mode",
		Timestamp: time.Now(),
	}
	writeJSONResponse(w, response, http.StatusOK)
}
