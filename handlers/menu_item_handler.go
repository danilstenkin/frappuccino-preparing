package handlers

import (
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"frappuccino/repositories"
	"frappuccino/utils" // Импортируем utils для проверки валидности
	"log"
	"net/http"
	"strings"
)

func CreateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var item models.MenuItem

	// Декодируем JSON из тела запроса в структуру MenuItem
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if item.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if item.Price <= 0 {
		http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
		return
	}

	// Проверка правильности размера (size) - он должен быть из перечисления item_size
	validSizes := []string{"small", "medium", "large"}
	if !utils.IsValidSize(validSizes, item.Size) {
		http.Error(w, "Invalid size", http.StatusBadRequest)
		return
	}

	// Проверяем, существуют ли все ингредиенты в инвентаре
	err = utils.ValidateIngredients(item.Ingredients)
	if err != nil {
		http.Error(w, "Ingredient validation failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}
	defer dbConn.Close()

	// Создаем новый элемент меню в базе данных
	id, err := repositories.CreateMenuItem(dbConn, item)
	if err != nil {
		http.Error(w, "Could not create menu item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Добавляем ингредиенты в таблицу menu_item_ingredients
	for _, ingredient := range item.Ingredients {
		err = repositories.AddIngredientToMenu(id, ingredient.IngredientID, ingredient.QuantityRequired)
		if err != nil {
			http.Error(w, "Failed to add ingredient to menu: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Отправляем успешный ответ с ID нового элемента
	response := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

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

func DeleteMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")

	if id == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	err := repositories.DeleteMenuItem(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Не удалось удалить элемент меню: %v", err), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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

func UpdateMenuItemHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/menu/")

	if id == "" {
		http.Error(w, "ID not found", http.StatusBadRequest)
		return
	}

	var item models.MenuItem

	// Декодируем JSON
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Простая валидация
	if item.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if item.Price <= 0 {
		http.Error(w, "Price must be greater than 0", http.StatusBadRequest)
		return
	}

	// Обновляем в БД
	err = repositories.UpdateMenuItem(id, item)
	if err != nil {
		http.Error(w, "Не удалось обновить элемент меню: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
