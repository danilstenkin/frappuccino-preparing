package handlers

import (
	"encoding/json"
	"frappuccino/models"
	"frappuccino/repositories"
	"net/http"
	"strings"
)

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	orders, err := repositories.GetOrders()
	if err != nil {
		http.Error(w, "Ошибка при получении заказов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var order models.Order

	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	// Простейшая валидация
	if order.CustomerID == 0 || order.TotalAmount <= 0 {
		http.Error(w, "Неверные данные заказа", http.StatusBadRequest)
		return
	}

	id, err := repositories.CreateOrder(order)
	if err != nil {
		http.Error(w, "Ошибка при создании заказа: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func GetOrderByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")

	if id == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	order, err := repositories.GetOrderById(id)
	if err != nil {
		http.Error(w, "Ошибка при получении заказа: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/orders/")

	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}

	// Ожидаем {"status": "completed"}
	var data struct {
		Status string `json:"status"`
	}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Status == "" {
		http.Error(w, "Неверный JSON или пустой статус", http.StatusBadRequest)
		return
	}

	err = repositories.UpdateOrderStatus(id, data.Status)
	if err != nil {
		http.Error(w, "Ошибка при обновлении: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Статус обновлён и история записана"}`))
}
