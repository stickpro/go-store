-- Seed data for computer electronics store
-- This file contains test data for manufacturers, categories, attribute groups, attributes, and products

-- Clean up existing data (optional - comment out if you want to preserve existing data)
-- TRUNCATE TABLE attribute_products, attributes, attribute_groups, products, manufacturers, categories CASCADE;

-- ============================================
-- MANUFACTURERS
-- ============================================
INSERT INTO manufacturers (id, name, slug, description, meta_title, meta_h1, is_enable)
VALUES ('11111111-1111-1111-1111-111111111101', 'Intel', 'intel', 'Leading manufacturer of processors and chipsets',
        'Intel Processors and Components', 'Intel Products', true),
       ('11111111-1111-1111-1111-111111111102', 'AMD', 'amd', 'Advanced Micro Devices - processors and graphics cards',
        'AMD Processors and Graphics Cards', 'AMD Products', true),
       ('11111111-1111-1111-1111-111111111103', 'NVIDIA', 'nvidia', 'Graphics processing units and AI computing',
        'NVIDIA Graphics Cards', 'NVIDIA Products', true),
       ('11111111-1111-1111-1111-111111111104', 'ASUS', 'asus', 'Motherboards, graphics cards, and computer components',
        'ASUS Computer Components', 'ASUS Products', true),
       ('11111111-1111-1111-1111-111111111105', 'MSI', 'msi',
        'Micro-Star International - motherboards and graphics cards', 'MSI Gaming Components', 'MSI Products', true),
       ('11111111-1111-1111-1111-111111111106', 'Gigabyte', 'gigabyte',
        'Motherboards, graphics cards, and PC components', 'Gigabyte PC Components', 'Gigabyte Products', true),
       ('11111111-1111-1111-1111-111111111107', 'Corsair', 'corsair',
        'Memory modules, power supplies, and cooling systems', 'Corsair Memory and Components', 'Corsair Products',
        true),
       ('11111111-1111-1111-1111-111111111108', 'Kingston', 'kingston', 'Memory modules and storage solutions',
        'Kingston Memory Solutions', 'Kingston Products', true),
       ('11111111-1111-1111-1111-111111111109', 'Samsung', 'samsung', 'SSDs, memory, and storage devices',
        'Samsung Storage Solutions', 'Samsung Products', true),
       ('11111111-1111-1111-1111-111111111110', 'Western Digital', 'western-digital',
        'Hard drives and storage solutions', 'Western Digital Storage', 'WD Products', true)

ON CONFLICT (id) DO NOTHING;

-- ============================================
-- CATEGORIES
-- ============================================
INSERT INTO categories (id, parent_id, name, slug, description, meta_title, meta_h1, is_enable)
VALUES
-- Main categories
('22222222-2222-2222-2222-222222222201', NULL, 'Processors', 'processors', 'CPUs for desktop and server systems',
 'Computer Processors', 'Processors (CPUs)', true),
('22222222-2222-2222-2222-222222222202', NULL, 'Graphics Cards', 'graphics-cards',
 'Video cards for gaming and professional work', 'Graphics Cards', 'Graphics Cards (GPUs)', true),
('22222222-2222-2222-2222-222222222203', NULL, 'Motherboards', 'motherboards', 'System boards for building PCs',
 'Motherboards', 'Motherboards', true),
('22222222-2222-2222-2222-222222222204', NULL, 'Memory (RAM)', 'memory-ram', 'RAM modules for computers',
 'Computer Memory', 'Memory (RAM)', true),
('22222222-2222-2222-2222-222222222205', NULL, 'Storage', 'storage', 'Hard drives and SSDs', 'Storage Devices',
 'Storage', true),
('22222222-2222-2222-2222-222222222206', NULL, 'Power Supplies', 'power-supplies', 'PSU for computer systems',
 'Power Supplies', 'Power Supplies (PSU)', true),
('22222222-2222-2222-2222-222222222207', NULL, 'Cooling Systems', 'cooling-systems', 'Coolers and fans for PCs',
 'Cooling Systems', 'PC Cooling', true),


