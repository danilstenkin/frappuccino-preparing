package utils

import (
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"strings"
)

// Проверяет, что размер валиден (он входит в допустимый список).
func IsValidSize(validSizes []string, size string) bool {
	for _, s := range validSizes {
		if strings.ToLower(s) == strings.ToLower(size) {
			return true
		}
	}
	return false
}

func ValidateIngredients(ingredients []models.IngredientInfo) error {
	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	// Проверяем каждый ингредиент
	for _, ingredient := range ingredients {
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM inventory WHERE id = $1)`
		err := dbConn.QueryRow(query, ingredient.IngredientID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("ошибка при проверке ингредиента с ID %d: %v", ingredient.IngredientID, err)
		}
		if !exists {
			return fmt.Errorf("ингредиент с ID %d не найден в инвентаре", ingredient.IngredientID)
		}
	}

	return nil
}
