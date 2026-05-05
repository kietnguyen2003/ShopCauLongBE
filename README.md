# Shop Cau Long Backend

Backend API for a badminton e-commerce system. The service supports product browsing, authentication, carts, checkout, orders, reviews, addresses, categories, admin management, and multi-coupon discounts.

## 1. Project Overview

Shop Cau Long Backend is a RESTful API built with Go, Gin, PostgreSQL, and GORM. It follows a Clean Architecture / DDD-inspired structure:

- Domain layer owns business entities and rules.
- Application layer owns use cases and ports.
- Infrastructure layer implements database, security, config, migrations, and seed data.
- Interface layer exposes HTTP handlers and request/response models.

The current checkout flow is cart-based: users add products to cart, choose an address, optionally apply one or more coupons, and create an order. Backend recalculates all prices and discounts server-side.

## 2. Features

- User registration, login, JWT authentication, token refresh, password reset, and password change.
- Admin login and admin-only endpoints.
- Product browsing, search, filtering, sorting, category browsing.
- Product management for admin.
- Category management for admin.
- Cart management with stock validation.
- Address management with default address support.
- Order creation from current cart.
- Order status update with stock restock when status becomes `cancelled`.
- Product reviews:
  - User can create multiple reviews.
  - Review create can be partial: rating only, comment only, or both.
  - User must have bought the product before reviewing.
  - Review update changes rating only.
- Coupon system:
  - Admin creates and manages coupons.
  - Authenticated users can view active coupons they have not used.
  - Cart coupon validation supports one or multiple coupon codes.
  - Orders can apply multiple coupons.
  - `usage_limit`, `usage_limit_per_user`, date windows, min order, and max discount rules are enforced.
  - Coupon redemptions and coupon usage count are saved during order creation.

## 3. Tech Stack

- Language: Go 1.21
- HTTP framework: Gin
- Database: PostgreSQL
- ORM: GORM
- Authentication: JWT
- Password hashing: bcrypt
- Local database tooling: Docker Compose, PostgreSQL 15, pgAdmin

## 4. System Architecture

![System Architecture](snapshot/image.png)

Dependency direction:

```text
HTTP handlers -> Application services -> Domain entities
Application ports <- Infrastructure repositories
```

## 5. Database Design

Main tables:

- `users`: user credentials and admin flag.
- `products`: product catalog, stock, category, status.
- `categories`: product categories with soft-delete style status.
- `carts`: active user carts.
- `cart_items`: products and quantities in cart.
- `addresses`: user shipping/contact addresses.
- `orders`: order header, status, subtotal, discount, total, coupon summary.
- `order_items`: purchased product snapshots.
- `reviews`: product reviews from users.
- `coupons`: coupon definitions managed by admin.
- `coupon_redemptions`: coupon usage history per user and order.

Important order fields:

- `subtotal_amount`: total before discounts.
- `discount_amount`: total discount amount.
- `total_amount`: final payable amount.
- `coupon_id`: first applied coupon id for backward compatibility.
- `coupon_code`: comma-separated legacy coupon summary.
- `coupon_codes`: comma-separated applied coupon codes.

Important coupon fields:

- `code`: unique uppercase coupon code.
- `discount_type`: `percentage` or `fixed_amount`.
- `discount_value`: discount value.
- `min_order_amount`: minimum cart/order amount.
- `max_discount_amount`: optional cap for percentage discounts.
- `usage_limit`: optional global usage limit.
- `used_count`: total successful redemptions.
- `usage_limit_per_user`: max redemptions per user.
- `start_at`, `end_at`: optional validity window.
- `is_active`: whether customers can see/use the coupon.

## 6. API Documentation

All responses use:

```json
{
  "code": 200,
  "msg": "Message",
  "data": {}
}
```

Detailed examples are maintained in `api.json`.

### Auth