-- Subcategories for Processors
('22222222-2222-2222-2222-222222222211', '22222222-2222-2222-2222-222222222201', 'Intel Processors', 'intel-processors',
 'Intel CPUs', 'Intel Processors', 'Intel CPUs', true),
('22222222-2222-2222-2222-222222222212', '22222222-2222-2222-2222-222222222201', 'AMD Processors', 'amd-processors',
 'AMD CPUs', 'AMD Processors', 'AMD CPUs', true),

-- Subcategories for Graphics Cards
('22222222-2222-2222-2222-222222222221', '22222222-2222-2222-2222-222222222202', 'NVIDIA Graphics Cards',
 'nvidia-graphics', 'NVIDIA GPUs', 'NVIDIA Graphics Cards', 'NVIDIA GPUs', true),
('22222222-2222-2222-2222-222222222212', '22222222-2222-2222-2222-222222222201', 'AMD Graphics Cards', 'amd-graphics',
 'AMD GPUs', 'AMD Graphics Cards', 'AMD GPUs', true),

-- Subcategories for Storage
('22222222-2222-2222-2222-222222222251', '22222222-2222-2222-2222-222222222205', 'SSD Drives', 'ssd-drives',
 'Solid State Drives', 'SSD Storage', 'SSD Drives', true),
('22222222-2222-2222-2222-222222222252', '22222222-2222-2222-2222-222222222205', 'HDD Drives', 'hdd-drives',
 'Hard Disk Drives', 'HDD Storage', 'HDD Drives', true)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- ATTRIBUTE GROUPS
-- ============================================
INSERT INTO attribute_groups (id, name, slug, description)
VALUES ('33333333-3333-3333-3333-333333333301', 'Processor Specifications', 'processor_specifications',
        'Technical specifications for processors'),
       ('33333333-3333-3333-3333-333333333302', 'Graphics Card Specifications', 'graphics_card_specifications',
        'Technical specifications for graphics cards'),
       ('33333333-3333-3333-3333-333333333303', 'Memory Specifications', 'memory_specifications',
        'Technical specifications for RAM'),
       ('33333333-3333-3333-3333-333333333304', 'Storage Specifications', 'storage_specifications',
        'Technical specifications for storage devices'),
       ('33333333-3333-3333-3333-333333333305', 'General Specifications', 'general_specifications',
        'General technical specifications')

ON CONFLICT (id) DO NOTHING;

-- ============================================
-- ATTRIBUTES
-- ============================================
INSERT INTO attributes (id,
                        attribute_group_id,
                        name,
                        slug,
                        type,
                        unit,
                        is_filterable,
                        is_visible,
                        is_required,
                        sort_order)
