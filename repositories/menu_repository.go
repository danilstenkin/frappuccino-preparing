package repositories

import (
	"database/sql"
	"fmt"
	"frappuccino/models"

	"github.com/lib/pq" // Подключаем pq для работы с массивами
)

func CreateMenuItem(db *sql.DB, item models.MenuItem) (int, error) {
	query := `INSERT INTO menu_items (name, description, price, category) 
              VALUES ($1, $2, $3, $4) RETURNING id`

	var id int
	err := db.QueryRow(query, item.Name, item.Description, item.Price, pq.Array(item.Category)).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not insert menu item: %v", err)
	}
	return id, nil
}
