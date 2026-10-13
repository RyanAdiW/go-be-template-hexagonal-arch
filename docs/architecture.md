# Architecture guide and conventions

This document defines the architecture guide and conventions for this template.
Use it when adding modules, reviewing changes, or extending shared infrastructure.
The implementation status below distinguishes existing capabilities from planned
adoption; not all capabilities described here already exist. Code fragments are illustrative;
imports, constructors, and error handling must be integrated before use.

The decisions cover module boundaries, transactions, persistence mapping,
testing, platform lifecycle, and HTTP contracts. Business-specific requirements
remain outside the template. Retain this template's `domain` package name.

## Implementation status

| Decision | Current template | Adoption |
| --- | --- | --- |
| Feature-owned modular monolith | Present | Document ownership and cross-module dependency rules |
| Pure domain and application | Present for user use cases | Preserve; keep ORM, HTTP and provider details in adapters |
| Inbound/outbound application ports | Outbound repository port only, without directional package | Add inbound contracts and directional port packages |
| Composition builders and transport registrars | Wiring in `user/module.go` | Move wiring into per-module app builders |
| Cohesive repositories | One user repository | Keep cohesive; split by aggregate/query only as needed |
| Compile-time port assertions | Missing for user repository | Add for production port implementations |
| Context-scoped transaction port | Missing | Add framework-free port and GORM adapter |
| Explicit persistence mapping | Present | Extend conventions for nullability, dates and decimals |
| Separate HTTP DTO files | DTOs inside handler | Extract DTOs and mapping alongside handler |
| Shared response/error contract | Manual envelopes; platform knows user errors | Add typed helpers and module-owned error translation |
| Final-status request observability | Error rendered after metrics/logging | Render errors before observing status |
| Resource lifecycle | Basic shutdown | Add startup cleanup, connection timeout and joined errors |
| Config-driven request limits | Body limit hardcoded | Honor configured body limit and normalize CORS |
| Versioned OpenAPI | Missing | Embed contract and verify it against routes |
| Optional local dotenv loading | Environment-only | Optional convenience with environment precedence |
| Deterministic unit and adapter tests | Application fake test and live integration test | Extend behavioral coverage; retain live integration tests |
| Explicit migrations | Present | Preserve and document deployment/recovery procedures |
| External side-effect policy | Not yet needed by example | Document ordering, retries and idempotency before adding providers |

## Module ownership and dependency direction

Each business table has one writing module. Its domain owns business invariants,
its application exposes operations, and its outbound adapter implements storage.
Another module calls public application command/query ports, never concrete
repositories, persistence models, or another module's repository port.

A read-only reporting adapter may join tables from multiple modules when it
implements a dedicated query port. It gains no write ownership. A workflow
belongs to the module that owns the business capability; create a capability
module when no existing owner fits, rather than a generic service/orchestrator
module. Synchronous operations remain synchronous when callers need their
result or consistency guarantee.

Shared code contains technical primitives, not business entities or rules.
Only composition code imports concrete implementations to assemble modules.
Use `snake_case` for multiword directory names; package names may remain ordinary
lowercase Go identifiers. Avoid circular module dependencies.

Proposed layout:

```text
internal/
  app/
    bootstrap.go
    user_module.go
  shared/application/port/outbound/
    transactor.go
  platform/
    database/
    httpresponse/
    httpserver/
  user/
    domain/
    application/
      port/inbound/
      port/outbound/
    adapter/
      inbound/http/
        module.go
        user_handler.go
        user_dto.go
      outbound/postgres/
        user_repository.go
        user_model.go
```

### Application ports and compile-time checks

An inbound port exposes a cohesive operation. An outbound port describes what
that operation needs. Keep both contracts free of Echo, GORM and provider SDKs.
If command/result types would create an import cycle, define them in the inbound
contract package or a module-owned application types package.

```go
// user/application/port/inbound/create_user.go
package inbound

type CreateUserCommand struct {
    Email string
    Name  string
}

type UserCreator interface {
    Execute(context.Context, CreateUserCommand) (domain.User, error)
}
```

```go
// user/application/create_user.go
var _ inbound.UserCreator = (*CreateUser)(nil)

// user/adapter/outbound/postgres/user_repository.go
var _ outbound.UserRepository = (*Repository)(nil)
```

