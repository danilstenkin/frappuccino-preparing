// Файл: main.go
package main

import (
	"frappuccino/router"
	"log"
)

func main() {
	// Настроим маршруты
	router.SetupRouter()

	// Сервер теперь слушает на всех интерфейсах, а не только на localhost
	log.Println("Server is running on http://0.0.0.0:8080")
}
