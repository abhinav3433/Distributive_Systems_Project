package events

type OrderCreatedEvent struct {
	EventType string  `json:"event_type"`
	OrderID   int     `json:"order_id"`
	UserID    int     `json:"user_id"`
	Amount    float64 `json:"amount"`
}
