package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func ApplyDatabaseHardening(db *gorm.DB) error {
	statements := []string{
		`ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id BIGINT`,

		`UPDATE products p SET category_id = c.id FROM categories c WHERE p.category_id IS NULL AND p.category = c.name`,
		`ALTER TABLE products ALTER COLUMN price TYPE BIGINT USING price::BIGINT`,

		`ALTER TABLE orders ALTER COLUMN subtotal_amount TYPE BIGINT USING subtotal_amount::BIGINT`,
		`ALTER TABLE orders ALTER COLUMN discount_amount TYPE BIGINT USING discount_amount::BIGINT`,
		`ALTER TABLE orders ALTER COLUMN total_amount TYPE BIGINT USING total_amount::BIGINT`,
		`ALTER TABLE order_items ALTER COLUMN price TYPE BIGINT USING price::BIGINT`,

		`ALTER TABLE coupons ALTER COLUMN discount_value TYPE BIGINT USING discount_value::BIGINT`,
		`ALTER TABLE coupons ALTER COLUMN min_order_amount TYPE BIGINT USING min_order_amount::BIGINT`,
		`ALTER TABLE coupons ALTER COLUMN max_discount_amount TYPE BIGINT USING max_discount_amount::BIGINT`,
		`ALTER TABLE coupon_redemptions ALTER COLUMN discount_amount TYPE BIGINT USING discount_amount::BIGINT`,

		`CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_products_status ON products(status)`,
		`CREATE INDEX IF NOT EXISTS idx_cart_items_product_id ON cart_items(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_one_default_address_per_user ON addresses(user_id) WHERE is_default = TRUE`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	constraints := []string{
		`ALTER TABLE addresses ADD CONSTRAINT fk_addresses_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`,
		`ALTER TABLE carts ADD CONSTRAINT fk_carts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`,
		`ALTER TABLE cart_items ADD CONSTRAINT fk_cart_items_cart FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE`,
		`ALTER TABLE cart_items ADD CONSTRAINT fk_cart_items_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT`,
		`ALTER TABLE products ADD CONSTRAINT fk_products_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE RESTRICT`,
		`ALTER TABLE orders ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT`,
		`ALTER TABLE order_items ADD CONSTRAINT fk_order_items_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE`,
		`ALTER TABLE order_items ADD CONSTRAINT fk_order_items_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT`,
		`ALTER TABLE reviews ADD CONSTRAINT fk_reviews_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`,
		`ALTER TABLE reviews ADD CONSTRAINT fk_reviews_product FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE`,
		`ALTER TABLE coupon_redemptions ADD CONSTRAINT fk_coupon_redemptions_coupon FOREIGN KEY (coupon_id) REFERENCES coupons(id) ON DELETE RESTRICT`,
		`ALTER TABLE coupon_redemptions ADD CONSTRAINT fk_coupon_redemptions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT`,
		`ALTER TABLE coupon_redemptions ADD CONSTRAINT fk_coupon_redemptions_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE`,
		`ALTER TABLE notifications ADD CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE`,
		`ALTER TABLE products ADD CONSTRAINT chk_products_stock_non_negative CHECK (stock >= 0)`,
		`ALTER TABLE products ADD CONSTRAINT chk_products_price_non_negative CHECK (price >= 0)`,
		`ALTER TABLE cart_items ADD CONSTRAINT chk_cart_items_quantity_positive CHECK (quantity > 0)`,
		`ALTER TABLE order_items ADD CONSTRAINT chk_order_items_quantity_positive CHECK (quantity > 0)`,
		`ALTER TABLE reviews ADD CONSTRAINT chk_reviews_rating CHECK (rating IS NULL OR (rating >= 1 AND rating <= 5))`,
		`ALTER TABLE orders ADD CONSTRAINT chk_orders_status CHECK (status IN ('pending', 'confirmed', 'cancelled'))`,
		`ALTER TABLE products ADD CONSTRAINT chk_products_status CHECK (status IN ('active', 'inactive'))`,
		`ALTER TABLE coupons ADD CONSTRAINT chk_coupons_discount_type CHECK (discount_type IN ('percentage', 'fixed_amount'))`,
	}

	for _, constraint := range constraints {
		if err := addConstraintIfMissing(db, constraint); err != nil {
			return err
		}
	}

	return nil
}

func addConstraintIfMissing(db *gorm.DB, alterStatement string) error {
	escapedStatement := strings.ReplaceAll(alterStatement, `'`, `''`)
	return db.Exec(fmt.Sprintf(`
DO $$
BEGIN
	BEGIN
		EXECUTE '%s';
	EXCEPTION
		WHEN duplicate_object THEN NULL;
		WHEN invalid_table_definition THEN NULL;
	END;
END $$;
`, escapedStatement)).Error
}
