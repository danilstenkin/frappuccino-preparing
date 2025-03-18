package models

type MenuItem struct {
	ID                   int         `json:"id"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	Price                float64     `json:"price"`
	Category             []string    `json:"category"` // Используем срез строк
	Allergens            []string    `json:"allergens"`
	CustomizationOptions interface{} `json:"customization_options"` // Если тебе нужно работать с JSONB
	Size                 string      `json:"size"`
	Metadata             interface{} `json:"metadata"` // Для хранения JSON
}
