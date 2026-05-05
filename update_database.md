# Database Update Notes

This document summarizes the database and backend updates implemented from `update_database.json`.

## 1. Database Hardening Migration

Added a new hardening migration file:

```text
internal/infrastructure/database/hardening.go
```

It is called from:

```text
internal/infrastructure/database/migrations.go
```

Current startup flow:

```text
database.AutoMigrate(...)
-> database.ApplyDatabaseHardening(...)
```

The goal is to keep GORM `AutoMigrate` for table creation, then apply stronger PostgreSQL-specific constraints and indexes.

## 2. Foreign Keys

The project previously used columns such as `user_id`, `product_id`, `order_id`, and `coupon_id`, but most of them were not enforced by PostgreSQL.

The hardening migration now adds real foreign keys.

### User Relations

```text
addresses.user_id -> users.id
carts.user_id -> users.id
orders.user_id -> users.id
reviews.user_id -> users.id
coupon_redemptions.user_id -> users.id
notifications.user_id -> users.id
```

### Product Relations

```text
cart_items.product_id -> products.id
order_items.product_id -> products.id
reviews.product_id -> products.id
```

### Order Relations

```text
order_items.order_id -> orders.id
coupon_redemptions.order_id -> orders.id
```

### Coupon Relations

```text
coupon_redemptions.coupon_id -> coupons.id
```

### Category Relations

```text
products.category_id -> categories.id
```

## 3. Delete Behavior

The migration uses different `ON DELETE` rules depending on the business meaning.

### Cascade Delete

Used for dependent data that should disappear with the owner:

```text
addresses -> users
carts -> users
cart_items -> carts
order_items -> orders
reviews -> users/products
notifications -> users
coupon_redemptions -> orders
```

### Restrict Delete

Used for important ecommerce history:

```text
orders -> users
cart_items/order_items -> products
coupon_redemptions -> users/coupons
products -> categories
```

This prevents deleting product/category/user records that are already referenced by order or coupon history.

## 4. Product Category Normalization

Before:

```text
products.category TEXT
```

Problem:

```text
products.category
```

was just text, so it could become inconsistent with the `categories` table.

Now products support:

```text
products.category_id BIGINT
products.category TEXT
```

`category_id` is the normalized relational field.

`category` is still kept for backward compatibility and display/snapshot-style usage.

## 5. Product Category Backfill

The hardening migration adds and fills `category_id` like this:

```sql
ALTER TABLE products ADD COLUMN IF NOT EXISTS category_id BIGINT;

UPDATE products p
SET category_id = c.id
FROM categories c
WHERE p.category_id IS NULL
  AND p.category = c.name;
```

This maps old product rows from text category names to real category IDs.

## 6. Product API Compatibility

Product APIs now support `category_id`.

Example create product:

```json
{
  "name": "Vot Cau Long Yonex",
  "description": "Mo ta san pham",
  "price": 1500000,
  "stock": 10,
  "image": "https://example.com/product.png",
  "category_id": 2,
  "category": "racket"
}
```

Backend behavior:

```text
If category_id is provided:
  use categories.id to resolve the category

If category_id is missing:
  fallback to category name for old clients
```

Response now includes both:

```json
{
  "category_id": 2,
  "category": "racket"
}
```

## 7. Product Query Update

Product listing now supports category filtering by ID:

```text
GET /api/products?category_id=2
```

Old category-name filtering still works:

```text
GET /api/products?category=racket
```

`category_id` is preferred because it is enforced by the database.

## 8. Money Type Update

Before, money values used `float64` in Go and `DOUBLE PRECISION` in PostgreSQL.

Problem:

```text
Floating point types can create rounding errors.
```

For VND ecommerce, money should be stored as integer dong.

Now money uses:

```text
Go: int64
PostgreSQL: BIGINT
```

## 9. Money Fields Changed

### Products

```text
products.price
```

### Orders

```text
orders.subtotal_amount
orders.discount_amount
orders.total_amount
```

### Order Items

```text
order_items.price
```

### Coupons

```text
coupons.discount_value
coupons.min_order_amount
coupons.max_discount_amount
```

### Coupon Redemptions

```text
coupon_redemptions.discount_amount
```

## 10. Money Migration SQL

The hardening migration converts existing columns:

```sql
ALTER TABLE products ALTER COLUMN price TYPE BIGINT USING price::BIGINT;
ALTER TABLE orders ALTER COLUMN subtotal_amount TYPE BIGINT USING subtotal_amount::BIGINT;
ALTER TABLE orders ALTER COLUMN discount_amount TYPE BIGINT USING discount_amount::BIGINT;
ALTER TABLE orders ALTER COLUMN total_amount TYPE BIGINT USING total_amount::BIGINT;
ALTER TABLE order_items ALTER COLUMN price TYPE BIGINT USING price::BIGINT;
ALTER TABLE coupons ALTER COLUMN discount_value TYPE BIGINT USING discount_value::BIGINT;
ALTER TABLE coupons ALTER COLUMN min_order_amount TYPE BIGINT USING min_order_amount::BIGINT;
ALTER TABLE coupons ALTER COLUMN max_discount_amount TYPE BIGINT USING max_discount_amount::BIGINT;
ALTER TABLE coupon_redemptions ALTER COLUMN discount_amount TYPE BIGINT USING discount_amount::BIGINT;
```

