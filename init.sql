-- 1. ENUM Types
CREATE TYPE order_status AS ENUM ('pending', 'preparing', 'completed', 'canceled');
CREATE TYPE payment_method AS ENUM ('cash', 'card', 'online');
CREATE TYPE staff_role AS ENUM ('barista', 'cashier', 'manager');
CREATE TYPE item_size AS ENUM ('small', 'medium', 'large');
CREATE TYPE unit_type AS ENUM ('grams', 'ml', 'pcs');

-- 2. Customers Table
CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    preferences JSONB
);

-- 3. Orders Table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER REFERENCES customers(id),
    status order_status DEFAULT 'pending',
    special_instructions JSONB,
    total_amount NUMERIC(10,2),
    order_date TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Order Status History
CREATE TABLE order_status_history (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    status order_status,
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Menu Items
CREATE TABLE menu_items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price NUMERIC(10,2) NOT NULL,
    category TEXT[],
    allergens TEXT[],
    customization_options JSONB,
    size item_size,
    metadata JSONB
);

-- 6. Order Items
CREATE TABLE order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id INTEGER REFERENCES menu_items(id),
    quantity INTEGER NOT NULL,
    price_at_order_time NUMERIC(10,2) NOT NULL,
    customization JSONB
);

-- 7. Inventory
CREATE TABLE inventory (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    unit unit_type,
    price_per_unit NUMERIC(10,2),
    last_updated TIMESTAMPTZ DEFAULT NOW()
);

-- 8. Menu Item Ingredients (Junction)
CREATE TABLE menu_item_ingredients (
    id SERIAL PRIMARY KEY,
    menu_item_id INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    ingredient_id INTEGER REFERENCES inventory(id),
    quantity_required INTEGER NOT NULL
);

-- 9. Price History
CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    menu_item_id INTEGER REFERENCES menu_items(id) ON DELETE CASCADE,
    price NUMERIC(10,2) NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

-- 10. Inventory Transactions
CREATE TABLE inventory_transactions (
    id SERIAL PRIMARY KEY,
    inventory_id INTEGER REFERENCES inventory(id) ON DELETE CASCADE,
    change_amount INTEGER NOT NULL,
    transaction_date TIMESTAMPTZ DEFAULT NOW(),
    reason TEXT
);

-- 11. Indexes
CREATE INDEX idx_orders_customer_id ON orders(customer_id);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_menu_items_search ON menu_items USING gin (to_tsvector('english', name || ' ' || description));
CREATE INDEX idx_inventory_name ON inventory(name);

-- 12. Mock Data

-- Customers
INSERT INTO customers (name, preferences) VALUES
('Alice Brown', '{"favorite_drink": "Latte", "no_sugar": true}'),
('Bob Smith', '{"allergy": "nuts"}'),
('Charlie Green', '{}');

-- Menu Items
INSERT INTO menu_items (name, description, price, category, allergens, customization_options, size, metadata) VALUES
('Latte', 'Classic milk coffee', 4.50, ARRAY['coffee', 'hot'], ARRAY['milk'], '{"syrup": "vanilla"}', 'medium', '{"season": "winter"}'),
('Espresso', 'Strong black coffee', 3.00, ARRAY['coffee'], ARRAY[]::TEXT[], '{}', 'small', '{}'),
('Muffin', 'Chocolate muffin', 2.00, ARRAY['dessert'], ARRAY['gluten', 'eggs'], '{}', NULL, '{}');

-- Inventory
INSERT INTO inventory (name, quantity, unit, price_per_unit) VALUES
('Coffee Beans', 10000, 'grams', 0.05),
('Milk', 5000, 'ml', 0.03),
('Chocolate', 2000, 'grams', 0.10),
('Flour', 3000, 'grams', 0.02),
('Eggs', 200, 'pcs', 0.15);

-- Menu Item Ingredients
INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity_required) VALUES
(1, 1, 100),
(1, 2, 200),
(2, 1, 80),
(3, 4, 150),
(3, 5, 2),
(3, 3, 50);

-- Orders
INSERT INTO orders (customer_id, status, special_instructions, total_amount, order_date) VALUES
(1, 'completed', '{"extra_shot": true}', 9.50, NOW() - INTERVAL '2 days'),
(2, 'preparing', '{}', 5.00, NOW()),
(3, 'pending', '{"no_milk": true}', 3.00, NOW());

-- Order Items
INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_order_time, customization) VALUES
(1, 1, 2, 4.50, '{"syrup": "caramel"}'),
(2, 2, 1, 3.00, '{}'),
(3, 2, 1, 3.00, '{}');

-- Order Status History
INSERT INTO order_status_history (order_id, status, changed_at) VALUES
(1, 'pending', NOW() - INTERVAL '3 days'),
(1, 'completed', NOW() - INTERVAL '2 days'),
(2, 'pending', NOW() - INTERVAL '1 day'),
(2, 'preparing', NOW());

-- Price History
INSERT INTO price_history (menu_item_id, price, changed_at) VALUES
(1, 4.00, NOW() - INTERVAL '6 months'),
(1, 4.50, NOW() - INTERVAL '1 month'),
(2, 3.00, NOW() - INTERVAL '3 months');

-- Inventory Transactions
INSERT INTO inventory_transactions (inventory_id, change_amount, transaction_date, reason) VALUES
(1, -200, NOW() - INTERVAL '1 day', 'Order #1'),
(2, -200, NOW() - INTERVAL '1 day', 'Order #1'),
(4, -150, NOW() - INTERVAL '2 days', 'Order #3');
