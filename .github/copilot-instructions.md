# Qashare - Copilot Instructions

Qashare is a full-stack expense tracking and bill splitting application with a Go backend (Gin framework) and Flutter mobile client.

## Build & Run Commands

### Server (Go Backend)

```bash
cd server

# Install dependencies
go get

# Run the server (requires PostgreSQL)
go run .

# Regenerate Swagger docs after changing API annotations
swag init

# Run with Docker (includes PostgreSQL)
docker compose up
```

### Client (Flutter App)

```bash
cd client

# Install dependencies
flutter pub get

# Run the app
flutter run

# Run tests
flutter test

# Run a single test file
flutter test test/path/to/test_file.dart
```

## Server Architecture

### Package Structure

```
server/
├── main.go              # Entry point, server initialization
├── routes/              # HTTP layer
│   ├── routes.go        # Central route registration
│   ├── auth.go          # Auth route group registration
│   ├── groups.go        # Groups route group registration
│   ├── handlers/        # Request handlers (controllers)
│   ├── middleware/      # Auth, group membership middleware
│   └── apierrors/       # HTTP-specific errors with status codes
├── db/                  # Database layer (PostgreSQL via pgx)
│   ├── connection.go    # Pool management, health checks
│   ├── errors.go        # DBError type, sentinel errors
│   ├── helpers.go       # Transactions, batch queries, utilities
│   ├── users.go         # User CRUD operations
│   ├── groups.go        # Group operations
│   └── expenses.go      # Expense operations
├── models/              # Shared data structures
├── utils/               # Validation, auth, helpers
│   ├── authentication.go # JWT generation/validation, password hashing
│   ├── validation.go    # Name, email validation
│   ├── api.go           # Response helpers (SendError, SendJSON)
│   └── errors.go        # UtilsError type
├── apperrors/           # Cross-layer error mapping
│   ├── interface.go     # AppError interface
│   └── mapper.go        # MapError function
└── migrations/          # SQL migration files
```

### Route Registration Pattern

Routes are organized by resource with dedicated registration functions:

```go
// routes/routes.go - Central registration
func RegisterRoutes(router *gin.Engine, pool *pgxpool.Pool) {
    RegisterAuthRoutes(router.Group("/auth"), pool)
    RegisterGroupsRoutes(router.Group("/groups"), pool)
    RegisterExpensesRoutes(router.Group("/expenses"), pool)
}

// routes/groups.go - Resource-specific registration
func RegisterGroupsRoutes(router *gin.RouterGroup, pool *pgxpool.Pool) {
    handler := handlers.NewGroupsHandler(pool)
    
    router.POST("/", middleware.RequireAuth(), handler.Create)
    router.GET("/:id", middleware.RequireAuth(), middleware.RequireGroupMember(pool), handler.GetGroup)
    router.POST("/:id/members", middleware.RequireAuth(), middleware.RequireGroupAdmin(pool), handler.AddMembers)
}
```

### Handler Pattern

Handlers are structs with a database pool, created via constructor:

```go
type AuthHandler struct {
    pool *pgxpool.Pool
}

func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
    return &AuthHandler{pool: pool}
}

// Handler methods use Swagger annotations for API docs
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{name=string,email=string,password=string} true "Registration"
// @Success 201 {object} models.User
// @Failure 400 {object} apierrors.AppError
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
    // Implementation
}
```

### Error Handling System

The codebase uses a three-layer error architecture:

#### 1. Domain Errors (`db.DBError`, `utils.UtilsError`)

Internal errors with codes but no HTTP concerns. Defined as sentinel errors:

```go
// db/errors.go
var (
    ErrNotFound = &DBError{Code: "NOT_FOUND", Message: "record not found"}
    ErrDuplicateKey = &DBError{Code: "DUPLICATE_KEY", Message: "duplicate key violation"}
)

// Custom messages via .Msg() or .Msgf()
return ErrNotFound.Msgf("user with email %s not found", email)
```

#### 2. API Errors (`apierrors.AppError`)

HTTP-aware errors with status codes, defined in `routes/apierrors/models.go`:

```go
var (
    ErrBadRequest = New(http.StatusBadRequest, "BAD_REQUEST", "The request is invalid", nil)
    ErrUserNotFound = New(http.StatusNotFound, "USER_NOT_FOUND", "User does not exist", nil)
    ErrBadCredentials = New(http.StatusUnauthorized, "BAD_CREDENTIALS", "Incorrect credentials", nil)
)
```

#### 3. Error Mapping (`apperrors.MapError`)

Converts domain errors to API errors in handlers:

