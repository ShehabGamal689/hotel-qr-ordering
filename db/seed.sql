-- Clear existing seed data to ensure determinism during container restarts
TRUNCATE TABLE order_items, orders, menu_items, rooms, properties CASCADE;

-- Insert a premium resort property
INSERT INTO properties (id, name) 
VALUES ('d290f1ee-6c54-4b01-90e6-d701748f0851', 'Aura Oasis Resort & Spa');

-- Insert rooms with unique tokens corresponding to rooms 101-105
INSERT INTO rooms (id, property_id, room_number, qr_token) VALUES 
('f71d5b3a-59b4-4b53-a5c2-9e2c608f65a1', 'd290f1ee-6c54-4b01-90e6-d701748f0851', '101', 'token_101'),
('f71d5b3a-59b4-4b53-a5c2-9e2c608f65a2', 'd290f1ee-6c54-4b01-90e6-d701748f0851', '102', 'token_102'),
('f71d5b3a-59b4-4b53-a5c2-9e2c608f65a3', 'd290f1ee-6c54-4b01-90e6-d701748f0851', '103', 'token_103'),
('f71d5b3a-59b4-4b53-a5c2-9e2c608f65a4', 'd290f1ee-6c54-4b01-90e6-d701748f0851', '104', 'token_104'),
('f71d5b3a-59b4-4b53-a5c2-9e2c608f65a5', 'd290f1ee-6c54-4b01-90e6-d701748f0851', '105', 'token_105');

-- Insert luxury room-service menu items across categories (Mains, Drinks, Desserts)
INSERT INTO menu_items (id, property_id, name, description, price, is_available, category) VALUES 
('e0f10c61-00aa-43e5-bf18-7b91932454a1', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Oasis Wagyu Burger', 'Flame-grilled 200g Wagyu beef patty, vintage white cheddar, caramelized balsamic onions, black truffle aioli on a toasted brioche bun, served with gold-dusted truffle fries.', 28.00, true, 'Mains'),
('e0f10c61-00aa-43e5-bf18-7b91932454a2', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Truffle Tagliatelle', 'Handmade egg pasta tossed in a rich parmigiano and wild black truffle butter cream, sautéed porcini mushrooms, finished with fresh microgreens and extra virgin olive oil.', 24.50, true, 'Mains'),
('e0f10c61-00aa-43e5-bf18-7b91932454a3', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Glazed Salmon Bowl', 'Pan-seared organic Atlantic salmon glazed in maple-teriyaki, warm quinoa pilaf, edamame, pickled English cucumber, avocado slices, sesame ginger dressing.', 26.00, true, 'Mains'),
('e0f10c61-00aa-43e5-bf18-7b91932454a4', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Citrus Mint Mocktail', 'Muddled fresh Persian limes, garden mint leaves, organic cucumber ribbons, premium sparkling tonic, touch of artisanal elderflower syrup.', 9.50, true, 'Drinks'),
('e0f10c61-00aa-43e5-bf18-7b91932454a5', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Cold Brew Tonic', 'Slow-dripped single-origin Ethiopian cold brew coffee poured over fever-tree Mediterranean tonic, garnished with a caramelized grapefruit wheel.', 8.00, true, 'Drinks'),
('e0f10c61-00aa-43e5-bf18-7b91932454b1', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Hibiscus Lavender Iced Tea', 'Steeped organic Egyptian hibiscus blossoms and French lavender flowers, sweetened with raw organic agave, finished with fresh blueberries and lemon.', 7.50, true, 'Drinks'),
('e0f10c61-00aa-43e5-bf18-7b91932454b2', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Warm Chocolate Lava Cake', 'Decadent 70% dark Belgian chocolate cake with a molten warm core, accompanied by a scoop of Tahitian vanilla bean gelato, fresh raspberries, and raspberry coulis.', 12.00, true, 'Desserts'),
('e0f10c61-00aa-43e5-bf18-7b91932454b3', 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Matcha Crème Brûlée', 'Silky, rich Uji matcha tea custard under a glass-like caramelized turbinado sugar shell, topped with fresh strawberries.', 11.50, true, 'Desserts');
