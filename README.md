# Go Hexagonal Architecture Template

Template Go modular monolith dengan Echo, PostgreSQL, GORM, golang-migrate, cleanenv, slog, JWT, OpenTelemetry, Prometheus, dan testing berlapis.

## Important boundary

`domain.User` tidak memiliki tag GORM atau import GORM. Model persistence hanya hidup di `internal/user/adapter/outbound/postgres`; adapter memetakan model tersebut ke domain entity. Schema dikelola oleh SQL migration eksplisit. API tidak pernah memanggil `AutoMigrate()`.

## Local setup

1. Salin `.env.example` ke `.env` dan isi `JWT_SECRET` dengan secret minimal 32 byte.
2. Jalankan `make docker-up`.
3. Jalankan `make migrate-up`.
4. Jalankan `make run`.

Health endpoints: `/health/live`, `/health/ready`; metrics: `/metrics`; API: `/api/v1/users`.
