package handlers

import (
	"encoding/json"
	"frappuccino/repositories"
	"net/http"
	"strings"
)

func GetOrderStatusHistoryHandler(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/order-status-history/")

	if orderID == "" {
		http.Error(w, "order_id не указан", http.StatusBadRequest)
		return
	}

	history, err := repositories.GetOrderStatusHistory(orderID)
	if err != nil {
		http.Error(w, "Ошибка при получении истории: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
