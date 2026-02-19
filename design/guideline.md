# scene framework usage guideline

This document explains how to build a scene **module** (a.k.a. lens) following the DDD and Clean Architecture conventions used across the repository. The goals are:

- keep domain logic isolated and testable
- make infrastructure replaceable
- expose delivery adapters without leaking transport concerns into the core

Most of the examples below reference files from `lens/authentication` because it showcases every layer.

## Module anatomy

Each module lives under `lens/<module-name>` and is structured by layers. A typical shape is:

```
lens/<module>/
  model.go            // domain entities/value objects
  error.go            // errcode definitions
  service.go          // public service interfaces
  repository.go       // repository (outbound port) interfaces
  context.go|api.go   // shared helpers/context keys
  delivery/           // per-scene applications (gin/arpc/etc.)
  repository/         // infrastructure implementations (gorm, mongo…)
  service/            // concrete use case services
  factory/            // wiring for repositories/services/apps
  gen/                // generated code (e.g. arpc stubs)
```

A module **must** declare a lens identifier:

```go
const Lens scene.ModuleName = "authentication"
```

`scene.ModuleName` provides helpers such as `Lens.TableName("users")` and `Lens.ImplName("Interface", "impl")`, ensuring unique names across the system.

## Module creation recipe

1. **Model the domain**
   - Define entities, aggregates, value objects, and domain methods in `model.go`. Keep invariants close to the data.
   - Introduce rich helper types (contexts, APIs) when multiple layers need to share state, e.g. `AuthContext` in `context.go`.
   - Declare domain errors with `errcode.NewErrorGroup` so they can travel across layers predictably.

2. **Describe the ports**
   - Define service interfaces in `<module>/service.go`. They extend `scene.Service` (providing `SrvImplName()`/`Setup()` hooks) and should not expose transport-specific details.
   - Define repository interfaces (outbound ports) in `<module>/repository.go`. They extend `scene.Named` and only return domain types.

3. **Implement repositories (infrastructure layer)**
   - Add infrastructure-specific packages under `repository/` (gorm, mongo, redis…). Each struct embeds/uses shared components such as `composition/orm` repositories.
   - Implement `ImplName()` to describe the adapter, e.g. `AuthenticationRepository.gorm`.
   - Repository methods interact with external systems and can return low-level errors. Translate them to domain errors where it clarifies intent, otherwise bubble up and let the service decide.
   - Keep transport/logging minimal—repositories may log for troubleshooting but must never swallow errors.

4. **Implement services (application layer)**
   - Place concrete services inside `service/`. Wire dependencies through struct fields tagged with ``` `aperture:""` ``` to let the DI container inject them.
   - Services orchestrate repositories and domain objects. Push domain invariants back to entities/value objects to keep services thin.
   - Convert every error to your module’s `errcode` before returning to delivery. Wrap external errors with helpers such as `ErrInternalError.WrapIfNot(err)` so callers always receive typed errors.
   - Use `logger.ILogger` and override its prefix in `Setup()` using `SrvImplName().Identifier()` to make logs searchable.
   - If a service needs request context (scene.Context), expose a `WithSceneContext` helper similar to `token.CtxProxy`.

5. **Deliver the use cases**
   - Under `delivery/`, add adapters per scene (`delivery/gin`, `delivery/arpc`, `delivery/middleware`, etc.). Each adapter composes actions or handlers that call services.
   - Keep DTOs/VOs (view objects) local to delivery packages. Transform domain entities into response objects explicitly (e.g. `UserNoPassword`).
   - Delivery code is the only layer aware of HTTP/RPC concepts such as status codes, headers, or gin contexts.
   - Use module-level middleware to inject contexts (e.g. `authentication.SetAuthContext`).

6. **Wire everything via factories**
   - Factories live under `factory/` and embed `scene.ModuleFactory`.
   - `Init()` registers repositories and services inside the global `registry`. Prefer constructor functions (`service.NewAuthenticationService`) so you can call `registry.Load(...)` to resolve dependencies declared with `aperture` tags.
   - `Apps()` returns delivery initializers per scene. These functions are executed by scene engines (`scenes/gin`, `scenes/arpc`) when building the application container.
   - Provide `Default()` values for factories that require configuration (e.g. secrets, header names).
   - Compose multiple factories with `scene.BuildInitArray` and `scene.BuildApps` inside your scene entrypoint.

