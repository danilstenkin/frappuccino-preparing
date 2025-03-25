package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"log"
	"strconv"
)

func GetOrders() ([]models.Order, error) {
	dbConn, err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `SELECT id, customer_id, status, special_instructions, total_amount, order_date FROM orders`
	rows, err := dbConn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		var specialInstructions sql.NullString

		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.Status,
			&specialInstructions,
			&order.TotalAmount,
			&order.OrderDate,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании заказа: %v", err)
		}

		if specialInstructions.Valid {
			err = json.Unmarshal([]byte(specialInstructions.String), &order.SpecialInstructions)
			if err != nil {
				return nil, fmt.Errorf("не удалось распарсить special_instructions: %v", err)
			}
		} else {
			order.SpecialInstructions = nil
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func CreateOrder(order models.Order) (int, error) {
	dbConn, err := db.InitDB()
	if err != nil {
		return 0, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	specialInstructionsJSON, err := json.Marshal(order.SpecialInstructions)
	if err != nil {
		return 0, fmt.Errorf("ошибка сериализации special_instructions: %v", err)
	}

	query := `INSERT INTO orders (customer_id, status, special_instructions, total_amount)
			  VALUES ($1, $2, $3, $4) RETURNING id`

	var id int
	err = dbConn.QueryRow(query, order.CustomerID, order.Status, specialInstructionsJSON, order.TotalAmount).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("не удалось создать заказ: %v", err)
	}

	return id, nil
}

func GetOrderById(idStr string) (models.Order, error) {
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return models.Order{}, fmt.Errorf("ошибка при преобразовании ID: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Println("Не удалось подключиться к БД:", err)
		return models.Order{}, fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}
	defer dbConn.Close()

	query := `SELECT id, customer_id, status, special_instructions, total_amount, order_date FROM orders WHERE id = $1`

	var order models.Order
	var specialInstructions sql.NullString

	err = dbConn.QueryRow(query, idInt).Scan(
		&order.ID,
		&order.CustomerID,
		&order.Status,
		&specialInstructions,
		&order.TotalAmount,
		&order.OrderDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Order{}, fmt.Errorf("заказ с таким ID не найден")
		}
		return models.Order{}, fmt.Errorf("ошибка при выполнении запроса: %v", err)
	}

	if specialInstructions.Valid {
		err = json.Unmarshal([]byte(specialInstructions.String), &order.SpecialInstructions)
		if err != nil {
			return models.Order{}, fmt.Errorf("не удалось распарсить special_instructions: %v", err)
		}
	} else {
		order.SpecialInstructions = nil
	}

	return order, nil
}

func UpdateOrderStatus(idStr string, status string) error {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("неверный формат ID: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	// 1. Обновляем статус в таблице orders
	query := `UPDATE orders SET status = $1 WHERE id = $2`
	result, err := dbConn.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении заказа: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("не удалось получить результат обновления: %v", err)
	}
	if affected == 0 {
		return fmt.Errorf("заказ с ID %v не найден", id)
	}

	// 2. Записываем в историю
	err = CreateOrderStatusHistory(dbConn, id, status)
	if err != nil {
		return fmt.Errorf("статус обновлён, но не удалось сохранить историю: %v", err)
	}

	return nil
}

func CreateOrderStatusHistory(dbConn *sql.DB, orderID int, status string) error {
	query := `INSERT INTO order_status_history (order_id, status) VALUES ($1, $2)`
	_, err := dbConn.Exec(query, orderID, status)
	if err != nil {
		return fmt.Errorf("ошибка при записи в историю статусов: %v", err)
	}
	return nil
}
