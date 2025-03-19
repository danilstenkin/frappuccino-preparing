package router

import (
	"frappuccino/handlers"
	"net/http"
)

func SetupRouter() {
	http.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateMenuItemHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetMenuItemsHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
