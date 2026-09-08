package model

type CartItem struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Quantity    int     `json:"quantity"`
}

type Cart struct {
	UserID     int        `json:"user_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
}

type AddToCartRequest struct {
	ProductID   int     `json:"product_id" binding:"required,gt=0"`
	ProductName string  `json:"product_name"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
}

type UpdateItemQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}
