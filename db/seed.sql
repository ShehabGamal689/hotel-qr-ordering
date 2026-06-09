-- Insert dummy properties
INSERT INTO properties (id, name) VALUES 
('11111111-1111-1111-1111-111111111111', 'The Grand Oasis Resort'),
('22222222-2222-2222-2222-222222222222', 'Boutique Lodge')
ON CONFLICT DO NOTHING;

-- Insert an admin user for The Grand Oasis Resort (password: password123)
-- bcrypt hash for password123: $2a$10$uT8n/ma8vHvWRicUoWH4IuaZDn1i34hDuUleyPNCw7x9EcijQStu2
INSERT INTO users (id, property_id, email, password_hash) VALUES 
('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'admin@grandoasis.com', '$2a$10$uT8n/ma8vHvWRicUoWH4IuaZDn1i34hDuUleyPNCw7x9EcijQStu2')
ON CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- Insert Property Services for The Grand Oasis Resort (all enabled)
INSERT INTO property_services (property_id, service_type, is_enabled) VALUES 
('11111111-1111-1111-1111-111111111111', 'fnb', TRUE),
('11111111-1111-1111-1111-111111111111', 'housekeeping', TRUE),
('11111111-1111-1111-1111-111111111111', 'laundry', TRUE),
('11111111-1111-1111-1111-111111111111', 'maintenance', TRUE),
('11111111-1111-1111-1111-111111111111', 'concierge', TRUE)
ON CONFLICT DO NOTHING;

-- Insert dummy rooms for The Grand Oasis Resort
INSERT INTO rooms (id, property_id, room_number, qr_token) VALUES 
('44444444-4444-4444-4444-444444444441', '11111111-1111-1111-1111-111111111111', '101', 'token_101_grand'),
('44444444-4444-4444-4444-444444444442', '11111111-1111-1111-1111-111111111111', '102', 'token_102_grand')
ON CONFLICT DO NOTHING;

-- Insert polymorphic Catalog Items
-- 1. Food & Beverage Items (Requires images, categories)
INSERT INTO catalog_items (id, property_id, service_type, name, description, price, is_available, attributes) VALUES 
('55555555-5555-5555-5555-555555555551', '11111111-1111-1111-1111-111111111111', 'fnb', 'Wagyu Beef Burger', 'Premium wagyu patty with truffle mayo, aged cheddar, and brioche bun.', 28.50, TRUE, '{"category": "Mains", "image_url": "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?auto=format&fit=crop&w=500&q=60"}'),
('55555555-5555-5555-5555-555555555552', '11111111-1111-1111-1111-111111111111', 'fnb', 'Artisan Iced Latte', 'Cold-brewed espresso over milk and artisanal ice cubes.', 8.00, TRUE, '{"category": "Drinks", "image_url": "https://images.unsplash.com/photo-1517701550927-30cf0b6d2e51?auto=format&fit=crop&w=500&q=60"}')
ON CONFLICT DO NOTHING;

-- 2. Housekeeping Items (Zero cost, quantity matters)
INSERT INTO catalog_items (id, property_id, service_type, name, description, price, is_available, attributes) VALUES 
('55555555-5555-5555-5555-555555555553', '11111111-1111-1111-1111-111111111111', 'housekeeping', 'Extra Towels', 'Request fresh Egyptian cotton towels.', 0.00, TRUE, '{"category": "Amenities"}'),
('55555555-5555-5555-5555-555555555554', '11111111-1111-1111-1111-111111111111', 'housekeeping', 'Make Up Room', 'Request full room cleaning service.', 0.00, TRUE, '{"category": "Services"}')
ON CONFLICT DO NOTHING;

-- 3. Laundry Items (Has service options)
INSERT INTO catalog_items (id, property_id, service_type, name, description, price, is_available, attributes) VALUES 
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'laundry', 'Dress Shirt', 'Standard cotton or linen dress shirt.', 15.00, TRUE, '{"category": "Garments", "options": ["Wash & Iron", "Dry Clean", "Iron Only"]}'),
('55555555-5555-5555-5555-555555555556', '11111111-1111-1111-1111-111111111111', 'laundry', 'Trousers', 'Formal or casual trousers.', 18.00, TRUE, '{"category": "Garments", "options": ["Wash & Iron", "Dry Clean", "Iron Only"]}')
ON CONFLICT DO NOTHING;

-- 4. Maintenance Items
INSERT INTO catalog_items (id, property_id, service_type, name, description, price, is_available, attributes) VALUES 
('55555555-5555-5555-5555-555555555557', '11111111-1111-1111-1111-111111111111', 'maintenance', 'AC Repair', 'Air conditioning unit is malfunctioning.', 0.00, TRUE, '{"category": "HVAC"}'),
('55555555-5555-5555-5555-555555555558', '11111111-1111-1111-1111-111111111111', 'maintenance', 'Plumbing Issue', 'Bathroom or sink plumbing problem.', 0.00, TRUE, '{"category": "Plumbing"}')
ON CONFLICT DO NOTHING;

-- 5. Concierge Items
INSERT INTO catalog_items (id, property_id, service_type, name, description, price, is_available, attributes) VALUES 
('55555555-5555-5555-5555-555555555559', '11111111-1111-1111-1111-111111111111', 'concierge', 'Wake-up Call', 'Request a morning wake-up call.', 0.00, TRUE, '{"category": "Requests", "requires_time": true}'),
('55555555-5555-5555-5555-555555555560', '11111111-1111-1111-1111-111111111111', 'concierge', 'Luggage Assistance', 'Request bellboy for luggage collection.', 0.00, TRUE, '{"category": "Services"}')
ON CONFLICT DO NOTHING;