| Method | Path | Auth | Description |
|---|---|---:|---|
| POST | `/auth/register` | No | Register user |
| POST | `/auth/login` | No | User login |
| POST | `/auth/admin-login` | No | Admin login |
| GET | `/auth/me` | Yes | Current user |
| POST | `/auth/logout` | Yes | Logout |
| POST | `/auth/refresh-token` | Yes | Refresh token |
| PUT | `/auth/change-password` | Yes | Change password |
| POST | `/auth/forgot-password` | No | Generate reset token |
| POST | `/auth/reset-password` | No | Reset password |

### Public Catalog

| Method | Path | Auth | Description |
|---|---|---:|---|
| GET | `/api/products` | No | Product list with query filters |
| GET | `/api/products/search` | No | Search products |
| GET | `/api/products/category/:category` | No | Products by category |
| GET | `/api/products/:id` | No | Product detail |
| GET | `/api/products/:id/reviews` | No | Product reviews |
| GET | `/api/categories` | No | Category list |
| GET | `/api/categories/:id` | No | Category detail |

### User APIs

| Method | Path | Auth | Description |
|---|---|---:|---|
| GET | `/api/cart` | Yes | Get cart |
| POST | `/api/cart/items` | Yes | Add items to cart |
| PUT | `/api/cart/items/:id` | Yes | Update cart item quantity |
| DELETE | `/api/cart/items/:id` | Yes | Delete cart item |
| DELETE | `/api/cart/clear` | Yes | Clear cart |
| POST | `/api/orders` | Yes | Create order from cart |
| GET | `/api/orders` | Yes | User orders |
| GET | `/api/addresses` | Yes | User addresses |
| POST | `/api/addresses` | Yes | Create address |
| PUT | `/api/addresses/:id` | Yes | Update address |
| DELETE | `/api/addresses/:id` | Yes | Delete address |
| PATCH | `/api/addresses/:id/default` | Yes | Set default address |
| POST | `/api/products/:id/reviews` | Yes | Create product review |
| PUT | `/api/reviews/:id` | Yes | Update review rating |
| DELETE | `/api/reviews/:id` | Yes | Delete own review |
| GET | `/api/coupons` | Yes | Available unused active coupons |
| GET | `/api/coupons/:id` | Yes | Available unused active coupon detail |
| POST | `/api/coupons/validate` | Yes | Validate coupon(s) against current cart |
| GET | `/api/notifications` | Yes | User notifications with pagination |
| GET | `/api/notifications/unread-count` | Yes | Unread notification count for red badge UI |
| PATCH | `/api/notifications/:id/read` | Yes | Mark one notification as read |
| PATCH | `/api/notifications/read-all` | Yes | Mark all notifications as read |
| GET | `/api/notifications/ws` | Yes | WebSocket stream for real-time notifications |

### Admin APIs

| Method | Path | Auth | Admin | Description |
|---|---|---:|---:|---|
| GET | `/api/admin/orders` | Yes | Yes | All orders |
| PUT | `/api/admin/orders/:id` | Yes | Yes | Update order status |
| POST | `/api/admin/products` | Yes | Yes | Create product |
| PUT | `/api/admin/products/:id` | Yes | Yes | Update product |
| DELETE | `/api/admin/products/:id` | Yes | Yes | Soft delete product |
| PATCH | `/api/admin/products/:id/stock` | Yes | Yes | Update stock |
| POST | `/api/admin/categories` | Yes | Yes | Create category |
| PUT | `/api/admin/categories/:id` | Yes | Yes | Update category |
| DELETE | `/api/admin/categories/:id` | Yes | Yes | Soft delete category |
| GET | `/api/admin/coupons` | Yes | Yes | List coupons |
| POST | `/api/admin/coupons` | Yes | Yes | Create coupon |
| GET | `/api/admin/coupons/:id` | Yes | Yes | Coupon detail |
| PUT | `/api/admin/coupons/:id` | Yes | Yes | Update coupon |
| DELETE | `/api/admin/coupons/:id` | Yes | Yes | Disable coupon |
| PATCH | `/api/admin/coupons/:id/status` | Yes | Yes | Activate/deactivate coupon |

