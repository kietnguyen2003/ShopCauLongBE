# WebSocket Notification Flow

This document explains the current WebSocket notification implementation in this project, especially how `AddClient`, `RemoveClient`, `Lock`, and `Unlock` work inside `internal/interfaces/http/notification_hub.go`.

## Goal

The notification WebSocket lets the backend push order status notifications to users in real time.

Current scenario:

1. A user opens a WebSocket connection to `/api/notifications/ws?token=<jwt>`.
2. The backend authenticates the JWT and upgrades the HTTP request to WebSocket.
3. The connection is registered in `NotificationHub`.
4. When an admin confirms or cancels an order, the backend creates a notification in PostgreSQL.
5. If the target user is online, the notification is pushed through every active WebSocket connection of that user.
6. If the user is offline, the notification still exists in the database and can be fetched later through REST APIs.

## Main Data Structure

```go
type NotificationHub struct {
	mu      sync.RWMutex
	clients map[uint]map[*websocket.Conn]struct{}
}
```

Meaning:

- `mu`: protects concurrent access to the `clients` map.
- `clients`: stores active WebSocket connections grouped by `userID`.
- `map[uint]...`: key is the authenticated user ID.
- `map[*websocket.Conn]struct{}`: set of active WebSocket connections for that user.

The nested map allows one user to have multiple online sessions at the same time, for example:

- browser tab 1
- browser tab 2
- mobile app
- another device

All of them can receive the same notification.

## Why Locking Is Needed

Gin handles requests concurrently. That means multiple goroutines can touch the hub at the same time:

- User A connects through WebSocket.
- User A disconnects.
- Admin updates an order and broadcasts a notification.
- Another browser tab connects at the same time.

Go maps are not safe for concurrent read/write access. Without a mutex, the app can panic with:

```text
fatal error: concurrent map read and map write
```

So the hub uses `sync.RWMutex`.

## `sync.RWMutex`

`RWMutex` has two lock modes:

```go
h.mu.Lock()
h.mu.Unlock()
```

Use this when writing to shared data.

```go
h.mu.RLock()
h.mu.RUnlock()
```

Use this when only reading shared data.

Rules:

- Many readers can hold `RLock` at the same time.
- Only one writer can hold `Lock`.
- While a writer holds `Lock`, readers must wait.
- While readers hold `RLock`, a writer must wait.

## AddClient

```go
func (h *NotificationHub) AddClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[userID][conn] = struct{}{}
}
```

### What It Does

`AddClient` registers a new WebSocket connection for a user.

Example:

```text
User 5 opens /api/notifications/ws
-> AddClient(5, conn)
```

After that, the hub may look like:

```text
clients = {
  5: {
    connA: {}
  }
}
```

If the same user opens another tab:

```text
clients = {
  5: {
    connA: {},
    connB: {}
  }
}
```

### Why It Uses `Lock`

`AddClient` modifies the shared `clients` map:

```go
h.clients[userID] = make(map[*websocket.Conn]struct{})
h.clients[userID][conn] = struct{}{}
```

Because this is a write operation, it must use:

```go
h.mu.Lock()
defer h.mu.Unlock()
```

### Why `defer Unlock`

`defer h.mu.Unlock()` guarantees the mutex is released when the function returns.

This keeps the code safe even if the function grows later and has multiple return paths.

## RemoveClient

```go
func (h *NotificationHub) RemoveClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		return
	}
	delete(h.clients[userID], conn)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}
```

### What It Does

`RemoveClient` unregisters a WebSocket connection when the user disconnects or when sending to that connection fails.

Example before removing:

```text
clients = {
  5: {
    connA: {},
    connB: {}
  }
}
```

If `connA` disconnects:

```text
RemoveClient(5, connA)
```

After removing:

```text
clients = {
  5: {
    connB: {}
  }
}
```

If `connB` also disconnects:

```text
clients = {}
```

The user key is removed when no connections remain. This keeps memory clean.

### Why It Uses `Lock`

`RemoveClient` deletes from the map:

```go
delete(h.clients[userID], conn)
delete(h.clients, userID)
```

Deleting from a map is a write operation, so it needs the exclusive write lock:

```go
h.mu.Lock()
defer h.mu.Unlock()
```

## BroadcastToUser

```go
func (h *NotificationHub) BroadcastToUser(userID uint, notification appNotification.NotificationResponse) {
	h.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(h.clients[userID]))
	for conn := range h.clients[userID] {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	payload := toNotificationHTTPResponse(notification)
	for _, conn := range connections {
		if err := conn.WriteJSON(payload); err != nil {
			h.RemoveClient(userID, conn)
			_ = conn.Close()
		}
	}
}
```

