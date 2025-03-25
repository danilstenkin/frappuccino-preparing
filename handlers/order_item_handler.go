package handlers

import (
	"encoding/json"
	"frappuccino/models"
	"frappuccino/repositories"
	"net/http"
	"strings"
)

func GetOrderItemsHandler(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, "/order-items/")

	if orderID == "" {
		http.Error(w, "order_id не указан", http.StatusBadRequest)
		return
	}

	items, err := repositories.GetOrderItemsByOrderID(orderID)
	if err != nil {
		http.Error(w, "Ошибка при получении состава заказа: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var item models.OrderItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	if item.OrderID == 0 || item.MenuItemID == 0 || item.Quantity <= 0 || item.PriceAtOrderTime <= 0 {
		http.Error(w, "Неверные данные позиции", http.StatusBadRequest)
		return
	}

	ok, err := repositories.HasEnoughIngredients(item.MenuItemID, item.Quantity)
	if err != nil {
		http.Error(w, "Ошибка проверки остатков: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Недостаточно ингредиентов для этого блюда", http.StatusBadRequest)
		return
	}

	id, err := repositories.CreateOrderItem(item)
	if err != nil {
		http.Error(w, "Ошибка при добавлении позиции: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = repositories.DeductIngredients(item.MenuItemID, item.Quantity)
	if err != nil {
		http.Error(w, "Ошибка при списании ингредиентов: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func DeleteOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/order-items/")

	if id == "" {
		http.Error(w, "ID не указан", http.StatusBadRequest)
		return
	}

	err := repositories.DeleteOrderItem(id)
	if err != nil {
		http.Error(w, "Ошибка при удалении позиции: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
