package repositories

import (
	"database/sql"
	"fmt"
	"frappuccino/db"
	"log"
)

func AddIngredientToMenu(menuItemID int, ingredientID int, quantityRequired int) error {
	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer dbConn.Close()

	// Добавляем ингредиент в menu_item_ingredients
	query := `INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity_required) 
			  VALUES ($1, $2, $3)`
	_, err = dbConn.Exec(query, menuItemID, ingredientID, quantityRequired)
	if err != nil {
		return fmt.Errorf("error inserting ingredient into menu_item_ingredients: %v", err)
	}

	return nil
}

func DeleteMenuItemDependencies(dbConn *sql.DB, menuItemID int) error {
	const logPrefix = "[DeleteMenuItemDependencies]"

	log.Printf("%s Remove from order_items for menu_item_id = %d", logPrefix, menuItemID)
	deleteOrderItemsQuery := `DELETE FROM order_items WHERE menu_item_id = $1`
	if _, err := dbConn.Exec(deleteOrderItemsQuery, menuItemID); err != nil {
		log.Printf("%s Error while deleting from order_items: %v", logPrefix, err)
		return fmt.Errorf("error removing dependencies from order_items: %v", err)
	}

	log.Printf("%s Remove from menu_item_ingredients for menu_item_id = %d", logPrefix, menuItemID)
	deleteIngredientsQuery := `DELETE FROM menu_item_ingredients WHERE menu_item_id = $1`
	if _, err := dbConn.Exec(deleteIngredientsQuery, menuItemID); err != nil {
		log.Printf("%s Error deleting from menu_item_ingredients: %v", logPrefix, err)
		return fmt.Errorf("error removing dependencies from menu_item_ingredients: %v", err)
	}

	return nil
}
