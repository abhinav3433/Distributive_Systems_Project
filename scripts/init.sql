

CREATE DATABASE auth_db;
CREATE DATABASE product_db;
CREATE DATABASE order_db;
CREATE DATABASE payment_db;

\c auth_db;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

\c product_db;

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price NUMERIC(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    category VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO products (id, name, description, price, stock, category) VALUES
(1, 'ProBook Laptop 15"', 'High-performance laptop with 16GB RAM and 512GB SSD', 1299.99, 25, 'Electronics'),
(2, 'UltraPhone 14 Pro', 'Flagship smartphone with triple camera system and OLED display', 999.00, 40, 'Electronics'),
(3, 'NoiseCancelling Headphones', 'Wireless over-ear headphones with active noise cancellation', 249.50, 60, 'Audio'),
(4, 'Mechanical RGB Keyboard', 'Tactile mechanical keyboard with customizable RGB backlighting', 89.99, 100, 'Accessories'),
(5, 'Wireless Ergonomic Mouse', 'Precision optical mouse with multi-device bluetooth pairing', 49.99, 120, 'Accessories'),
(6, '4K UltraHD Monitor 27"', 'IPS display panel with 144Hz refresh rate and HDR400', 399.99, 15, 'Electronics'),
(7, 'SmartWatch Pro V2', 'Fitness tracking smartwatch with heart rate & ECG sensors', 199.95, 50, 'Wearables'),
(8, 'OctaTab 11 Tablet', '11-inch tablet with stylus support and all-day battery life', 499.00, 30, 'Electronics'),
(9, 'Mirrorless 4K Camera', 'Professional digital camera with 24MP sensor and 4K video recording', 849.99, 10, 'Photography'),
(10, 'NextGen Gaming Console 1TB', 'Ultra-fast SSD gaming console supporting up to 120 FPS output', 499.99, 20, 'Gaming'),
(11, 'SlimBook Air 13"', 'Ultra-lightweight laptop with M-series silicon and 18-hour battery', 1099.00, 20, 'Electronics'),
(12, 'DeskStation Mini PC', 'Compact desktop computer with 32GB RAM and dual 4K display output', 749.50, 18, 'Electronics'),
(13, 'True Wireless Pro Earbuds', 'In-ear buds with spatial audio, transparency mode, and wireless charging case', 159.99, 85, 'Audio'),
(14, 'Hi-Fi Studio Monitor Speakers', 'Pair of 5-inch powered desktop reference monitors for precision audio', 299.00, 22, 'Audio'),
(15, 'Rugged Waterproof Bluetooth Speaker', '360-degree portable outdoor speaker with 24-hour battery and deep bass', 79.99, 70, 'Audio'),
(16, '65W GaN USB-C Fast Charger', 'Ultra-compact 3-port wall charger for laptops, tablets, and phones', 39.99, 150, 'Accessories'),
(17, 'Braided Thunderbolt 4 Cable 2m', '40Gbps high-speed data transfer and 100W power delivery cable', 29.50, 200, 'Accessories'),
(18, 'Ergonomic Aluminum Laptop Stand', 'Adjustable height cooling stand crafted from anodized aerospace aluminum', 45.00, 65, 'Accessories'),
(19, 'Active Fitness Tracker Band', 'Lightweight waterproof fitness band with sleep tracking and SpO2 monitor', 59.99, 90, 'Wearables'),
(20, 'GPS Adventure Smartwatch', 'Rugged outdoor multi-sport watch with topographic maps and solar charging', 349.00, 25, 'Wearables'),
(21, 'Professional Carbon Fiber Tripod', 'Lightweight 65-inch camera tripod with 360-degree fluid ball head', 139.99, 35, 'Photography'),
(22, '50mm f/1.8 Prime Portrait Lens', 'Fast aperture portrait lens with beautiful bokeh and silent autofocus', 219.00, 28, 'Photography'),
(23, 'Studio Bi-Color LED Ring Light', '18-inch dimmable ring light kit with phone holder and light stand', 69.99, 55, 'Photography'),
(24, 'Wireless Pro Game Controller', 'Ergonomic wireless controller with hall-effect triggers and rear paddles', 69.99, 75, 'Gaming'),
(25, '7.1 Surround Gaming Headset', 'Immersive headset with detachable noise-cancelling boom mic and memory foam', 89.99, 60, 'Gaming'),
(26, 'RGB Extended Gaming Mousepad', 'Micro-textured waterproof cloth gaming mat with 14 RGB lighting modes', 24.99, 110, 'Gaming')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price = EXCLUDED.price,
    stock = EXCLUDED.stock,
    category = EXCLUDED.category,
    updated_at = NOW();

SELECT setval('products_id_seq', (SELECT GREATEST(MAX(id), 1) FROM products));
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name ON products(name);

\c order_db;

CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    total_amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id) ON DELETE CASCADE,
    product_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    quantity INT NOT NULL
);

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

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_shipping_addresses_order_id ON order_shipping_addresses(order_id);

\c payment_db;

CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    order_id INT UNIQUE NOT NULL,
    user_id INT NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payment_method VARCHAR(50) DEFAULT 'CREDIT_CARD',
    transaction_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
