package router

import (
	"frappuccino/handlers"
	"net/http"
)

func SetupRouter() {
	http.HandleFunc("/menu", handlers.CreateMenuItemHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
