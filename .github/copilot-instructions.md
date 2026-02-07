# Copilot Instructions for Qashare

Qashare is a full-stack expense-splitting app with a Go/Gin REST API backend (`server/`) and a Flutter mobile client (`client/`). See per-directory instructions for detailed guidance:

- [`server/.github/copilot-instructions.md`](../server/.github/copilot-instructions.md)
- [`client/.github/copilot-instructions.md`](../client/.github/copilot-instructions.md)

## CI Checks

- **Go format** (`gofmt`), **Go tests**, **Swagger doc sync** (`swag init` then check for uncommitted changes)
- **Dart format**, **Flutter tests**
- **PR titles** must follow [Conventional Commits](https://www.conventionalcommits.org/)

## Repo-Wide Conventions

- **Git hooks**: `.githooks/` uses a dispatcher pattern (`pre-commit.d/`, etc.). Set up with `git config core.hooksPath .githooks`.
- **Maintainers**: Server — @pranaovs, Client — @sasvat007.
