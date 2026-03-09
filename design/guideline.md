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

- Favor rich domain models (entities/value objects/domain helpers) to enforce invariants and pure business rules instead of spreading conditionals across services.
- Prefer a rich domain model when the logic is stable, reusable, and does not depend on infrastructure. An anemic model also works for simple CRUD-style modules or transitional refactors, but do not let that become an excuse to leak business rules into delivery.
- Keep structs JSON/BSON/GORM tags in sync to simplify reuse across transports/persistence.
- When storing timestamps or enumerations, use strongly typed aliases or helper methods to avoid magic numbers in services.
- Offer helper APIs (e.g. `IsLoginInCtx`) for common cross-module queries.
- Good candidates for root-layer/domain helpers:
  - normalization rules (`NormalizeArtists`)
  - model-to-domain conversion (`MediaInfoCache.ToMediaInfo`)
  - deterministic selection logic (`PickMediaURL`, cache quality selection)
  - matching/resolution helpers that do not call repositories or external systems

## Responsibility Boundaries

- Put **pure rules** in the module root/domain layer:
  - no repository access
  - no HTTP/RPC context
  - no queue/task dispatch
  - deterministic input -> output logic
- Put **use case orchestration** in the service layer:
  - cache-first / fallback flows
  - cross-repository coordination
  - async task triggering
  - external provider calls
- Put **transport adaptation** in delivery:
  - request binding/validation
  - HTTP/RPC status and response formatting
  - middleware integration
- Delivery must not decide business workflows such as:
  - cache hit/miss policy
  - fallback order
  - enqueue/warmup timing
  - model reconstruction from persistence structs
- If a handler starts combining multiple service/cache calls to implement one business flow, that is usually a signal to move the orchestration into service and keep only pure conversion helpers in the root layer.

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
- Model-related invariants and pure selection/normalization rules should live on entities/value objects/domain helpers. Services focus on application workflows (e.g. validation, orchestration, cross-aggregate operations).
- Services may call other modules but must do so via interfaces registered in the `registry` to avoid tight coupling.
- When a service interface is exposed to other modules or RPC adapters, prefer methods that represent stable business capabilities instead of implementation details. For example, expose `ResolveMediaInfo(...)` rather than raw internal scheduling operations like `Enqueue...` unless background warmup is itself a business capability.

### Error Handling

- Service Layer can only return module-specific `errcode` errors.
- Service Layer **must** log every non-nil error returned by repository/external dependencies, with enough business context for troubleshooting (query params, entity IDs, operation name).
- Service Layer **must not** pass repository/driver/SQL/raw infrastructure error details to delivery/frontend (including `WithDetail(err)`-style passthrough).
- Service Layer may return only business errors, and at most attach an abstract/safe reason string that does not expose schema/table/column/SQL/path internals.
- Service Layer **must** propagate errors to delivery. Never swallow errors silently; if the use case recovers, log the discarded error explicitly.

### Additional rules

- Implement `Setup()` to finalize dependencies (e.g. prefixing loggers, validating configuration).
- If the application entrypoint has registered `logger.LoggerAddPrefix()`, injected loggers already receive the implementation name automatically. In that case, do not manually call `WithPrefix(...)` again in `Setup()` unless you intentionally want an additional sub-prefix. If the hook is not enabled, explicitly add a prefix in `Setup()` for every `scene.Service` / `scene.Named` implementation that owns logs.
- When services are context-aware, expose lightweight proxies (`WithSceneContext`) instead of letting delivery mutate the service directly.
- Prefer constructor functions for services/repositories instead of exporting structs directly. This keeps dependency wiring explicit and testable.
- A service may expose both:
  - cache-management methods returning raw cache records
  - resolve/use-case methods returning consumer-facing domain results
  as long as the naming clearly separates the two levels (`GetMediaInfoCache` vs `ResolveMediaInfo`).

## Delivery Layer Guidelines

- Delivery adapters translate transport details (HTTP routes, RPC methods, CLI commands) into service calls.
- Use scene-specific helpers (e.g. `sgin.AppRoutes`, `scenes/arpc` generated stubs) to keep handlers declarative.
- Validate inputs at the boundary using binding tags or explicit validators. Only pass clean, typed values into services.
- Map service errors to transport responses centrally (middleware or engine-level handler) so handlers can simply return the error.
- Delivery can set `scene.Context` values (user, locale, tracing) that downstream services access via helpers in the domain layer.
- Delivery should prefer one service call per business action. If a handler needs to:
  - check cache
  - fall back to upstream
  - enqueue background tasks
  - rebuild domain objects from cache structs
  then that logic belongs in service/root helpers rather than in the handler itself.

## Refactoring Heuristics

- When refactoring a fat handler:
  1. Move pure normalization/selection/conversion logic into the module root.
  2. Move cache-first/fallback/enqueue workflows into the service layer.
  3. Leave only binding and response shaping in delivery.
- Before adding a new service, ask whether the logic is:
  - a new reusable business capability
  - or just a few pure helpers plus existing service orchestration
- Do not create a new service only to host pure deterministic helpers. Keep those in the module root unless they need infrastructure access.

## Delivery Strategy

- Follow the sequence: **make it work, make it right, make it fast**.

### Make It Work

- First make the feature usable and safe enough to run:
  - it works end-to-end
  - it does not obviously crash
  - it does not create obviously broken state transitions
  - it still keeps the most basic layering
- At this stage, temporary compromises are acceptable, for example:
  - delivery contains too much orchestration logic
  - pure helper logic is duplicated in one or two places
  - domain behavior has not been fully extracted yet
- Even in this stage, the following are not acceptable:
  - silent data corruption
  - obvious security holes
  - swallowed important errors without logging
  - mixing transport concerns so deeply into lower layers that later cleanup becomes hard

### Make It Right

- Next align the code with this project’s DDD / Clean Architecture variant:
  - responsibility boundaries are explicit
  - delivery is thin
  - pure rules live in the module root/domain layer
  - orchestration lives in services
  - infrastructure details stay in repository/infrastructure layers
  - easy to maintain in the future
- This is the stage where temporary shortcuts from “make it work” should be paid back.
- If code works but the boundary is wrong, the work is still incomplete.

### Make It Fast

- Optimize only after correctness and responsibility boundaries are clear.
- Improve concrete bottlenecks such as:
  - repeated queries
  - unnecessary allocations or full-buffer reads
  - slow cache hit paths
  - expensive network/storage calls
- Do not trade away correctness or architectural clarity just to gain speed.
- A fast but boundary-breaking solution should usually be rejected or explicitly isolated and documented.

## Factories, DI, and cross-module access

- All dependencies are managed through `registry.Register`, `registry.Load`, and `aperture` tags. Avoid manual `new()` inside services unless you construct pure domain helpers.
- Factories own configuration lookup via `registry.Config`. Keep secrets or dynamic values here so services remain deterministic and testable.
- Cross-module usage must be defined via interfaces and registered implementations. Never import another module’s concrete types from `service/<impl>` or `repository/<impl>`—depend on the exported interfaces from the root package.

Following this guideline keeps new modules consistent with existing ones, passes code review faster, and ensures each scene application can freely assemble or replace infrastructure without touching the business logic.