Declare an assertion beside every production implementation of an application
port, including application services and provider adapters. Test doubles can
use assertions when helpful but do not require production conventions.
Handlers receive inbound interfaces instead of `*application.CreateUser`.

### Composition and registration

An `internal/app/user_module.go` builder creates repositories and use cases,
then returns the module's HTTP registrar. Bootstrap creates shared resources,
invokes builders and registers each module once. The registrar groups routes
and handlers without containing business logic.

```go
// internal/app/user_module.go: illustrative wiring after port refactoring.
func buildUserModule(db *gorm.DB, validate *validator.Validate) *userhttp.Module {
    repo := userpostgres.NewRepository(db)
    create := userapplication.NewCreateUser(repo, time.Now)
    get := userapplication.NewGetUser(repo)
    return userhttp.NewModule(userhttp.Dependencies{
        Creator: create,
        Reader: get,
        Validator: validate,
    })
}
```

Use named dependency structs when several ports make positional constructors
ambiguous. Do not add a service locator or global dependency container.

## Transactions across repositories and modules

Application code controls the atomic boundary through a framework-free port:

```go
// shared/application/port/outbound/transactor.go
package outbound

type Transactor interface {
    WithinTransaction(context.Context, func(context.Context) error) error
}
```

A GORM implementation can keep the active handle in a private typed context key:

```go
// platform/database/transactor.go
package database

type transactionContextKey struct{}

type Transactor struct { db *gorm.DB }

func (t *Transactor) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
    if _, ok := ctx.Value(transactionContextKey{}).(*gorm.DB); ok {
        return work(ctx)
    }
    return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        return work(context.WithValue(ctx, transactionContextKey{}, tx))
    })
}

func Connection(ctx context.Context, fallback *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(transactionContextKey{}).(*gorm.DB); ok {
        return tx.WithContext(ctx)
    }
    return fallback.WithContext(ctx)
}
```

Every repository operation, including reads, resolves its connection at call
time. Retaining or directly using the base handle would escape the transaction:

```go
err := database.Connection(ctx, r.db).Create(toModel(user)).Error
```

A workflow passes the callback context to participating modules' public
application ports. Each module writes through its own repository. Return errors
through all nested callbacks so the outer boundary rolls back. Nested calls
reuse the transaction; they do not independently commit or provide savepoints.
This design assumes one shared database/pool: it is not a distributed transaction
or a multi-database resolver. Add explicit ownership checks if supporting multiple
pools later.

Do not wrap a single-write example in a transaction just to exercise the helper.
Use a meaningful multi-write example and verify that failure rolls back all writes.

## Persistence models and mapping

Keep domain entities, HTTP DTOs and persistence rows as distinct types. Domain
models must not contain GORM tags, `gorm.Model`, `gorm.DeletedAt`, SQL scanner
implementations, or database column metadata.

Each table has one canonical private persistence model, updated with its SQL
migration. Repositories use `Select`, `Omit`, or explicit updates for partial
operations instead of inventing competing row models. Join results may use
read-only `Row` or `Projection` structs without `TableName` methods.
Repositories should be named for their aggregate or query as a module grows.
A join returning an aggregate belongs to that aggregate's repository; a dashboard
projection belongs to a dedicated query adapter.

Mapping functions list fields explicitly. Do not use reflection copying that
can silently omit fields. Preserve database NULL separately from zero values.

| Value | Generic rule |
| --- | --- |
| UUID | Use a UUID/domain ID value; serialize a canonical string |
| Exact decimal/money | Use exact decimal values or validated strings; never binary floats for monetary arithmetic |
| Instant | Store/compare UTC; choose and document RFC 3339 precision at the API boundary |
| Timestamp without timezone | Interpret using an explicit field/business timezone |
| Calendar date | Validate `YYYY-MM-DD` as a calendar date; do not convert it as an instant |
| Nullable scalar | Pointer/nullable adapter value with explicit mapping |
| JSON | Prefer typed values or validated `json.RawMessage` for a known contract |
| Enum | Explicit conversions; reject unsupported values |
| Soft delete | ORM deletion metadata stays in persistence; recovery reads opt in explicitly |
| Large integer ID | Use a JSON string when consumers cannot represent the full integer exactly |

Generate IDs before persistence when multiple writes or provider calls need the
same identifier. A decimal string alone does not implement exact arithmetic;
choose a domain money/decimal implementation when calculations are needed.
Keep SQL migrations authoritative; never introduce `AutoMigrate`.

