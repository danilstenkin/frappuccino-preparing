package repositories

import (
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"strconv"
)

func GetOrderStatusHistory(orderIDStr string) ([]models.OrderStatusHistory, error) {
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		return nil, fmt.Errorf("неверный формат order_id: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	query := `SELECT status, changed_at FROM order_status_history WHERE order_id = $1 ORDER BY changed_at`

	rows, err := dbConn.Query(query, orderID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе истории статусов: %v", err)
	}
	defer rows.Close()

	var history []models.OrderStatusHistory

	for rows.Next() {
		var h models.OrderStatusHistory
		err := rows.Scan(&h.Status, &h.ChangedAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании: %v", err)
		}
		history = append(history, h)
	}

	return history, nil
}
