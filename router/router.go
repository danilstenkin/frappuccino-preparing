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

	http.HandleFunc("/menu/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handlers.DeleteMenuItemHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetMenuItemsIDHandler(w, r)
		}
	})

	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetInventoryHandler(w, r)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