VALUES
-- Processor
('44444444-4444-4444-4444-444444444401', '33333333-3333-3333-3333-333333333301','Socket', 'socket', 'select', NULL, true, true, true, 1),
('44444444-4444-4444-4444-444444444403', '33333333-3333-3333-3333-333333333301','Cores', 'cores', 'number', NULL, true, true, true, 2),
('44444444-4444-4444-4444-444444444406', '33333333-3333-3333-3333-333333333301','Threads', 'threads', 'number', NULL, true, true, true, 3),
('44444444-4444-4444-4444-444444444409', '33333333-3333-3333-3333-333333333301','Base Clock', 'base_clock', 'number', 'GHz', false, true, false, 4),
('44444444-4444-4444-4444-444444444411', '33333333-3333-3333-3333-333333333301','Boost Clock', 'boost_clock', 'number', 'GHz', false, true, false, 5),
('44444444-4444-4444-4444-444444444413', '33333333-3333-3333-3333-333333333301','TDP', 'tdp', 'number', 'W', true, true, false, 6),
-- GPU
('44444444-4444-4444-4444-444444444421', '33333333-3333-3333-3333-333333333302','GPU Chipset', 'gpu_chipset', 'select', NULL, true, true, true, 1),
('44444444-4444-4444-4444-444444444424', '33333333-3333-3333-3333-333333333302','VRAM', 'vram', 'number', 'GB', true, true, true, 2),
('44444444-4444-4444-4444-444444444427', '33333333-3333-3333-3333-333333333302','Memory Bus', 'memory_bus', 'number', 'bit', true, true, false, 3),
('44444444-4444-4444-4444-444444444428', '33333333-3333-3333-3333-333333333302','Core Clock', 'core_clock', 'number', 'MHz', false, true, false, 4),
-- Memory
('44444444-4444-4444-4444-444444444441', '33333333-3333-3333-3333-333333333303','Memory Type', 'memory_type', 'select', NULL, true, true, true, 1),
('44444444-4444-4444-4444-444444444443', '33333333-3333-3333-3333-333333333303','Capacity', 'memory_capacity', 'number', 'GB', true, true, true, 2),
('44444444-4444-4444-4444-444444444445', '33333333-3333-3333-3333-333333333303','Speed', 'memory_speed', 'number', 'MHz', true, true, false, 3),
-- Storage
('44444444-4444-4444-4444-444444444461', '33333333-3333-3333-3333-333333333304','Capacity', 'storage_capacity', 'number', 'GB', true, true, true, 1),
('44444444-4444-4444-4444-444444444464', '33333333-3333-3333-3333-333333333304','Interface', 'storage_interface', 'select', NULL, true, true, true, 2),
('44444444-4444-4444-4444-444444444466', '33333333-3333-3333-3333-333333333304','Read Speed', 'read_speed', 'number', 'MB/s', false, true, false, 3),
('44444444-4444-4444-4444-444444444467', '33333333-3333-3333-3333-333333333304','Write Speed', 'write_speed', 'number', 'MB/s', false, true, false, 4),
-- General
('44444444-4444-4444-4444-444444444481', '33333333-3333-3333-3333-333333333305',
 'Warranty', 'warranty', 'number', 'years', true, true, false, 1),
('44444444-4444-4444-4444-444444444482', '33333333-3333-3333-3333-333333333305','Color', 'color', 'select', NULL, true, true, false, 2),
('44444444-4444-4444-4444-444444444483', '33333333-3333-3333-3333-333333333305','RGB Lighting', 'rgb_lighting', 'boolean', NULL, true, true, false, 3)
ON CONFLICT (id) DO NOTHING;
-- ============================================
-- ATTRIBUTE VALUES
-- ============================================
INSERT INTO attribute_values (attribute_id,
                              value,
                              value_normalized,
                              value_numeric,
                              display_order)
VALUES
-- Socket
('44444444-4444-4444-4444-444444444401', 'LGA1700', 'lga1700', NULL, 1),
-- Cores
('44444444-4444-4444-4444-444444444403', '8', '8', 8, 1),
-- Threads
('44444444-4444-4444-4444-444444444406', '16', '16', 16, 1),
-- Base / Boost clock
('44444444-4444-4444-4444-444444444409', '3.6 GHz', '3.6ghz', 3.6, 1),
('44444444-4444-4444-4444-444444444411', '5.1 GHz', '5.1ghz', 5.1, 1),
-- TDP
('44444444-4444-4444-4444-444444444413', '125W', '125w', 125, 1),
-- GPU chipset
('44444444-4444-4444-4444-444444444421', 'RTX 4070', 'rtx4070', NULL, 1),
('44444444-4444-4444-4444-444444444421', 'RTX 4080', 'rtx4080', NULL, 2),
('44444444-4444-4444-4444-444444444421', 'RX 7800 XT', 'rx7800xt', NULL, 3),
-- VRAM
('44444444-4444-4444-4444-444444444424', '8GB GDDR6', '8gb', 8, 1),
('44444444-4444-4444-4444-444444444424', '12GB GDDR6', '12gb', 12, 2),
('44444444-4444-4444-4444-444444444424', '16GB GDDR6X', '16gb', 16, 3),
-- Memory
('44444444-4444-4444-4444-444444444441', 'DDR4', 'ddr4', NULL, 1),
('44444444-4444-4444-4444-444444444441', 'DDR5', 'ddr5', NULL, 2),
('44444444-4444-4444-4444-444444444443', '16GB', '16gb', 16, 1),
('44444444-4444-4444-4444-444444444443', '32GB', '32gb', 32, 2),
('44444444-4444-4444-4444-444444444445', '3200 MHz', '3200', 3200, 1),
('44444444-4444-4444-4444-444444444445', '6000 MHz', '6000', 6000, 2),
-- Storage
('44444444-4444-4444-4444-444444444461', '500GB', '500gb', 500, 1),
('44444444-4444-4444-4444-444444444461', '1TB', '1024gb', 1024, 2),
('44444444-4444-4444-4444-444444444461', '2TB', '2048gb', 2048, 3),
-- General
('44444444-4444-4444-4444-444444444483', 'Yes', 'true', 1, 1),
('44444444-4444-4444-4444-444444444483', 'No', 'false', 0, 2)
ON CONFLICT (id) DO NOTHING;


