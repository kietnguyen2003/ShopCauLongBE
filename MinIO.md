# MinIO Usage Guide

This project uses MinIO as local S3-compatible object storage.

MinIO is used to store uploaded files such as product images. The database should store only the returned file URL or object key, not the binary file content.

## 1. Why MinIO

Before MinIO, product images were usually represented by external image URLs.

With MinIO, the project can support this flow:

```text
Admin uploads image
-> Backend receives multipart file
-> Backend uploads file to MinIO
-> MinIO stores the object
-> Backend returns object_key and public URL
-> Admin saves the URL into product.image
```

Benefits:

- Keeps uploaded files out of PostgreSQL.
- Works like Amazon S3, so the project can migrate to S3 later.
- Good for product images, avatars, invoices, exports, and backups.
- Easy to run locally with Docker Compose.

## 2. Docker Compose Service

MinIO is configured in:

```text
docker-compose.yml
```

Service:

```yaml
minio:
  image: minio/minio:latest
  container_name: order-backend-minio
  environment:
    MINIO_ROOT_USER: ${MINIO_ACCESS_KEY:-minioadmin}
    MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY:-minioadmin}
  command: server /data --console-address ":9001"
  ports:
    - "9000:9000"
    - "9001:9001"
  volumes:
    - minio_data:/data
```

Ports:

```text
9000 -> MinIO S3-compatible API
9001 -> MinIO web console
```

Local URLs:

```text
MinIO API:     http://localhost:9000
MinIO Console: http://localhost:9001
```

Default credentials:

```text
Username: minioadmin
Password: minioadmin
```

## 3. Environment Variables

Configured in:

```text
.env.example
```

Variables:

```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=shop-cau-long
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000
```

Meaning:

- `MINIO_ENDPOINT`: MinIO API endpoint used by the backend.
- `MINIO_ACCESS_KEY`: access key for MinIO.
- `MINIO_SECRET_KEY`: secret key for MinIO.
- `MINIO_BUCKET`: bucket used by this backend.
- `MINIO_USE_SSL`: `true` for HTTPS, `false` for local HTTP.
- `MINIO_PUBLIC_URL`: base URL used to build public file URLs.

## 4. Backend Config

Config is loaded in:

```text
internal/infrastructure/config/config.go
```

The app config has:

```go
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}
```

This is loaded inside:

```go
func Load() *Config
```

Default local values are used if environment variables are missing.

## 5. Storage Implementation

MinIO storage code lives in:

```text
internal/infrastructure/storage/minio_storage.go
```

Main type:

```go
type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}
```

Responsibilities:

- Create MinIO client.
- Ensure bucket exists.
- Set public read policy.
- Upload object.
- Return bucket, object key, public URL, content type, and size.

## 6. Startup Flow

In:

```text
cmd/server/main.go
```

The backend initializes storage:

```go
objectStorage, err := storage.NewMinIOStorage(cfg.MinIO)
if err != nil {
	log.Fatal("Failed to initialize object storage:", err)
}
```

Then creates the upload service:

```go
uploadService := upload.NewService(objectStorage)
uploadHandler := httpHandlers.NewUploadHandler(uploadService)
```

Then registers the route:

```go
admin.POST("/uploads/images", uploadHandler.UploadImage)
```

Full route:

```text
POST /api/admin/uploads/images
```

This route is admin-only because it is inside:

```go
admin := api.Group("/admin")
admin.Use(authHandler.AdminMiddleware())
```

## 7. Upload API

Endpoint:

```http
POST /api/admin/uploads/images
Authorization: Bearer <admin_token>
Content-Type: multipart/form-data
```

Form fields:

```text
file   required image file
folder optional folder name, default products
```

Supported file types:

```text
image/jpeg
image/png
image/webp
image/gif
```

Max file size:

```text
5 MB
```

## 8. Curl Example

Login as admin first:

```bash
curl -X POST http://localhost:8080/auth/admin-login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

Copy the returned token, then upload an image:

```bash
curl -X POST http://localhost:8080/api/admin/uploads/images \
  -H "Authorization: Bearer <admin_token>" \
  -F "file=@/path/to/product.webp" \
  -F "folder=products"
```

Example response:

```json
{
  "code": 201,
  "msg": "Upload image successfully",
  "data": {
    "bucket": "shop-cau-long",
    "object_key": "products/56f4f6a2-1c77-4b0c-a5fd-72d2c56ad70e.webp",
    "url": "http://localhost:9000/shop-cau-long/products/56f4f6a2-1c77-4b0c-a5fd-72d2c56ad70e.webp",
    "content_type": "image/webp",
    "size": 123456
  }
}
```

## 9. How Object Keys Are Built

Object key creation happens in:

```go
func buildObjectKey(folder, fileName string) string
```

Example input:

```text
folder = products
fileName = racket.webp
```

Example output:

```text
products/56f4f6a2-1c77-4b0c-a5fd-72d2c56ad70e.webp
```

The backend uses a UUID to avoid filename collisions.

## 10. Public URL Format

The backend builds URLs like:

```text
<MINIO_PUBLIC_URL>/<MINIO_BUCKET>/<object_key>
```

Example:

```text
http://localhost:9000/shop-cau-long/products/abc.webp
```

This URL can be saved into:

```text
products.image
categories.image
```

## 11. Bucket Creation

Before uploading, the storage layer calls:

```go
EnsureBucket(ctx)
```

It checks whether the bucket exists:

```go
BucketExists(ctx, bucket)
```

If missing, it creates the bucket:

```go
MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
```

## 12. Public Read Policy

The current implementation sets a public read policy:

```go
SetBucketPolicy(ctx, bucket, policy)
```

This allows browser/frontend clients to load image URLs directly.

Policy effect:

```text
Anyone can GET objects inside the bucket.
```

This is convenient for public product images.

For private files such as invoices, user documents, or internal exports, do not use public URLs. Use presigned URLs instead.

## 13. Error Cases

The upload API can return:

```text
400 file is required
400 unsupported file type
400 file size exceeds limit
500 storage/minio error
```

Common causes:

- MinIO container is not running.
- Wrong MinIO credentials.
- Bucket policy cannot be set.
- Uploaded file is not an allowed image type.
- Uploaded file is larger than 5 MB.

## 14. How To Run Locally

Start services:

```bash
docker-compose up -d
```

Open MinIO console:

```text
http://localhost:9001
```

Login:

```text
minioadmin / minioadmin
```

Start backend:

```bash
go run cmd/server/main.go
```

Upload image using the API.

Then check the bucket in the MinIO console:

```text
shop-cau-long
```

## 15. Where It Fits In The Architecture

Current structure:

```text
HTTP handler
-> application/upload service
-> infrastructure/storage MinIO client
-> MinIO bucket
```

Files:

```text
internal/interfaces/http/upload_handler.go
internal/interfaces/http/upload_models.go
internal/application/upload/dto.go
internal/application/upload/ports.go
internal/application/upload/service.go
internal/infrastructure/storage/minio_storage.go
```

This keeps storage details out of handlers and business logic.

## 16. Suggested Frontend Flow

For product image upload:

```text
Admin selects image
-> Frontend sends multipart upload to /api/admin/uploads/images
-> Backend returns URL
-> Frontend puts URL into product create/update payload
-> Backend saves URL in products.image
```

Example product create after upload:

```json
{
  "name": "Vot Cau Long Yonex",
  "description": "Mo ta san pham",
  "price": 1500000,
  "stock": 10,
  "image": "http://localhost:9000/shop-cau-long/products/abc.webp",
  "category_id": 2
}
```

## 17. Production Notes

For production:

- Use strong MinIO credentials.
- Do not expose the MinIO console publicly.
- Use HTTPS.
- Consider private buckets plus presigned URLs for non-public files.
- Add file scanning/validation if users can upload content.
- Consider separating public product images from private user files by bucket or prefix.
