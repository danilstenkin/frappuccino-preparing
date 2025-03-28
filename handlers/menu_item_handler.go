package handlers

import (
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"frappuccino/repositories"
	"frappuccino/utils"
	"log"
	"net/http"
	"strings"
)

// CREATE MENU --------------------------------------------------------------
func CreateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to create menu item")

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Printf("Invalid method: %s. Expected POST.", r.Method)
		return
	}

	var item models.MenuItem

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		log.Printf("Error decoding JSON: %v", err)
		return
	}

	if item.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		log.Printf("Error: Name is required in the request body")
		return
	}

	if item.Price <= 0 {
		http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
		log.Printf("Error: Invalid price %f, must be greater than 0", item.Price)
		return
	}

	validSizes := []string{"small", "medium", "large"}
	if !utils.IsValidSize(validSizes, item.Size) {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		log.Printf("Invalid size: %s", item.Size)
		return
	}

	err = utils.ValidateIngredients(item.Ingredients)
	if err != nil {
		http.Error(w, "Ingredient validation failed: "+err.Error(), http.StatusBadRequest)
		log.Printf("Ingredient validation failed: %v", err)
		return
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer dbConn.Close()

	id, err := repositories.CreateMenuItem(dbConn, item)
	if err != nil {
		http.Error(w, "Could not create menu item: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error creating menu item: %v", err)
		return
	}

	for _, ingredient := range item.Ingredients {
		err = repositories.AddIngredientToMenu(id, ingredient.IngredientID, ingredient.QuantityRequired)
		if err != nil {
			http.Error(w, "Failed to add ingredient to menu: "+err.Error(), http.StatusInternalServerError)
			log.Printf("Failed to add ingredient ID %d to menu item ID %d: %v", ingredient.IngredientID, id, err)
			return
		}
	}

	log.Printf("Menu item created successfully with ID: %d", id)

	response := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// DELETE MENU---------------------------------------------------------------------------------

func DeleteMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")

	if id == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	err := repositories.DeleteMenuItem(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete menu item: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UPDATE -----------------------------------------------------------------------------------------
func UpdateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	const logPrefix = "[UpdateMenuItemHandler]"

	id := strings.TrimPrefix(r.URL.Path, "/menu/")
	if id == "" {
		log.Printf("%s Missing menu item ID in URL", logPrefix)
		http.Error(w, "Menu item ID is required", http.StatusBadRequest)
		return
	}

	var item models.MenuItem

	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Printf("%s Failed to decode JSON: %v", logPrefix, err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if item.Name == "" {
		log.Printf("%s Validation failed: name is missing", logPrefix)
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if item.Price <= 0 {
		log.Printf("%s Validation failed: price is non-positive (%f)", logPrefix, item.Price)
		http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
		return
	}

	validSizes := []string{"small", "medium", "large"}
	if !utils.IsValidSize(validSizes, item.Size) {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		log.Printf("Invalid size: %s", item.Size)
		return
	}

	err = utils.ValidateIngredients(item.Ingredients)
	if err != nil {
		log.Printf("%s Ingredient validation failed: %v", logPrefix, err)
		http.Error(w, "Ingredient validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%s Updating menu item ID: %s", logPrefix, id)
	err = repositories.UpdateMenuItem(id, item)
	if err != nil {
		log.Printf("%s Failed to update menu item: %v", logPrefix, err)
		http.Error(w, "Failed to update menu item: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("%s Menu item with ID %s updated successfully", logPrefix, id)
	w.WriteHeader(http.StatusOK)
}

// GET --------------------------------------------------------------------------------
func GetMenuItemsHandler(w http.ResponseWriter, r *http.Request) {
	// Шаг 1: Получаем данные из базы данных
	items, err := repositories.GetMenuItems()
	if err != nil {
		http.Error(w, "Не удалось получить элементы меню: "+err.Error(), http.StatusInternalServerError)
		log.Println("Ошибка при получении элементов меню:", err)
		return
	}

	// Шаг 2: Отправляем ответ с элементами меню в формате JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// GET BY ID -----------------------------------------------------------------------------------

func GetMenuItemsIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")

	if id == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	items, err := repositories.GetMenuItemByID(id)
	if err != nil {
		http.Error(w, "Не удалось получить элементы меню: "+err.Error(), http.StatusBadRequest)
		log.Println("Ошибка при получении элементов меню:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
