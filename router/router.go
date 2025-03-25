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
		} else if r.Method == http.MethodPut {
			handlers.UpdateMenuItemHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateInventoryHandler(w, r)
		} else if r.Method == http.MethodGet {
			handlers.GetInventoryHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/inventory/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetInventoryByIDHandler(w, r)
		} else if r.Method == http.MethodPut {
			handlers.UpdateInventoryHandler(w, r)
		} else if r.Method == http.MethodDelete {
			handlers.DeleteInventoryHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetOrdersHandler(w, r)
		} else if r.Method == http.MethodPost {
			handlers.CreateOrderHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetOrderByIDHandler(w, r)
		} else if r.Method == http.MethodPut {
			handlers.UpdateOrderHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/order-items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handlers.CreateOrderItemHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/order-items/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.GetOrderItemsHandler(w, r)
		} else if r.Method == http.MethodDelete {
			handlers.DeleteOrderItemHandler(w, r)
		} else {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
