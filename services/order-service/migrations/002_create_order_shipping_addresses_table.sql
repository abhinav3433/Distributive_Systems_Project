-- Migration 002: Create order_shipping_addresses table
-- Stores immutable address snapshots per order (1-to-1 relationship with orders)
-- Safe for databases with existing orders (no data overwritten or corrupted)

CREATE TABLE IF NOT EXISTS order_shipping_addresses (
    id SERIAL PRIMARY KEY,
    order_id INT UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    recipient_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50) NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(50) NOT NULL,
    country VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_order_shipping_addresses_order_id ON order_shipping_addresses(order_id);