-- ============================================
-- PRODUCTS
-- ============================================
INSERT INTO products (id, name, model, slug, description, meta_title, meta_h1, sku, price, quantity, stock_status,
                      manufacturer_id, is_enable, weight, sort_order)
VALUES
-- Intel Processors
('55555555-5555-5555-5555-555555555501', 'Intel Core i7-13700K', 'BX8071513700K', 'intel-core-i7-13700k',
 'High-performance 16-core processor for gaming and content creation. Features Intel Thread Director and PCIe 5.0 support.',
 'Intel Core i7-13700K Processor', 'Intel Core i7-13700K', 'CPU-INT-13700K', 42999.00, 15, 'In Stock',
 '11111111-1111-1111-1111-111111111101', true, 0.5, 1),
('55555555-5555-5555-5555-555555555502', 'Intel Core i9-13900K', 'BX8071513900K', 'intel-core-i9-13900k',
 'Flagship 24-core processor with exceptional performance for demanding workloads and extreme gaming.',
 'Intel Core i9-13900K Processor', 'Intel Core i9-13900K', 'CPU-INT-13900K', 59999.00, 8, 'In Stock',
 '11111111-1111-1111-1111-111111111101', true, 0.5, 2),

-- AMD Processors
('55555555-5555-5555-5555-555555555503', 'AMD Ryzen 7 7800X3D', '100-100000910WOF', 'amd-ryzen-7-7800x3d',
 'Gaming-focused 8-core processor with 3D V-Cache technology for superior gaming performance.',
 'AMD Ryzen 7 7800X3D Processor', 'AMD Ryzen 7 7800X3D', 'CPU-AMD-7800X3D', 44999.00, 12, 'In Stock',
 '11111111-1111-1111-1111-111111111102', true, 0.45, 3),
('55555555-5555-5555-5555-555555555504', 'AMD Ryzen 9 7950X', '100-100000514WOF', 'amd-ryzen-9-7950x',
 'Ultimate 16-core processor for professionals and enthusiasts with leading-edge performance.',
 'AMD Ryzen 9 7950X Processor', 'AMD Ryzen 9 7950X', 'CPU-AMD-7950X', 64999.00, 6, 'In Stock',
 '11111111-1111-1111-1111-111111111102', true, 0.45, 4),

-- NVIDIA Graphics Cards
('55555555-5555-5555-5555-555555555505', 'ASUS ROG Strix RTX 4070', 'ROG-STRIX-RTX4070-O12G', 'asus-rtx-4070-strix',
 'Premium RTX 4070 graphics card with advanced cooling and factory overclock. Perfect for 1440p gaming.',
 'ASUS ROG Strix RTX 4070 Graphics Card', 'ASUS RTX 4070 Strix', 'GPU-ASUS-4070-STR', 67999.00, 10, 'In Stock',
 '11111111-1111-1111-1111-111111111104', true, 1.8, 5),
('55555555-5555-5555-5555-555555555506', 'MSI GeForce RTX 4080 SUPRIM X', 'RTX 4080 SUPRIM X 16G',
 'msi-rtx-4080-suprim', 'Top-tier RTX 4080 with exceptional cooling and premium build quality for 4K gaming.',
 'MSI RTX 4080 SUPRIM X Graphics Card', 'MSI RTX 4080 SUPRIM', 'GPU-MSI-4080-SUP', 129999.00, 5, 'In Stock',
 '11111111-1111-1111-1111-111111111105', true, 2.2, 6),
