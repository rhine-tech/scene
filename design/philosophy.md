# scene framework design philosophy

The **scene** framework is built around two complementary ideas: model the business as self-contained modules (lenses) and keep every dependency pointing inward as mandated by Clean Architecture. This section captures the principles that guide every package in the repository.

## Module-first domain driven design

- Each domain is a **lens** (`scene.ModuleName`) with its own entities, errors, and service/repository interfaces under `lens/<module>`.
- Modules own their ubiquitous language. Delivery layers never speak directly to persistence models; they interact with services that return domain objects.
- Cross-module collaboration happens through explicitly exported interfaces and is wired via the registry, preventing accidental coupling.

## Clean architecture layering

- The dependency rule is strict: `delivery -> services -> domain -> repository interfaces`. Infrastructure adapters implement those interfaces in sibling packages (e.g. `repository/gorm`).
- Domain code carries behavior (methods on entities/value objects) so that application services remain thin orchestrators.
- Delivery adapters (gin, arpc, websocket, etc.) reside under `scenes/` and only access modules through the transport-agnostic contracts provided at the lens root.

## Scenes, applications, and factories

- A **scene** bundles a delivery engine with a set of module applications. The same module can expose multiple apps (HTTP, RPC) without duplicating business logic.
- `scene.ModuleFactory` instances declare how a module should be wired (repositories, services, apps). `scene.BuildInitArray` and `scene.BuildApps` compose factories so a scene can be assembled declaratively.
- Factories double as documentation for required configuration and provide `Default()` values to seed sensible demos.

## Dependency management and DI

- Dependencies are resolved through the lightweight `registry` package. Components register themselves and declare requirements via struct fields tagged with `` `aperture:""` ``.
- Because registrations happen in code (usually inside factories), there is no hidden magic: tracing who provides an implementation is as simple as searching for `registry.Register`.
- Configuration travels through `registry.Config`, keeping services deterministic and making it easy to supply overrides during tests.

## Error handling and observability

- Modules define their own `errcode` groups so every exported error has a stable code and message. Service layers only return these errors, ensuring consistent responses across deliveries.
- Logging uses the pluggable `logger.ILogger` interface. Implementations attach module/implementation identifiers (via `ImplName`) so logs are correlated with DI bindings.
- Context propagation relies on `scene.Context`, allowing delivery adapters to enrich the request (authentication, tracing) without leaking transport dependencies into services.

## Infrastructure composition

- Shared infrastructure helpers live under `composition/` (ORM thin wrappers, etc.) while reusable drivers sit under `infrastructure/` (logger, datasource, cache).
- Repository implementations combine these helpers with the domain contracts, keeping persistence concerns isolated and replaceable.
- Code generation tools under `cmd/scene` automate repetitive plumbing (e.g. ARPC stubs), but the generated code still respects the same layering rules.

## Evolvability and testability

- Modules can be tested in isolation by substituting repository/service interfaces. There are no global singletons other than the controlled registry, which tests can reset.
- Observing the dependency rule allows multiple teams to work on different lenses without merge conflicts—the only shared boundary is the exported contracts.
- Adding a new delivery technology (a new scene) requires only writing an adapter and wiring it through factories; the domain and services stay untouched.

These principles ensure the framework remains modular, predictable, and aligned with domain-driven design even as new infrastructure or delivery options are introduced.
