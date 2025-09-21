# Multi-User Architecture Plan

## Data Model Changes

- Introduce a `users` table (PostgreSQL) with columns:
  - `id UUID PRIMARY KEY`
  - `email VARCHAR(320) UNIQUE NOT NULL`
  - `password_hash TEXT NOT NULL`
  - `first_name VARCHAR(100) NOT NULL`
  - `last_name VARCHAR(100) NOT NULL`
  - `created_at TIMESTAMPTZ DEFAULT now()`
  - `updated_at TIMESTAMPTZ DEFAULT now()`
- Replace the existing `config` table with `user_settings` keyed by `user_id`.
- Add `user_id` foreign keys to `expenses` and `recurring_expenses` tables.
- JSON storage backend will be deprecated for multi-user operation; enforce PostgreSQL requirement.

## Authentication Flow

- Use JWT access tokens signed with `JWT_SECRET` and persisted in Redis for session invalidation.
- Support two roles (`admin`, `user`). Admin tokens include the elevated role claim and are required for management endpoints.
- Expose endpoints inspired by go-auth-boilerplate:
  - `POST /api/v1/user/signup`
  - `POST /api/v1/user/login`
  - `POST /api/v1/user/logout`
  - `GET /api/v1/session` (current user info)
  - `PATCH /api/v1/user/update_password`
- Provide admin-only APIs:
  - `GET /api/v1/admin/users`
  - `PATCH /api/v1/admin/users/role`
- Support self-service profile updates via `/api/v1/user/profile` (GET/PATCH) and password changes at `/api/v1/user/update_password`.
- Successful password updates or email changes invalidate the current JWT, forcing a new login.
- Require `Authorization: Bearer <token>` header for protected API routes.
- Replace `RequireAPIAuth` / `RequireWebAuth` with middleware that validates JWTs and attaches the `user_id` to the request context.

## Component Responsibilities

- `internal/user` package: repository + service for managing users, hashing passwords, validation.
- `internal/auth` package: JWT manager, Redis session store, middleware helpers.
- `internal/storage`: refactored interface to accept `userID string` and operate on per-user data; Postgres implementation updated to apply `WHERE user_id = $1`.
- `cmd/expenseowl`: server initialization now wires Redis and user services, configures routes, and ensures multi-user endpoints are registered alongside SPA handlers.

## Frontend Updates

- Add a Vue view for login/register that interacts with JSON endpoints.
- Store JWT token client-side (e.g., `localStorage`) and attach it to fetch requests via an interceptor.
- Handle 401 responses by clearing token and redirecting to login.
- Remove legacy form-based `/login` template.

## Migration Strategy

- Provide a one-time script to migrate existing single-user data by creating a default user and updating records with the new `user_id`.
- Update documentation to highlight the new environment variables: `JWT_SECRET`, Redis settings, etc.