('55555555-5555-5555-5555-555555555507', 'Gigabyte RTX 4070 Gaming OC', 'GV-N4070GAMING OC-12GD',
 'gigabyte-rtx-4070-gaming', 'Reliable RTX 4070 with Windforce cooling system and RGB lighting.',
 'Gigabyte RTX 4070 Gaming OC', 'Gigabyte RTX 4070', 'GPU-GB-4070-GAM', 62999.00, 14, 'In Stock',
 '11111111-1111-1111-1111-111111111106', true, 1.6, 7),

-- AMD Graphics Cards
('55555555-5555-5555-5555-555555555508', 'ASUS TUF RX 7800 XT', 'TUF-RX7800XT-O16G-GAMING', 'asus-rx-7800-xt',
 'Powerful AMD graphics card with 16GB VRAM, excellent for high-resolution gaming.',
 'ASUS TUF RX 7800 XT Graphics Card', 'ASUS RX 7800 XT', 'GPU-ASUS-7800XT', 58999.00, 9, 'In Stock',
 '11111111-1111-1111-1111-111111111104', true, 1.7, 8),

-- Motherboards
('55555555-5555-5555-5555-555555555509', 'ASUS ROG Strix Z790-E Gaming', 'ROG STRIX Z790-E GAMING WIFI',
 'asus-z790-e-strix', 'Premium Intel Z790 motherboard with DDR5, PCIe 5.0, WiFi 6E, and comprehensive cooling.',
 'ASUS ROG Strix Z790-E Motherboard', 'ASUS Z790-E Gaming', 'MB-ASUS-Z790-E', 44999.00, 7, 'In Stock',
 '11111111-1111-1111-1111-111111111104', true, 1.5, 9),
('55555555-5555-5555-5555-555555555510', 'MSI MAG X670E Tomahawk', 'MAG X670E TOMAHAWK WIFI', 'msi-x670e-tomahawk',
 'High-performance AMD X670E motherboard with excellent VRM and connectivity options.',
 'MSI MAG X670E Tomahawk Motherboard', 'MSI X670E Tomahawk', 'MB-MSI-X670E-TH', 39999.00, 8, 'In Stock',
 '11111111-1111-1111-1111-111111111105', true, 1.4, 10),

-- RAM
('55555555-5555-5555-5555-555555555511', 'Corsair Vengeance DDR5 32GB', 'CMK32GX5M2B6000C36',
 'corsair-vengeance-ddr5-32gb', 'High-speed DDR5 memory kit (2x16GB) running at 6000MHz with optimized latencies.',
 'Corsair Vengeance DDR5 32GB', 'Corsair DDR5 32GB', 'RAM-COR-DDR5-32', 14999.00, 20, 'In Stock',
 '11111111-1111-1111-1111-111111111107', true, 0.3, 11),
('55555555-5555-5555-5555-555555555512', 'Kingston Fury Beast DDR4 32GB', 'KF432C16BBK2/32', 'kingston-fury-ddr4-32gb',
 'Reliable DDR4 memory kit (2x16GB) at 3200MHz for mainstream systems.', 'Kingston Fury Beast DDR4 32GB',
 'Kingston DDR4 32GB', 'RAM-KNG-DDR4-32', 8999.00, 25, 'In Stock', '11111111-1111-1111-1111-111111111108', true, 0.25,
 12),

-- SSDs
('55555555-5555-5555-5555-555555555513', 'Samsung 980 PRO 1TB', 'MZ-V8P1T0BW', 'samsung-980-pro-1tb',
 'Flagship PCIe 4.0 NVMe SSD with exceptional speeds up to 7000 MB/s read.', 'Samsung 980 PRO 1TB SSD',
 'Samsung 980 PRO 1TB', 'SSD-SAM-980P-1TB', 12999.00, 18, 'In Stock', '11111111-1111-1111-1111-111111111109', true,
 0.08, 13),
