package handlers

import (
	"encoding/json"
	"frappuccino/repositories"
	"log"
	"net/http"
)

func GetInventoryHandler(w http.ResponseWriter, r *http.Request) {
	items, err := repositories.GetInventoryItems()
	if err != nil {
		http.Error(w, "Не удалось получить инвентарь: "+err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка получения инвентаря:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
