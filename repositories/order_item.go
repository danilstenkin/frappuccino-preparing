package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"strconv"
)

func GetOrderItemsByOrderID(orderIDStr string) ([]models.OrderItem, error) {
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return nil, fmt.Errorf("неверный формат order_id: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `SELECT id, order_id, menu_item_id, quantity, price_at_order_time, customization FROM order_items WHERE order_id = $1`

	rows, err := dbConn.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе order_items: %v", err)
	}
	defer rows.Close()

	var items []models.OrderItem

	for rows.Next() {
		var item models.OrderItem
		var customization sql.NullString

		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Quantity,
			&item.PriceAtOrderTime,
			&customization,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}

		if customization.Valid {
			err = json.Unmarshal([]byte(customization.String), &item.Customization)
			if err != nil {
				return nil, fmt.Errorf("не удалось распарсить кастомизацию: %v", err)
			}
		} else {
			item.Customization = nil
		}

		items = append(items, item)
	}

	return items, nil
}

func CreateOrderItem(item models.OrderItem) (int, error) {
	dbConn, err := db.InitDB()
	if err != nil {
		return 0, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	customJSON, err := json.Marshal(item.Customization)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации кастомизации: %v", err)
	}

	query := `INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order_time, customization)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`

	var id int
	err = dbConn.QueryRow(query, item.OrderID, item.MenuItemID, item.Quantity, item.PriceAtOrderTime, customJSON).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("не удалось добавить позицию в заказ: %v", err)
	}

	return id, nil
}

func DeleteOrderItem(idStr string) error {
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("неверный ID: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `DELETE FROM order_items WHERE id = $1`
	result, err := dbConn.Exec(query, idInt)
	if err != nil {
		return fmt.Errorf("ошибка при удалении позиции: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить результат удаления: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("позиция с ID %v не найдена", idInt)
	}

	return nil
}

func HasEnoughIngredients(menuItemID int, quantity int) (bool, error) {
	dbConn, err := db.InitDB()
	if err != nil {
		return false, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `
	SELECT 
		i.name,
		i.quantity AS stock_quantity,
		mii.quantity_required * $2 AS required_quantity
	FROM 
		menu_item_ingredients mii
	JOIN 
		inventory i ON mii.ingredient_id = i.id
	WHERE 
		mii.menu_item_id = $1
	`

	rows, err := dbConn.Query(query, menuItemID, quantity)
	if err != nil {
		return false, fmt.Errorf("ошибка при проверке остатков: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var stock, required int
		if err := rows.Scan(&name, &stock, &required); err != nil {
			return false, fmt.Errorf("ошибка при сканировании остатков: %v", err)
		}

		if stock < required {
			return false, fmt.Errorf("недостаточно ингредиента: %s (нужно %d, есть %d)", name, required, stock)
		}
	}

	return true, nil
}

func DeductIngredients(menuItemID int, quantity int) error {
	dbConn, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `
	SELECT 
		ingredient_id,
		quantity_required * $2 AS total_to_deduct
	FROM 
		menu_item_ingredients
	WHERE 
		menu_item_id = $1
	`

	rows, err := dbConn.Query(query, menuItemID, quantity)
	if err != nil {
		return fmt.Errorf("ошибка при получении ингредиентов: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ingredientID, total int
		if err := rows.Scan(&ingredientID, &total); err != nil {
			return fmt.Errorf("ошибка при сканировании: %v", err)
		}

		update := `UPDATE inventory SET quantity = quantity - $1 WHERE id = $2`
		_, err := dbConn.Exec(update, total, ingredientID)
		if err != nil {
			return fmt.Errorf("ошибка при списании ингредиента #%d: %v", ingredientID, err)
		}
	}

	return nil
}
