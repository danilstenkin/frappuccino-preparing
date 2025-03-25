package models

type OrderItem struct {
	ID               int                    `json:"id"`
	OrderID          int                    `json:"order_id"`
	MenuItemID       int                    `json:"menu_item_id"`
	Quantity         int                    `json:"quantity"`
	PriceAtOrderTime float64                `json:"price_at_order_time"`
	Customization    map[string]interface{} `json:"customization"`
}
