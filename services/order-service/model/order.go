package model

import (
	"strings"
	"time"
)

const (
	StatusPending   = "PENDING"
	StatusConfirmed = "CONFIRMED"
	StatusCancelled = "CANCELLED"
)

type OrderItem struct {
	ID          int     `json:"id"`
	OrderID     int     `json:"order_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type ShippingAddress struct {
	ID            int       `json:"id,omitempty"`
	OrderID       int       `json:"order_id,omitempty"`
	RecipientName string    `json:"recipient_name"`
	PhoneNumber   string    `json:"phone_number"`
	AddressLine1  string    `json:"address_line1"`
	AddressLine2  string    `json:"address_line2,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

func (a *ShippingAddress) FormatSummary() string {
	if a == nil {
		return ""
	}
	parts := []string{}
	if a.RecipientName != "" {
		parts = append(parts, a.RecipientName)
	}
	if a.AddressLine1 != "" {
		parts = append(parts, a.AddressLine1)
	}
	if a.AddressLine2 != "" {
		parts = append(parts, a.AddressLine2)
	}
	cityStateZip := []string{}
	if a.City != "" {
		cityStateZip = append(cityStateZip, a.City)
	}
	if a.State != "" {
		cityStateZip = append(cityStateZip, a.State)
	}
	if a.PostalCode != "" {
		cityStateZip = append(cityStateZip, a.PostalCode)
	}
	if len(cityStateZip) > 0 {
		parts = append(parts, strings.Join(cityStateZip, ", "))
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}
	return strings.Join(parts, ", ")
}

type Order struct {
	ID              int              `json:"id"`
	UserID          int              `json:"user_id"`
	Status          string           `json:"status"`
	TotalAmount     float64          `json:"total_amount"`
	ShippingAddress string           `json:"shipping_address,omitempty"`
	Address         *ShippingAddress `json:"address,omitempty"`
	Items           []OrderItem      `json:"items"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type OrderItemRequest struct {
	ProductID   int     `json:"product_id" binding:"required,gt=0"`
	Quantity    int     `json:"quantity" binding:"required,gt=0"`
	Price       float64 `json:"price"`
	ProductName string  `json:"product_name"`
}

type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
	ShippingAddress string             `json:"shipping_address,omitempty"`
	Address         *ShippingAddress   `json:"address,omitempty"`
}