### What It Does

`BroadcastToUser` sends one notification to every active WebSocket connection of a user.

Example:

```text
Admin confirms order #12 for user 5
-> backend creates notification
-> BroadcastToUser(5, notification)
-> connA receives notification
-> connB receives notification
```

### Why It Uses `RLock`

At first, the function only needs to read the `clients` map:

```go
for conn := range h.clients[userID] {
	connections = append(connections, conn)
}
```

So it uses:

```go
h.mu.RLock()
h.mu.RUnlock()
```

This allows multiple broadcasts to read the map concurrently when no one is writing.

### Why It Copies Connections First

The function copies the user's current connections into a slice:

```go
connections := make([]*websocket.Conn, 0, len(h.clients[userID]))
for conn := range h.clients[userID] {
	connections = append(connections, conn)
}
```

Then it unlocks before writing to WebSocket connections:

```go
h.mu.RUnlock()
```

This is important because `conn.WriteJSON` can be slow or fail. If the hub kept the lock while writing, other users trying to connect or disconnect would be blocked unnecessarily.

The current approach keeps the critical section short:

1. Lock only long enough to copy pointers.
2. Unlock.
3. Write messages outside the lock.

### What Happens If Write Fails

If sending fails:

```go
if err := conn.WriteJSON(payload); err != nil {
	h.RemoveClient(userID, conn)
	_ = conn.Close()
}
```

The hub assumes the connection is dead or unhealthy.

It then:

1. Removes the connection from `clients`.
2. Closes the WebSocket connection.

## WebSocket Handler Lifecycle

The WebSocket endpoint is implemented in `StreamNotifications`.

Simplified flow:

```go
func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.hub.AddClient(userID.(uint), conn)
	defer func() {
		h.hub.RemoveClient(userID.(uint), conn)
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
```

### Step-by-Step

1. `AuthMiddleware` validates the JWT.
2. The handler reads `user_id` from Gin context.
3. The HTTP connection is upgraded to WebSocket.
4. `AddClient` stores the connection in the hub.
5. A `defer` is registered to remove and close the connection later.
6. The handler enters a read loop.
7. If the browser closes the tab, network drops, or the client disconnects, `ReadMessage` returns an error.
8. The function returns.
9. The deferred cleanup runs:

```go
h.hub.RemoveClient(userID.(uint), conn)
_ = conn.Close()
```

## Why The Read Loop Exists

Even if the backend mainly sends notifications, the server still needs to read from the WebSocket connection.

The read loop helps detect when the client disconnects:

```go
for {
	if _, _, err := conn.ReadMessage(); err != nil {
		return
	}
}
```

Without this loop, the backend might keep dead connections in memory for too long.

## Complete Notification Flow

```text
User connects:
GET /api/notifications/ws?token=<jwt>
-> AuthMiddleware validates token
-> StreamNotifications upgrades connection
-> AddClient(userID, conn)

Admin updates order:
PUT /api/admin/orders/:id
-> orderService.UpdateOrderStatus(...)
-> notificationService.CreateOrderStatusNotification(...)
-> notificationRepo.Create(...)
-> notificationHub.BroadcastToUser(userID, notification)

User online:
-> WebSocket receives notification immediately

User offline:
-> no active connection exists
-> notification remains in PostgreSQL
-> frontend later calls GET /api/notifications/unread-count
-> UI shows red badge
```

## Current Limitations

This implementation stores active WebSocket connections in memory.

That is fine for:

- local development
- one backend instance
- demo or portfolio project
- small deployments

If the app runs multiple backend instances, one user's WebSocket may be connected to instance A while the admin request is handled by instance B. In that case, instance B cannot see instance A's in-memory connection map.

To support multiple instances, add a shared message broker such as:

- Redis Pub/Sub
- Kafka
- NATS

Then each backend instance can subscribe to notification events and push to its own connected clients.

## Frontend Usage Example

```js
const socket = new WebSocket(`ws://localhost:8080/api/notifications/ws?token=${token}`);

socket.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log("New notification:", notification);
};

socket.onclose = () => {
  console.log("Notification socket closed");
};
```

Unread badge:

```text
GET /api/notifications/unread-count
```

Notification list:

```text
GET /api/notifications?page=1&limit=20
```

Mark as read:

```text
PATCH /api/notifications/:id/read
```

Mark all as read:

```text
PATCH /api/notifications/read-all
```