## 7. Authentication Flow

1. User registers via `POST /auth/register`, or logs in via `POST /auth/login`.
2. Backend returns JWT token and user info.
3. Client sends token on protected routes:

```http
Authorization: Bearer <token>
```

4. Middleware validates JWT and stores `user_id` and `is_admin` in Gin context.
5. Admin routes additionally require `is_admin=true`.
6. WebSocket notification clients may also pass the same JWT as `?token=<token>` when connecting to `/api/notifications/ws`.

Seed admin:

- Username: `admin`
- Password: `password`

## 8. Cart & Checkout Flow

1. User logs in.
2. User adds products:

```json
{
  "items": [
    { "product_id": 1, "quantity": 2 }
  ]
}
```

3. Backend checks product exists and requested quantity does not exceed stock.
4. User views cart via `GET /api/cart`.
5. User selects/creates address.
6. Optional: user validates coupon(s) with current cart.
7. User creates order:

```json
{
  "address_id": 1,
  "coupon_codes": ["SALE10", "FIX10K"]
}
```

8. Backend recalculates product prices, subtotal, discounts, final total, stock changes, and coupon redemptions.
9. Order creation runs in a transaction. If any step fails, order is not created and cart is not cleared.
10. Cart is cleared only after successful order creation.

## 9. Coupon Flow

### Admin coupon setup

Admin creates coupon:

```json
{
  "code": "SALE10",
  "name": "Giam 10%",
  "description": "Giam 10% cho don tu 500k",
  "discount_type": "percentage",
  "discount_value": 10,
  "min_order_amount": 500000,
  "max_discount_amount": 100000,
  "usage_limit": 100,
  "usage_limit_per_user": 1,
  "start_at": "2026-05-01T00:00:00Z",
  "end_at": "2026-05-31T23:59:59Z",
  "is_active": true
}
```

Rules:

- `code` is trimmed and uppercased before saving.
- `discount_type` must be `percentage` or `fixed_amount`.
- Percentage value must be from 1 to 100.
- Fixed amount must be greater than 0.
- `usage_limit_per_user` defaults to 1 if omitted.

### Customer coupon display

`GET /api/coupons` requires login and returns only:

- active coupons
- coupons the current user has not redeemed yet

If user already used a coupon, `GET /api/coupons/:id` returns `coupon not found`.

### Validate coupon(s)

Single coupon:

```json
{
  "code": "SALE10"
}
```

Multiple coupons:

```json
{
  "coupon_codes": ["SALE10", "FIX10K"]
}
```

Backend uses the current cart to calculate:

- `subtotal_amount`
- `discount_amount`
- `total_amount`

### Apply coupon(s) to order

Order API supports old and new payloads.

Backward compatible single coupon:

```json
{
  "address_id": 1,
  "coupon_code": "SALE10"
}
```

Multiple coupons:

```json
{
  "address_id": 1,
  "coupon_codes": ["SALE10", "FIX10K"]
}
```

Backend validates coupon(s) again during order creation and enforces:

- active status
- start/end validity
- min order amount
- global usage limit
- usage limit per user
- duplicate coupon code rejection
- total discount cannot exceed subtotal

## 10. Installation & Setup

Prerequisites:

- Go 1.21+
- Docker and Docker Compose
- PostgreSQL if not using Docker

Install dependencies:

```bash
go mod tidy
```

Start database services:

```bash
docker-compose up -d
```

If you want to create the PostgreSQL database/schema manually instead of relying on GORM auto migration:

```bash
psql -U postgres -f scripts/create_database.sql
```

Run backend:

```bash
go run cmd/server/main.go
```

Run tests:

```bash
env GOCACHE=/private/tmp/shop-go-cache go test ./...
```

## 11. Environment Variables

Create `.env` from `.env.example`:

```bash
cp .env.example .env
```

Variables:

| Name | Example | Description |
|---|---|---|
| `PORT` | `8080` | Backend HTTP port |
| `DATABASE_URL` | `postgres://user:password@localhost:5432/order_demo?sslmode=disable` | PostgreSQL connection URL |
| `JWT_SECRET` | `your-secret-key` | JWT signing secret |
| `POSTGRES_DB` | `order_demo` | Docker Postgres database |
| `POSTGRES_USER` | `user` | Docker Postgres user |
| `POSTGRES_PASSWORD` | `password` | Docker Postgres password |
| `PGADMIN_DEFAULT_EMAIL` | `admin@example.com` | pgAdmin login email |
| `PGADMIN_DEFAULT_PASSWORD` | `admin` | pgAdmin login password |

## 12. Docker Setup

This repository includes `docker-compose.yml` for local PostgreSQL and pgAdmin.

Start services:

```bash
docker-compose up -d
```

Stop services:

```bash
docker-compose down
```

View PostgreSQL logs:

```bash
docker-compose logs -f postgres
```

Services:

- PostgreSQL: `localhost:5432`
- pgAdmin: `http://localhost:5050`

There is no application Dockerfile in the current repository snapshot; run the Go backend locally with `go run cmd/server/main.go`.

## 13. Screenshots / Demo

The backend can be tested with the companion UI repository:

```bash
git clone https://github.com/kietnguyen2003/ShopCauLongUI.git
cd ShopCauLongUI
git fetch
git checkout ddd+clean
npm i
npm run dev
```

Recommended local setup:

1. Start PostgreSQL and pgAdmin for this backend:

```bash
docker-compose up -d
```

2. Start this backend:

```bash
go run cmd/server/main.go
```

3. Start the frontend from `ShopCauLongUI`:

```bash
npm run dev
```

The backend CORS config currently allows:

- `http://localhost:3000`
- `http://localhost:3001`

Demo flows you can test:

- Register/login and call protected APIs with JWT.
- Add products to cart and create an order.
- Admin creates two coupons, user validates both, and creates an order with multiple coupons.
- User coupon list hides coupons already redeemed by that user.
- Cancel a pending order and confirm product stock is restocked.
- Create a review only after buying the product.

### UI screenshots

The `snapshot/` folder contains UI screenshots named by screen/function.

#### Trang chu

![Trang chu](snapshot/trang-chu.png)

#### Dang ky

![Dang ky](snapshot/dang-ky.png)

#### Dang nhap

![Dang nhap](snapshot/dang-nhap.png)

#### Dang nhap admin

![Dang nhap admin](snapshot/dang-nhap-admin.png)

#### Danh sach san pham

![Danh sach san pham](snapshot/danh-sach-san-pham.png)

#### Gio hang

![Gio hang](snapshot/gio-hang.png)

#### Quan li dia chi

![Quan li dia chi](snapshot/quan-li-address.png)

#### Quan li don hang

![Quan li don hang](snapshot/quan-li-don-hang.png)

#### Quan li don hang admin

![Quan li don hang admin](snapshot/quan-li-don-hang-admin.png)

#### Quan li san pham admin

![Quan li san pham admin](snapshot/quan-li-san-pham-admin.png)

#### Quan li categories

![Quan li categories](snapshot/quan-li-categories.png)

#### Quan li coupon

![Quan li coupon](snapshot/quan-li-coupon.png)

## 14. Future Improvements

- Add automated integration tests with a test database.
- Add OpenAPI/Swagger documentation generated from handlers/models.
- Add pagination to reviews and orders.
- Add order detail endpoint for users.
- Improve coupon stacking rules, for example allow/disallow combining specific coupon types.
- Add payment integration and payment status.
- Add audit logs for admin coupon/product changes.
- Add soft delete timestamps consistently across entities.
- Add stock row locking during order creation to better prevent overselling under concurrency.
- Add Dockerfile for the Go API service.
