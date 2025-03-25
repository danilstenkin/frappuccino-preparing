package models

import "time"

type OrderStatusHistory struct {
	Status    string    `json:"status"`
	ChangedAt time.Time `json:"changed_at"`
}
