package models

type MenuItem struct {
	ID                   int                    `json:"id"`
	Name                 string                 `json:"name"`
	Description          string                 `json:"description"`
	Price                float64                `json:"price"`
	Category             []string               `json:"category"`
	Allergens            []string               `json:"allergens"`
	CustomizationOptions map[string]interface{} `json:"customization_options"`
	Size                 string                 `json:"size"`
	Metadata             map[string]interface{} `json:"metadata"`
	Ingredients          []IngredientInfo       `json:"ingredients"`
}

type IngredientInfo struct {
	IngredientID     int `json:"ingredient_id"`
	QuantityRequired int `json:"quantity_required"`
}