## 11. Coupon Calculation Update

Coupon money values now use `int64`.

Fixed amount coupon:

```text
discount_value = 10000
```

means:

```text
10,000 VND
```

Percentage coupon:

```text
discount_type = percentage
discount_value = 10
```

means:

```text
10%
```

Calculation:

```go
discount = subtotalAmount * discountValue / 100
```

Because the result is an integer, fractional dong is automatically truncated.

## 12. Check Constraints

The hardening migration adds constraints to prevent invalid values at the database level.

### Product Constraints

```text
stock >= 0
price >= 0
status IN ('active', 'inactive')
```

### Cart and Order Item Constraints

```text
cart_items.quantity > 0
order_items.quantity > 0
```

### Review Constraints

```text
rating IS NULL OR rating BETWEEN 1 AND 5
```

### Order Constraints

```text
status IN ('pending', 'confirmed', 'cancelled')
```

### Coupon Constraints

```text
discount_type IN ('percentage', 'fixed_amount')
```

## 13. Index Updates

Added indexes for common ecommerce queries.

```text
idx_products_category_id
idx_orders_user_id
idx_orders_status
idx_products_status
idx_cart_items_product_id
idx_coupons_code
```

These help with:

```text
product filtering
user order history
admin order filtering
active product listing
cart/product lookup
coupon lookup by code
```

## 14. One Default Address Per User

Added partial unique index:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS idx_one_default_address_per_user
ON addresses(user_id)
WHERE is_default = TRUE;
```

This enforces:

```text
One user can have many addresses,
but only one address can be default.
```

## 15. Checkout Stock Update

Before, checkout loaded products, calculated new stock in memory, then saved product rows.

That can be unsafe under concurrent checkout requests.

Now checkout stock decrement happens inside the order transaction using an atomic conditional update:

```sql
UPDATE products
SET stock = stock - ?
WHERE id = ?
  AND stock >= ?;
```

If `RowsAffected == 0`, the repository returns:

```go
product.ErrInsufficientStock
```

This reduces the risk of overselling when two users buy the same product at the same time.

## 16. Checkout Transaction Flow

Current transaction flow:

```text
Begin transaction
-> Atomically decrease product stock
-> Create order
-> Create order_items
-> If coupons exist:
     lock coupon row
     check per-user usage
     create coupon_redemption
     increment used_count
-> Clear cart items
-> Commit
```

If any step fails:

```text
Rollback
```

## 17. Files Changed

Main database files:

```text
internal/infrastructure/database/hardening.go
internal/infrastructure/database/migrations.go
internal/infrastructure/database/gorm_product_repository.go
internal/infrastructure/database/gorm_order_repository.go
internal/infrastructure/database/gorm_coupon_repository.go
```

Domain files:

```text
internal/domain/product/product.go
internal/domain/order/order.go
internal/domain/coupon/coupon.go
```

Application files:

```text
internal/application/product/*
internal/application/cart/*
internal/application/order/*
internal/application/coupon/*
```

HTTP model/handler files:

```text
internal/interfaces/http/product_models.go
internal/interfaces/http/product_handler.go
internal/interfaces/http/cart_models.go
internal/interfaces/http/order_models.go
internal/interfaces/http/coupon_models.go
```

Docs:

```text
api.json
```

## 18. Verification

The project was formatted with:

```bash
gofmt
```

Tests were run with:

```bash
go test ./...
```

Result:

```text
PASS
```

## 19. Important Migration Notes

Foreign keys and check constraints can fail to apply if existing database data is invalid.

Before applying to a real database, check for orphan rows, for example:

```sql
SELECT *
FROM orders o
LEFT JOIN users u ON u.id = o.user_id
WHERE u.id IS NULL;
```

Also check money values before converting to `BIGINT`:

```sql
SELECT price
FROM products
WHERE price != price::BIGINT;
```

If these queries return rows, clean or migrate the data first.

## 20. Remaining Ideas From The Plan

Not yet implemented:

```text
created_at / updated_at BIGINT -> TIMESTAMPTZ
users.password -> users.password_hash
simplify coupon schema
remove cart_items.user_id
soft delete products with deleted_at
standardize API error codes
```

These are good follow-up tasks, but each one changes either API behavior or data contracts and should be handled in separate focused updates.
