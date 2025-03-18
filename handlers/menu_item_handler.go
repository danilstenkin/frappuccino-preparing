package handlers

import (
	"encoding/json"
	"frappuccino/db"
	"frappuccino/models"
	"frappuccino/repositories"
	"log"
	"net/http"
)

func CreateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	var item models.MenuItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer dbConn.Close()

	id, err := repositories.CreateMenuItem(dbConn, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"id": id}
	json.NewEncoder(w).Encode(response)
}
