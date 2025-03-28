package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"frappuccino/utils"
	"log"
	"strconv"

	"github.com/lib/pq"
)

// CREATE ---------------------------------------------------------------------------
func CreateMenuItem(db *sql.DB, item models.MenuItem) (int, error) {
	customizationOptionsJSON, err := json.Marshal(item.CustomizationOptions)
	if err != nil {
		return 0, fmt.Errorf("could not serialize customization_options: %v", err)
	}

	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return 0, fmt.Errorf("could not serialize metadata: %v", err)
	}

	query := `INSERT INTO menu_items (name, description, price, category, allergens, customization_options, size, metadata) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	var id int

	err = db.QueryRow(query, item.Name, item.Description, item.Price, pq.Array(item.Category), pq.Array(item.Allergens),
		customizationOptionsJSON, item.Size, metadataJSON).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("could not insert menu item: %v", err)
	}
	return id, nil
}

// DELETE --------------------------------------------------------------------------------------
func DeleteMenuItem(idstr string) error {
	const logPrefix = "[DeleteMenuItem]"

	idint, err := strconv.Atoi(idstr)
	if err != nil {
		log.Printf("%s incorrect ID: %v", logPrefix, err)
		return fmt.Errorf("Error converting ID: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Printf("%s Failed to connect to DB: %v", logPrefix, err)
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	log.Printf("%s Connection to the database was successful", logPrefix)

	if err := DeleteMenuItemDependencies(dbConn, idint); err != nil {
		return err
	}

	log.Printf("%s Remove item from menu_items with ID= %d", logPrefix, idint)
	query := `DELETE FROM menu_items WHERE id = $1`
	result, err := dbConn.Exec(query, idint)
	if err != nil {
		log.Printf("%s Error while deleting from menu_items: %v", logPrefix, err)
		return fmt.Errorf("failed to delete menu item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("%s Error getting number of deleted rows: %v", logPrefix, err)
		return fmt.Errorf("error getting number of deleted rows: %v", err)
	}
	if rowsAffected == 0 {
		log.Printf("%s Menu item with ID %d not found", logPrefix, idint)
		return fmt.Errorf("menu item with ID %v not found", idint)
	}

	log.Printf("%s Menu item with ID %d successfully removed", logPrefix, idint)
	return nil
}

// UPDATE ------------------------------------------------------------------
func UpdateMenuItem(idStr string, item models.MenuItem) error {
	const logPrefix = "[UpdateMenuItem]"

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("%s Invalid ID format: %v", logPrefix, err)
		return fmt.Errorf("invalid ID format: %v", err)
	}

	dbConn, err := db.InitDB()
	if err != nil {
		log.Printf("%s Failed to connect to DB: %v", logPrefix, err)
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	defer dbConn.Close()

	if err := utils.ValidateIngredients(item.Ingredients); err != nil {
		log.Printf("%s Ingredient validation failed: %v", logPrefix, err)
		return fmt.Errorf("ingredient validation failed: %v", err)
	}

	customizationOptionsJSON, err := json.Marshal(item.CustomizationOptions)
	if err != nil {
		return fmt.Errorf("could not serialize customization_options: %v", err)
	}
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return fmt.Errorf("could not serialize metadata: %v", err)
	}

	query := `UPDATE menu_items 
		SET name=$1, description=$2, price=$3, category=$4, allergens=$5, 
		    customization_options=$6, size=$7, metadata=$8 
		WHERE id=$9`
	result, err := dbConn.Exec(query,
		item.Name, item.Description, item.Price,
		pq.Array(item.Category), pq.Array(item.Allergens),
		customizationOptionsJSON, item.Size, metadataJSON, idInt,
	)
	if err != nil {
		log.Printf("%s Failed to update menu item: %v", logPrefix, err)
		return fmt.Errorf("failed to update menu item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("menu item with ID %d not found", idInt)
	}

	if err := DeleteMenuItemDependencies(dbConn, idInt); err != nil {
		log.Printf("%s Failed to delete dependencies: %v", logPrefix, err)
		return err
	}

	for _, ingredient := range item.Ingredients {
		if err := AddIngredientToMenu(idInt, ingredient.IngredientID, ingredient.QuantityRequired); err != nil {
			log.Printf("%s Failed to add ingredient ID %d: %v", logPrefix, ingredient.IngredientID, err)
			return err
		}
	}

	log.Printf("%s Menu item ID %d updated successfully", logPrefix, idInt)
	return nil
}

// GET -----------------------------------------------------------------------------------
func GetMenuItems() ([]models.MenuItem, error) {
	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}
	defer dbConn.Close()

	// Запрашиваем все данные из таблицы menu_items
	rows, err := dbConn.Query("SELECT id, name, description, price, category, allergens, customization_options, size, metadata FROM menu_items")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить элементы меню: %v", err)
	}
	defer rows.Close()

	var items []models.MenuItem

	// Проходим по строкам результата и сканируем данные
	for rows.Next() {
		var item models.MenuItem
		var category []byte                     // Для сканирования поля TEXT[] как []byte
		var allergens []byte                    // Для сканирования поля TEXT[] как []byte
		var customizationOptions sql.NullString // Для обработки поля customization_options
		var metadata sql.NullString             // Для обработки поля metadata
		var size sql.NullString

		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &category, &allergens, &customizationOptions, &size, &metadata)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании строки: %v", err)
		}

		// Преобразуем []byte category в []string с помощью pq.Array
		if category != nil {
			err = pq.Array(&item.Category).Scan(category)
			if err != nil {
				return nil, fmt.Errorf("ошибка при сканировании категории: %v", err)
			}
		}

		// Преобразуем []byte allergens в []string с помощью pq.Array
		if allergens != nil {
			err = pq.Array(&item.Allergens).Scan(allergens)
			if err != nil {
				return nil, fmt.Errorf("ошибка при сканировании аллергенов: %v", err)
			}
		}

		// Обработка customizationOptions, если оно не NULL
		if customizationOptions.Valid {
			err = json.Unmarshal([]byte(customizationOptions.String), &item.CustomizationOptions)
			if err != nil {
				return nil, fmt.Errorf("ошибка при сканировании customization_options: %v", err)
			}
		} else {
			item.CustomizationOptions = nil
		}

		// Обработка metadata, если оно не NULL
		if metadata.Valid {
			err = json.Unmarshal([]byte(metadata.String), &item.Metadata)
			if err != nil {
				return nil, fmt.Errorf("ошибка при сканировании metadata: %v", err)
			}
		} else {
			item.Metadata = nil
		}

		// Обработка поля size
		if size.Valid {
			item.Size = size.String
		} else {
			item.Size = "" // Если size NULL, то можно установить значение по умолчанию
		}

		// Добавляем элемент в слайс
		items = append(items, item)
	}

	// Проверка на ошибки при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по строкам: %v", err)
	}

	return items, nil
}

// GET BY ID ------------------------------------------------------------------------------
func GetMenuItemByID(idstr string) ([]models.MenuItem, error) {
	// Преобразуем строку в int
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при преобразовании ID: %v", err)
	}

	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		log.Println("Не удалось подключиться к БД:", err)
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}
	defer dbConn.Close()

	// Запрос для получения элемента меню по ID
	query := `SELECT id, name, description, price, category, allergens, customization_options, size, metadata 
			  FROM menu_items WHERE id = $1`

	var item models.MenuItem
	var category []byte
	var allergens []byte
	var customizationOptions sql.NullString
	var metadata sql.NullString
	var size sql.NullString

	// Выполняем запрос
	err = dbConn.QueryRow(query, idint).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&category,
		&allergens,
		&customizationOptions,
		&size,
		&metadata,
	)

	if err == sql.ErrNoRows {
		// Если записи с таким ID нет
		return nil, fmt.Errorf("элемент меню с таким ID не найден")
	} else if err != nil {
		return nil, fmt.Errorf("ошибка при запросе элемента меню: %v", err)
	}

	// Преобразуем []byte category в []string с помощью pq.Array
	if category != nil {
		err = pq.Array(&item.Category).Scan(category)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании категории: %v", err)
		}
	}

	// Преобразуем []byte allergens в []string с помощью pq.Array
	if allergens != nil {
		err = pq.Array(&item.Allergens).Scan(allergens)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании аллергенов: %v", err)
		}
	}

	// Обработка customizationOptions, если оно не NULL
	if customizationOptions.Valid {
		err = json.Unmarshal([]byte(customizationOptions.String), &item.CustomizationOptions)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании customization_options: %v", err)
		}
	} else {
		item.CustomizationOptions = nil
	}

	// Обработка metadata, если оно не NULL
	if metadata.Valid {
		err = json.Unmarshal([]byte(metadata.String), &item.Metadata)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании metadata: %v", err)
		}
	} else {
		item.Metadata = nil
	}

	// Обработка поля size
	if size.Valid {
		item.Size = size.String
	} else {
		item.Size = "" // Если size NULL, то можно установить значение по умолчанию
	}

	// Возвращаем элемент меню в виде слайса (т.к. мы ожидаем слайс в API)
	items := []models.MenuItem{item}
	return items, nil
}