## HTTP DTOs, validation and response contracts

Move request/response structs and mapping functions into `user_dto.go`. Keep
route registration in the module registrar, and request parsing, validation,
application calls and response delivery in the handler. Application services
must not return HTTP status codes or response envelopes.

Validate transport shape at the HTTP boundary and business invariants inside
the owning domain/application. Bind only intended request fields; parse path and
query parameters deliberately rather than treating persistence models as input.
Document unknown-field policy, nullability, defaults, timestamps and enum handling.

Use typed shared response helpers rather than handwritten maps:

```go
type Envelope[T any] struct {
    Data T `json:"data"`
}

func Created[T any](c echo.Context, data T) error {
    return c.JSON(http.StatusCreated, Envelope[T]{Data: data})
}
```

Provide detail/action, collection and pagination helpers. Empty collections
serialize as `[]`. A possible generic pagination contract is
`{"data":{"items":[],"meta":{"page":1,"per_page":20,"total":0,"total_page":0}}}`.
Validate positive page/size and a maximum size before queries; document ordering
and overflow-safe page-count calculation. These exact field names are a template
choice and should be documented as part of the API contract.

Do not wrap `204`, redirects, binary responses or provider acknowledgments in
JSON. Helpers accept raw DTOs to prevent accidental double envelopes.

### Module-owned error translation

The current platform HTTP server imports `user/domain`. Remove that dependency:
the user HTTP adapter translates known domain errors; a shared HTTP error type
carries status, public code/message, details and an optional wrapped cause.

```go
// In the user HTTP adapter, after an application call.
if errors.Is(err, domain.ErrUserNotFound) {
    return httpresponse.WrapError(err, http.StatusNotFound,
        "USER_NOT_FOUND", "user not found", nil)
}
```

The platform handles shared HTTP errors and Echo errors, with unexpected errors
mapped to a safe `500 INTERNAL_ERROR`. Preserve `errors.Is` through `Unwrap`.
Always include nullable `details` and a request ID matching the response header.
Validation details identify fields/rules without echoing submitted values.

Define distinct defaults for 400, 401, 403, 404, 405, 409, 413, 422, 429 and 503;
unknown/internal errors must not expose their raw messages. In particular,
readiness failure must remain `503 SERVICE_UNAVAILABLE`.

## Observability and safe diagnostics

Propagate/generate request IDs. Name traces and metric labels using HTTP method
and registered route template to avoid identifiers or query values in labels.
Render errors before recording response status:

```go
err := next(c)
if err != nil {
    c.Error(err)
}
metrics.Observe(c.Request().Method, c.Path(), c.Response().Status, time.Since(start))
return nil // The error was already dispatched to the centralized handler.
```

Verify middleware order for handler errors, authentication failures, panics and
body-limit rejection; one helper change does not prove all paths are observed.
Keep HTTP counters/duration, structured logs and optional tracing. Set trace
error/status information deliberately rather than assuming it follows HTTP status.

Request logs contain request ID, method, route, final status and duration. Avoid
bodies, query values, authorization headers, passwords and raw provider/database
error messages. Safe server diagnostics can include public error code, Go error
types and a stable fingerprint. A fingerprint is for correlation, not a security
boundary; sensitive or guessable values must not be exposed through diagnostics.
Configure ORM logging to omit SQL parameter values as well.

Inject a Prometheus registry and use its matching gatherer for `/metrics` so tests
or multiple app instances do not depend on global collector registration.
Export tracing only when enabled; close exporters during process cleanup.

## Configuration, startup and shutdown

Honor `HTTP_MAX_BODY_BYTES` in Echo's body limiter instead of the current fixed
`1M`. Trim CORS origins and drop empty entries. Validate positive timeouts, body
limits, token TTL, valid ports and pool relationships at startup.

Add `DATABASE_CONNECT_TIMEOUT`; disable automatic GORM ping when explicitly
controlling its timeout. Close the pool when the bounded startup ping fails.
As resources are added, unwind already acquired resources on any later startup
failure. Cleanup must also run when startup exits early; `os.Exit` bypasses defers,
so a `run() error` function with cleanup followed by exit in `main` is preferable.

Keep liveness independent of PostgreSQL and readiness bounded with `503` on
failure. Keep the template's empty `204` liveness response.