('55555555-5555-5555-5555-555555555514', 'Samsung 980 PRO 2TB', 'MZ-V8P2T0BW', 'samsung-980-pro-2tb',
 'High-capacity PCIe 4.0 NVMe SSD for professionals and power users.', 'Samsung 980 PRO 2TB SSD', 'Samsung 980 PRO 2TB',
 'SSD-SAM-980P-2TB', 23999.00, 12, 'In Stock', '11111111-1111-1111-1111-111111111109', true, 0.08, 14),
('55555555-5555-5555-5555-555555555515', 'WD Black SN850X 1TB', 'WDS100T2X0E', 'wd-black-sn850x-1tb',
 'Gaming-focused PCIe 4.0 SSD with Game Mode 2.0 for optimized performance.', 'WD Black SN850X 1TB SSD',
 'WD SN850X 1TB', 'SSD-WD-850X-1TB', 11999.00, 15, 'In Stock', '11111111-1111-1111-1111-111111111110', true, 0.07, 15),

-- Power Supplies
('55555555-5555-5555-5555-555555555516', 'Corsair RM850x 850W', 'CP-9020200-EU', 'corsair-rm850x',
 '80 PLUS Gold certified modular PSU with quiet operation and Japanese capacitors.', 'Corsair RM850x 850W PSU',
 'Corsair RM850x', 'PSU-COR-RM850X', 13999.00, 11, 'In Stock', '11111111-1111-1111-1111-111111111107', true, 2.5, 16),

-- Cooling
('55555555-5555-5555-5555-555555555517', 'Corsair iCUE H150i Elite LCD', 'CW-9060062-WW', 'corsair-h150i-elite-lcd',
 '360mm AIO liquid cooler with customizable LCD screen and powerful cooling.', 'Corsair H150i Elite LCD Cooler',
 'Corsair H150i LCD', 'COOL-COR-H150LCD', 24999.00, 9, 'In Stock', '11111111-1111-1111-1111-111111111107', true, 1.8,
 17)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- ATTRIBUTE-PRODUCT RELATIONSHIPS
