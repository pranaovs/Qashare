# Copilot Instructions — Qashare Server

Go REST API using Gin, PostgreSQL (pgx/v5), and JWT authentication.

## Build & Run

```sh
go get                                              # install dependencies
go run .                                            # run server (needs PostgreSQL)
go test -v ./...                                    # run all tests
go test -v -run TestFunctionName ./path/to/package  # run a single test
gofmt -l .                                          # check formatting (CI enforced)
swag init                                           # regenerate Swagger docs
```

Docker: `docker compose up` (edit `docker-compose.yml` for env vars first).

**Configuration**: All via environment variables with `.env` fallback (`godotenv`). See `config/load.go` for variable names and defaults. Duration values are in seconds.

## Architecture

Layered architecture with strict separation — requests flow through: **routes → middleware → handlers → db**.

- **`routes/`** — Route registration. Each resource file (e.g., `auth.go`, `expenses.go`) wires handler methods to paths under `{basepath}/v1/`.
- **`routes/handlers/`** — HTTP handlers. Each resource has a handler struct (e.g., `AuthHandler`, `ExpensesHandler`) constructed via `New*Handler(pool, config)`. Handlers parse requests, validate, call `db` functions, and respond.
- **`routes/middleware/`** — Gin middleware. `RequireAuth` extracts JWT user ID into context. `VerifyExpenseAccess`/`VerifyExpenseAdmin` check group membership and cache the fetched expense in `gin.Context` to avoid double-fetching.
- **`db/`** — Data access layer using `pgx/v5` with connection pooling (`pgxpool`). Raw SQL queries, no ORM.
- **`models/`** — Shared data structs with `json` and `db` struct tags. Struct embedding for detail types (e.g., `ExpenseDetails` embeds `Expense`).
- **`config/`** — Typed config structs populated from env vars.
- **`utils/`** — Validation, JWT, password hashing (bcrypt), structured logging (`slog` JSON), and response helpers.
- **`migrations/`** — Forward-only SQL migrations (`NNNN_name.up.sql`). Applied automatically on startup with SHA-256 checksum integrity verification.

## Three-Layer Error System

This is the most important pattern. Errors flow through three layers:

1. **Domain errors** (`db.DBError`, `utils.UtilsError`) — Sentinel errors with `Code` + `Message`. Use `.Msg()` / `.Msgf()` to create copies with custom messages. Both implement `apperrors.AppError` interface.
2. **API errors** (`routes/apierrors.AppError`) — HTTP-aware errors with `HTTPCode`, `MachineCode`, `Message`. Defined as package-level vars (e.g., `ErrBadRequest`, `ErrUserNotFound`).
3. **Error mapping** (`apperrors.MapError`) — Handlers map domain → API errors with inline maps:
   ```go
   utils.SendError(c, apperrors.MapError(err, map[error]*apierrors.AppError{
       db.ErrNotFound:    apierrors.ErrUserNotFound,
       db.ErrDuplicateKey: apierrors.ErrEmailAlreadyExists,
   }))
   ```
   Custom messages (via `.Msg()`) propagate through to the API response automatically.

## Key Conventions

- **Swagger annotations**: All handler functions must have godoc annotations for Swagger. Run `swag init` after changes — CI will fail if `docs/` is out of sync.
- **Response helpers**: Always use `utils.SendJSON`, `utils.SendError`, `utils.SendOK`, or `utils.SendData`. Never call `c.JSON` directly in handlers.
- **Middleware context**: Middleware sets values in `gin.Context` (`UserIDKey`, `ExpenseKey`, `GroupIDKey`). Handlers retrieve via `MustGetUserID(c)`, `MustGetExpenseID(c)`, etc. — these panic if the middleware is missing, triggering Gin's 500 recovery.
- **Database scanning**: Model structs use `db:"column_name"` tags. Use `db.ScanStruct` / `db.ScanStructs` for row scanning and `db.BuildSelectQuery` for dynamic column selection — avoid manual `Scan()` calls.
- **Transactions**: Use `db.WithTransaction(ctx, pool, func(ctx, tx) error { ... })` for atomic multi-statement operations.
- **Migrations**: Forward-only numbered SQL files. Never modify an already-applied migration — always create a new one.