```go
user, err := db.GetUser(ctx, pool, userID)
if err != nil {
    utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
        db.ErrNotFound: apierrors.ErrUserNotFound,
    }))
    return
}
```

Custom messages automatically propagate through the chain - if a db function returns `ErrNotFound.Msg("custom message")`, that message reaches the API response.

### Middleware

#### Authentication Middleware

```go
// Extracts user ID from JWT, stores in context
router.GET("/me", middleware.RequireAuth(), handler.Me)

// In handler, retrieve with:
userID := middleware.MustGetUserID(c)  // Panics if missing (server misconfiguration)
// Or safely:
userID, ok := middleware.GetUserID(c)
```

#### Group Access Middleware

```go
// Check membership before allowing access
router.GET("/:id", middleware.RequireAuth(), middleware.RequireGroupMember(pool), handler.GetGroup)

// Check admin status for mutations
router.POST("/:id/members", middleware.RequireAuth(), middleware.RequireGroupAdmin(pool), handler.AddMembers)

// In handler, retrieve validated group ID:
groupID := middleware.MustGetGroupID(c)
```

### Database Layer Conventions

#### Function Signatures

All db functions follow this pattern:
```go
func GetUser(ctx context.Context, pool *pgxpool.Pool, userID string) (models.User, error)
func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *models.User) error
```

#### Transactions

Use `WithTransaction` for multi-step operations:

```go
err := db.WithTransaction(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
    // All operations use tx instead of pool
    _, err := tx.Exec(ctx, "INSERT INTO users ...")
    if err != nil {
        return err  // Triggers rollback
    }
    _, err = tx.Exec(ctx, "INSERT INTO guests ...")
    return err
})
// Commits on success, rolls back on error or panic
```

#### Batch Operations

```go
queries := []db.BatchQuery{
    {SQL: "INSERT INTO ...", Args: []any{...}},
    {SQL: "UPDATE ...", Args: []any{...}},
}
err := db.ExecuteInBatch(ctx, pool, queries)
```

#### Helper Functions

```go
// Check existence without loading full record
exists, err := db.RecordExists(ctx, pool, "users", "email = $1", email)

// Check if user is group member
isMember, err := db.MemberOfGroup(ctx, pool, userID, groupID)

// Verify all users are group members (for expense validation)
err := db.AllMembersOfGroup(ctx, pool, userIDs, groupID)
```

### Validation

Validation functions return cleaned values and errors with custom messages:

```go
// utils/validation.go
name, err := utils.ValidateName(request.Name)   // Returns trimmed, validated name
email, err := utils.ValidateEmail(request.Email) // Returns lowercase, validated email

// Errors include specific messages:
// ErrInvalidName.Msg("name must be 3-64 characters...")
// ErrInvalidEmail.Msg("email does not match required format")
```

### Response Helpers

```go
utils.SendJSON(c, http.StatusCreated, user)     // Send data with status
utils.SendData(c, data)                          // Send data with 200 OK
utils.SendOK(c, "operation successful")          // Send message with 200 OK
utils.SendError(c, err)                          // Send AppError or 500 for unknown errors
utils.SendAbort(c, http.StatusForbidden, "msg")  // Abort request chain
```

### JWT Authentication

```go
// Generate token (includes user_id claim, configurable expiry via JWT_EXPIRY env)
token, err := utils.GenerateJWT(userID)

// Extract user ID from Authorization header
userID, err := utils.ExtractUserID(c.GetHeader("Authorization"))

// Environment variables:
// JWT_SECRET - Secret key (random if not set, tokens won't survive restarts)
// JWT_EXPIRY - Token lifetime in seconds (default: 86400 = 24h)
```

### Migrations

SQL files in `migrations/` with naming convention `NNNN_description.up.sql`:
- `0001_init_schema.up.sql`
- `0002_guest_tracking.up.sql`

Migrations run automatically on startup via `db.Migrate(pool, migrationsDir)`.

## Client Architecture

### Package Structure

```
client/lib/
├── main.dart           # App entry, route definitions
├── Screens/            # UI pages
├── Models/             # Data models with Result pattern
├── Service/            # API communication
│   └── api_service.dart
└── Config/             # API URL, token storage
```

### API Service Pattern

Static methods returning Result types:

```dart
class ApiService {
  static Future<LoginResult> loginUser({required String email, required String password}) async {
    // Returns LoginResult.success(token: ...) or LoginResult.error("message")
  }
}
```

### Conventions

- Token storage via `flutter_secure_storage`
- Material Design 3 with dynamic color support
- Named routes defined in `main.dart`
