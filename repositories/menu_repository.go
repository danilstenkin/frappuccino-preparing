package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"frappuccino/db"
	"frappuccino/models"
	"log"
	"strconv"

	"github.com/lib/pq"
)

func CreateMenuItem(db *sql.DB, item models.MenuItem) (int, error) {
	// Сериализуем map[string]interface{} в JSON
	customizationOptionsJSON, err := json.Marshal(item.CustomizationOptions)
	if err != nil {
		return 0, fmt.Errorf("could not serialize customization_options: %v", err)
	}
	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return 0, fmt.Errorf("could not serialize metadata: %v", err)
	}

	// Подготовка SQL-запроса для вставки данных
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

func DeleteMenuItem(idstr string) error {
	// Преобразуем строку в int
	idint, err := strconv.Atoi(idstr)
	if err != nil {
		return fmt.Errorf("ошибка при преобразовании ID: %v", err)
	}

	// Подключаемся к базе данных
	dbConn, err := db.InitDB()
	if err != nil {
		log.Println("Не удалось подключиться к БД:", err)
		return fmt.Errorf("не удалось подключиться к базе данных: %v", err)
	}
	defer dbConn.Close()

	// Шаг 1: Удаляем зависимые записи из order_items
	deleteQuery := `DELETE FROM order_items WHERE menu_item_id = $1`
	_, err = dbConn.Exec(deleteQuery, idint)
	if err != nil {
		log.Println("Ошибка при удалении зависимых записей из order_items:", err)
		return fmt.Errorf("не удалось удалить зависимые записи из order_items: %v", err)
	}

	// Шаг 2: Удаляем элемент из menu_items
	query := `DELETE FROM menu_items WHERE id = $1`
	result, err := dbConn.Exec(query, idint)
	if err != nil {
		log.Println("Ошибка при удалении элемента из menu_items:", err)
		return fmt.Errorf("не удалось удалить элемент меню: %v", err)
	}

	// Получаем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println("Ошибка при получении количества удалённых строк:", err)
		return fmt.Errorf("ошибка при получении количества удалённых строк: %v", err)
	}

	// Если строка не была удалена
	if rowsAffected == 0 {
		return fmt.Errorf("элемент меню с ID %v не найден", idint)
	}

	return nil
}

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