On termination, drain HTTP requests before closing resources they use; then close
the pool/providers and flush tracing. Use a bounded shutdown context and retain
all failures:

```go
return errors.Join(httpShutdownErr, databaseCloseErr, providerCloseErr)
```

Define the forced-stop policy when draining exceeds its deadline. The normal
path must keep the pool open until in-flight requests finish.

Optional dotenv loading is a developer convenience. Load an optional local file
before reading environment config, preserve existing environment values, tolerate
only a missing file, and reject malformed files. Production must work without it.
Do not commit secrets. Environment-only loading remains a valid template default.

## Versioned OpenAPI and documentation

Maintain `api/openapi/v1/openapi.json` for implemented endpoints, embed it in the
binary, and expose a spec/docs route behind a documentation flag. Use capability
or module tags; express authentication using security schemes and operation
metadata, not role tags. Default documentation visibility deliberately by environment.

Validate specification syntax, route/method coverage, request/response schemas,
error responses and security requirements in CI. Keep endpoint examples in sync.
If Swagger UI is used, pin or self-host its assets instead of copying an unpinned
CDN dependency. Document only implemented routes.

## Testing conventions and acceptance evidence

Use deterministic clocks, fresh fixtures per case and fakes at application port
boundaries. Prefer typed table-driven cases when several scenarios share a runner.
Keep assertions in `<target>_test.go`; optionally move large fixture/case builders
and doubles to `<target>_dummy_test.go`, with shared infrastructure in
`init_test.go`. Do not split trivial tests solely to satisfy a naming rule.

HTTP tests should verify status, JSON contract, validation, safe errors and
identity handling. SQL-mock adapter tests verify query/error/mapping behavior;
they do not prove PostgreSQL constraints, locking, or transaction rollback.
Retain Testcontainers integration tests for those behaviors.

Required evidence when implementing these decisions:

- Application and adapter ports match implementations at compile time.
- Cross-module commands go through owner application ports.
- A multi-write failure rolls back; nested operations use the same transaction.
- HTTP errors have correct envelopes and final metric/log statuses.
- Sensitive values are absent from public errors and captured diagnostics.
- Body size follows config and CORS normalization behaves as documented.
- Startup failure releases acquired resources; shutdown drains active requests.
- OpenAPI matches implemented routes and authentication behavior.

Run relevant tests with the race detector, formatting, vet and builds. Add CI
import-boundary checks if conventions alone become insufficient. Keep Docker
integration checks available in a Docker-enabled environment.

## Migrations and external side effects

SQL migration execution stays separate from API startup. Document expand/contract
changes, backups, dirty-state recovery, forward fixes and realistic rollback
limits. Keep the API free of schema auto-migration. Adoption of an existing legacy
schema is project-specific and does not belong in the default starter.

For each external side effect, define whether it occurs before/after commit,
whether failure fails the operation, idempotency keys, retry policy and how to
reconcile timeouts with unknown outcomes. A database rollback cannot undo an
email, uploaded object, provider charge or broker publication.

Do not automatically hold network calls inside database transactions. Choose
ordering from the business guarantee. Reliable asynchronous publication requires
a transactional outbox and idempotent consumers; in-process events alone do not
provide durable delivery. Add these mechanisms when a real workflow needs them.

Reusable email/storage/push capabilities may expose framework-free shared ports
once multiple modules need them. Keep SDK clients in platform/outbound adapters,
while business templates and provider-specific behavior remain appropriately
owned. Avoid empty provider scaffolding and mandatory cloud credentials in the
base template.

## Implementation sequence

1. Extract HTTP DTOs and response/error helpers; fix final-status observability,
   configured body limits and CORS normalization.
2. Add directional ports, compile-time assertions, per-module app builders and
   a module HTTP registrar. Preserve existing endpoint behavior.
3. Add the transaction port/resolver with meaningful rollback coverage, then
   improve connection timeout, startup cleanup and shutdown error handling.
4. Add versioned embedded OpenAPI and contract verification; extend architecture,
   migration and test conventions as implemented capabilities grow.
5. Add reporting queries, external provider ports, outbox or richer serialization
   types only when an example/business requirement exercises them.

Project-specific token audiences, login compatibility, provider requirements,
business tables, migration-parity checks, and schema-adoption baselines remain
outside the template. Generic JWT verification and safe identity extraction
remain applicable; identity, role and session models follow each project's
requirements.
