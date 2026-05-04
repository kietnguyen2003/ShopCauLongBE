-- Shop Cau Long Backend - PostgreSQL database bootstrap script
-- Usage:
--   psql -U postgres -f scripts/create_database.sql
--
-- If you are using docker-compose from this repo, the database/user are already
-- created by Postgres environment variables. This script is useful when you want
-- to create the schema manually.

CREATE DATABASE order_demo;

\connect order_demo;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS categories (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    image TEXT,
    status TEXT DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DOUBLE PRECISION NOT NULL,
    stock BIGINT DEFAULT 0,
    image TEXT,
    category TEXT,
    status TEXT DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    customer_name TEXT,
    phone TEXT,
    address TEXT,
    email TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses (user_id);

CREATE TABLE IF NOT EXISTS carts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    status TEXT DEFAULT 'active',
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS cart_items (
    id BIGSERIAL PRIMARY KEY,
    cart_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity BIGINT NOT NULL,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id ON cart_items (cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_user_id ON cart_items (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cart_product ON cart_items (cart_id, product_id);

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    subtotal_amount DOUBLE PRECISION DEFAULT 0,
    discount_amount DOUBLE PRECISION DEFAULT 0,
    total_amount DOUBLE PRECISION NOT NULL,
    coupon_id BIGINT,
    coupon_code TEXT,
    coupon_codes TEXT,
    status TEXT DEFAULT 'pending',
    customer_name TEXT,
    phone TEXT,
    address TEXT,
    email TEXT,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL,
    quantity BIGINT NOT NULL,
    image TEXT,
    category TEXT,
    description TEXT,
    created_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items (order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items (product_id);

CREATE TABLE IF NOT EXISTS reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    rating BIGINT,
    comment TEXT,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_reviews_product_id ON reviews (product_id);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews (user_id);

CREATE TABLE IF NOT EXISTS coupons (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT,
    description TEXT,
    discount_type TEXT NOT NULL,
    discount_value DOUBLE PRECISION NOT NULL,
    min_order_amount DOUBLE PRECISION DEFAULT 0,
    max_discount_amount DOUBLE PRECISION,
    usage_limit BIGINT,
    used_count BIGINT DEFAULT 0,
    usage_limit_per_user BIGINT DEFAULT 1,
    start_at BIGINT,
    end_at BIGINT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at BIGINT,
    updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    discount_amount DOUBLE PRECISION NOT NULL,
    created_at BIGINT
);

CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_id ON coupon_redemptions (coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_user_id ON coupon_redemptions (user_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_order_id ON coupon_redemptions (order_id);

-- Compatibility cleanup for older schema versions that allowed only one coupon
-- redemption per order.
ALTER TABLE coupon_redemptions DROP CONSTRAINT IF EXISTS coupon_redemptions_order_id_key;
DROP INDEX IF EXISTS uni_coupon_redemptions_order_id;

-- Seed data equivalent to internal/infrastructure/database/seed.go.
INSERT INTO users (username, email, password, is_admin, created_at, updated_at)
SELECT
    'admin',
    'admin@demo.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    TRUE,
    EXTRACT(EPOCH FROM NOW())::BIGINT,
    EXTRACT(EPOCH FROM NOW())::BIGINT
WHERE NOT EXISTS (SELECT 1 FROM users WHERE is_admin = TRUE);

INSERT INTO categories (name, description, image, status, created_at, updated_at)
VALUES
    ('clothing', 'Quan ao cau long', 'https://cdn.shopvnb.com/uploads/gallery/ao-cau-long-mizuno-vm1076-nam-xanh_1728499552.webp', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('racket', 'Vot cau long', 'https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-yonex-arcsaber-11-pro-chinh-hang-1.webp', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT)
ON CONFLICT (name) DO NOTHING;

INSERT INTO products (name, description, price, stock, image, category, status, created_at, updated_at)
VALUES
    ('Áo Cầu Lông Yonex Pro', 'Áo cầu lông chính hãng Yonex, chất liệu thoáng mát, thấm hút mồ hôi tốt', 450000, 20, 'https://www.yonex.com/media/catalog/product/a/l/all_10700_629.png?quality=80&bg-color=248,248,248,0.75&fit=bounds&height=300&width=240&canvas=240:300', 'clothing', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('Quần Cầu Lông Victor', 'Quần cầu lông Victor, thiết kế năng động, thoải mái khi vận động', 380000, 15, 'https://cdn.shopvnb.com/uploads/gallery/quan-cau-long-victor-q41-nam-xanh_1715654738.webp', 'clothing', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('Áo Cầu Lông Mizuno', 'Giày cầu lông Mizuno chuyên nghiệp, đế chống trượt, hỗ trợ di chuyển tốt', 1200000, 10, 'https://cdn.shopvnb.com/img/300x300/uploads/gallery/ao-cau-long-mizuno-vm1076-nam-xanh_1728499552.webp', 'clothing', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('Vợt Cầu Lông Yonex Arcsaber 11', 'Vợt cầu lông Yonex Arcsaber 11, công nghệ tiên tiến, phù hợp cho người chơi chuyên nghiệp', 3500000, 8, 'https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-yonex-arcsaber-11-pro-chinh-hang-1.webp', 'racket', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT),
    ('Vợt Cầu Lông Victor Jetspeed S12', 'Vợt Victor Jetspeed S12, thiết kế aerodynamic, tốc độ swing nhanh', 2800000, 12, 'https://cdn.shopvnb.com/img/600x600/uploads/san_pham/combo-mua-vot-cau-long-victor-jetspeed-12-js-12-tang-vot-js120-2-vot-victor-tk9-1.webp', 'racket', 'active', EXTRACT(EPOCH FROM NOW())::BIGINT, EXTRACT(EPOCH FROM NOW())::BIGINT);