-- ============================================
INSERT INTO attribute_products (product_id, attribute_id)
VALUES
-- Intel Core i7-13700K
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444401'), -- Socket LGA1700
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444405'), -- 16 Cores
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444407'), -- 24 Threads
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444409'), -- 3.6 GHz Base
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444411'), -- 5.1 GHz Boost
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444413'), -- 125W TDP
('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- Intel Core i9-13900K
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444401'), -- Socket LGA1700
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444407'), -- 24 Threads
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444410'), -- 4.2 GHz Base
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444412'), -- 5.4 GHz Boost
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444414'), -- 170W TDP
('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- AMD Ryzen 7 7800X3D
('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444402'), -- Socket AM5
('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444403'), -- 8 Cores
('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444406'), -- 16 Threads
('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444413'), -- 125W TDP
('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- AMD Ryzen 9 7950X
('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444402'), -- Socket AM5
('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444405'), -- 16 Cores
('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444408'), -- 32 Threads
('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444414'), -- 170W TDP
('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- ASUS ROG Strix RTX 4070
('55555555-5555-5555-5555-555555555505', '44444444-4444-4444-4444-444444444421'), -- RTX 4070 Chipset
('55555555-5555-5555-5555-555555555505', '44444444-4444-4444-4444-444444444425'), -- 12GB GDDR6
('55555555-5555-5555-5555-555555555505', '44444444-4444-4444-4444-444444444427'), -- 256-bit Bus
('55555555-5555-5555-5555-555555555505', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555505', '44444444-4444-4444-4444-444444444483'), -- RGB Lighting

-- MSI GeForce RTX 4080 SUPRIM X
('55555555-5555-5555-5555-555555555506', '44444444-4444-4444-4444-444444444422'), -- RTX 4080 Chipset
('55555555-5555-5555-5555-555555555506', '44444444-4444-4444-4444-444444444426'), -- 16GB GDDR6X
('55555555-5555-5555-5555-555555555506', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555506', '44444444-4444-4444-4444-444444444483'), -- RGB Lighting

-- Gigabyte RTX 4070 Gaming OC
('55555555-5555-5555-5555-555555555507', '44444444-4444-4444-4444-444444444421'), -- RTX 4070 Chipset
('55555555-5555-5555-5555-555555555507', '44444444-4444-4444-4444-444444444425'), -- 12GB GDDR6
('55555555-5555-5555-5555-555555555507', '44444444-4444-4444-4444-444444444427'), -- 256-bit Bus
('55555555-5555-5555-5555-555555555507', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555507', '44444444-4444-4444-4444-444444444483'), -- RGB Lighting

-- ASUS TUF RX 7800 XT
('55555555-5555-5555-5555-555555555508', '44444444-4444-4444-4444-444444444423'), -- RX 7800 XT Chipset
('55555555-5555-5555-5555-555555555508', '44444444-4444-4444-4444-444444444426'), -- 16GB GDDR6
('55555555-5555-5555-5555-555555555508', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555508', '44444444-4444-4444-4444-444444444483'), -- RGB Lighting

-- Corsair Vengeance DDR5 32GB
('55555555-5555-5555-5555-555555555511', '44444444-4444-4444-4444-444444444442'), -- DDR5
('55555555-5555-5555-5555-555555555511', '44444444-4444-4444-4444-444444444444'), -- 32GB
('55555555-5555-5555-5555-555555555511', '44444444-4444-4444-4444-444444444446'), -- 6000 MHz
('55555555-5555-5555-5555-555555555511', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555511', '44444444-4444-4444-4444-444444444482'), -- Black

-- Kingston Fury Beast DDR4 32GB
('55555555-5555-5555-5555-555555555512', '44444444-4444-4444-4444-444444444441'), -- DDR4
('55555555-5555-5555-5555-555555555512', '44444444-4444-4444-4444-444444444444'), -- 32GB
('55555555-5555-5555-5555-555555555512', '44444444-4444-4444-4444-444444444445'), -- 3200 MHz
('55555555-5555-5555-5555-555555555512', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555512', '44444444-4444-4444-4444-444444444482'), -- Black

-- Samsung 980 PRO 1TB
('55555555-5555-5555-5555-555555555513', '44444444-4444-4444-4444-444444444462'), -- 1TB
('55555555-5555-5555-5555-555555555513', '44444444-4444-4444-4444-444444444464'), -- NVMe PCIe 4.0
('55555555-5555-5555-5555-555555555513', '44444444-4444-4444-4444-444444444466'), -- 7000 MB/s Read
('55555555-5555-5555-5555-555555555513', '44444444-4444-4444-4444-444444444467'), -- 5000 MB/s Write
('55555555-5555-5555-5555-555555555513', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- Samsung 980 PRO 2TB
('55555555-5555-5555-5555-555555555514', '44444444-4444-4444-4444-444444444463'), -- 2TB
('55555555-5555-5555-5555-555555555514', '44444444-4444-4444-4444-444444444464'), -- NVMe PCIe 4.0
('55555555-5555-5555-5555-555555555514', '44444444-4444-4444-4444-444444444466'), -- 7000 MB/s Read
('55555555-5555-5555-5555-555555555514', '44444444-4444-4444-4444-444444444467'), -- 5000 MB/s Write
('55555555-5555-5555-5555-555555555514', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- WD Black SN850X 1TB
('55555555-5555-5555-5555-555555555515', '44444444-4444-4444-4444-444444444462'), -- 1TB
('55555555-5555-5555-5555-555555555515', '44444444-4444-4444-4444-444444444464'), -- NVMe PCIe 4.0
('55555555-5555-5555-5555-555555555515', '44444444-4444-4444-4444-444444444466'), -- 7000 MB/s Read
('55555555-5555-5555-5555-555555555515', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty

-- Corsair RM850x 850W
('55555555-5555-5555-5555-555555555516', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555516', '44444444-4444-4444-4444-444444444482'), -- Black

-- Corsair iCUE H150i Elite LCD
('55555555-5555-5555-5555-555555555517', '44444444-4444-4444-4444-444444444481'), -- 3 years warranty
('55555555-5555-5555-5555-555555555517', '44444444-4444-4444-4444-444444444483')  -- RGB Lighting
ON CONFLICT (product_id, attribute_id) DO NOTHING;

-- Success message
SELECT 'Seed data for computer electronics has been inserted successfully!' AS message;