7. **Test and validate**
   - Unit test domain logic without touching the registry. Mock repositories by implementing the interface.
   - Integration tests can spin up real infrastructure adapters; ensure they stay under `*_test.go` next to the implementation.
   - Validate configuration/registration in `factory` tests to prevent runtime DI errors.

## Domain Layer Guidelines

- Favor value objects/methods on entities to enforce invariants instead of spreading conditionals across services.
- Keep structs JSON/BSON/GORM tags in sync to simplify reuse across transports/persistence.
- When storing timestamps or enumerations, use strongly typed aliases or helper methods to avoid magic numbers in services.
- Offer helper APIs (e.g. `IsLoginInCtx`) for common cross-module queries.

## Repository Layer Design Principle

### Responsibilities

- Repository Layer is the boundary to external systems (databases, caches, message brokers, third-party APIs).
- It maps persistence schemas to domain entities and is allowed to use infrastructure helpers such as `composition/orm`.
- It must not leak DB-specific types to services—convert everything to domain structs or pagination models from `model`.

### Error Handling

- Repository Layer can return any error (driver, network, etc.). Prefer wrapping them with module errors when it clarifies semantics (`ErrUserNotFound`, `ErrAuthenticationFailed`).
- Errors can be logged in the repository layer for diagnostics but **must** be propagated to the service layer.
- Do not translate infrastructure errors into HTTP codes—leave transport decisions to delivery.

### Additional rules

- Keep transactions and connection lifecycle inside the repository. Accept dependencies (ORM handles, clients) through constructors.
- Expose pagination via `model.PaginationResult` so the service/delivery layers can stream results consistently.
- Use `scene.Named`’s `ImplName()` so observability/logging exposes which adapter is used.

## Service Layer Design Principle

### Responsibilities

- Service Layer orchestrates business use cases. It coordinates repositories, domain logic, and other services through interfaces.
- Model-related invariants should live on the entities/value objects. Services focus on application workflows (e.g. validation, orchestration, cross-aggregate operations).
- Services may call other modules but must do so via interfaces registered in the `registry` to avoid tight coupling.

### Error Handling

- Service Layer can only return module-specific `errcode` errors.
- Service Layer **must** log every non-nil error returned by repository/external dependencies, with enough business context for troubleshooting (query params, entity IDs, operation name).
- Service Layer **must not** pass repository/driver/SQL/raw infrastructure error details to delivery/frontend (including `WithDetail(err)`-style passthrough).
- Service Layer may return only business errors, and at most attach an abstract/safe reason string that does not expose schema/table/column/SQL/path internals.
- Service Layer **must** propagate errors to delivery. Never swallow errors silently; if the use case recovers, log the discarded error explicitly.

### Additional rules

- Implement `Setup()` to finalize dependencies (e.g. prefixing loggers, validating configuration).
- When services are context-aware, expose lightweight proxies (`WithSceneContext`) instead of letting delivery mutate the service directly.
- Prefer constructor functions for services/repositories instead of exporting structs directly. This keeps dependency wiring explicit and testable.

## Delivery Layer Guidelines

- Delivery adapters translate transport details (HTTP routes, RPC methods, CLI commands) into service calls.
- Use scene-specific helpers (e.g. `sgin.AppRoutes`, `scenes/arpc` generated stubs) to keep handlers declarative.
- Validate inputs at the boundary using binding tags or explicit validators. Only pass clean, typed values into services.
- Map service errors to transport responses centrally (middleware or engine-level handler) so handlers can simply return the error.
- Delivery can set `scene.Context` values (user, locale, tracing) that downstream services access via helpers in the domain layer.

## Factories, DI, and cross-module access

- All dependencies are managed through `registry.Register`, `registry.Load`, and `aperture` tags. Avoid manual `new()` inside services unless you construct pure domain helpers.
- Factories own configuration lookup via `registry.Config`. Keep secrets or dynamic values here so services remain deterministic and testable.
- Cross-module usage must be defined via interfaces and registered implementations. Never import another module’s concrete types from `service/<impl>` or `repository/<impl>`—depend on the exported interfaces from the root package.

Following this guideline keeps new modules consistent with existing ones, passes code review faster, and ensures each scene application can freely assemble or replace infrastructure without touching the business logic.
