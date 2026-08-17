// Package registry implements module-scoped dependency declaration and
// injection.
//
// A Container owns one module's objects. Load declares an owned object,
// Register also makes it injectable inside that Container, and Export also
// makes it available to other Containers. Requires and Provides expose the
// resulting module boundary without constructing a second metadata model.
// Fields tagged with `aperture` contribute to Requires. Load, Register, and
// Export accept existing values, and their generic forms return the value
// unchanged so declaration and assignment can share one statement.
//
// Module factories create and connect their internal objects with ordinary Go
// code. External module dependencies are injected after every module has
// declared its objects, so field-based dependency cycles are supported.
//
// Scope owns global values, injection hooks, and one flat set of module
// Containers. Build validates and imports exports, orders the Containers, and
// performs the injection pass. Reset removes that module set while preserving
// global values and hooks. Create a Scope per test or independent runtime.
// Package-level Set, Use, and hook functions delegate to the default Scope.
// Registry deliberately has no lifecycle behavior; the root scene package
// owns module setup and teardown.
package registry
