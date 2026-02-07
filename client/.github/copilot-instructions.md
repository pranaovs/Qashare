# Copilot Instructions — Qashare Client

Flutter mobile app for tracking shared expenses. Communicates with the Go backend via REST API.

## Build & Run

```sh
flutter pub get                          # install dependencies
flutter run                              # run app
flutter test                             # run all tests
flutter test test/specific_test.dart     # run a single test file
dart format . --set-exit-if-changed      # check formatting (CI enforced)
flutter analyze                          # run static analysis
```

Uses FVM (Flutter Version Manager) pinned to `stable` channel (see `.fvmrc`).
Linting rules from `package:flutter_lints/flutter.yaml` (see `analysis_options.yaml`).

## Build-Time Configuration

API settings are injected at build time via `dart_defines.json`:
- `API_HOST` — server URL (default: `https://qashare.pranaovs.me`)
- `API_BASE_PATH` — API base path (default: `/api`)

These are read in `lib/Config/api_config.dart` using `String.fromEnvironment`.

## Architecture

- **`lib/Config/`** — App configuration.
  - `api_config.dart` — `ApiConfig` class with static getters for `baseUrl` and `server`. Supports runtime override via `setServer()`.
  - `token_storage.dart` — `TokenStorage` wraps `flutter_secure_storage` for JWT persistence.

- **`lib/Service/api_service.dart`** — Single `ApiService` class with static methods for all API calls (register, login, groups, expenses, settlements, etc.). Each method returns a typed result object.

- **`lib/Models/`** — Data models. Each model file typically contains:
  - A data class with a `factory Model.fromJson(Map<String, dynamic>)` constructor
  - A result wrapper class with `factory Result.success(...)` / `factory Result.error(String)` named constructors

- **`lib/Screens/`** — UI pages as `StatefulWidget`s. Navigation uses named routes defined in `main.dart`.

## Key Conventions

- **Material 3 + dynamic color**: Uses `dynamic_color` package for platform-adaptive theming. Falls back to blue seed color.
- **Result pattern**: API calls return result objects with `isSuccess`, data fields, and `errorMessage` — not exceptions. Screens check `result.isSuccess` before accessing data.
- **Auth flow**: JWT token stored/retrieved via `TokenStorage`. Authenticated requests pass `"Authorization": "Bearer $token"` header. The `AuthcheckPage` screen handles initial routing based on token presence.
- **Named routes**: All routes defined in `main.dart`'s `routes` map. Some routes receive arguments via `ModalRoute.of(context)!.settings.arguments`.
