// Package orm integrates GORM with Scene's data-source, logging, setup, and
// dependency-injection lifecycle.
//
// # Boundary
//
// Package orm is intentionally not a repository framework and does not define
// a database-independent query language. A module's repository interface is
// the persistence abstraction. Its GORM adapter uses this package to obtain a
// native *gorm.DB and is free to use GORM scopes, clauses, associations, or
// parameterized SQL when those best express the query.
//
// A repository injects the concrete component and creates a context-bound
// session for each operation:
//
//	type userRepository struct {
//		db *orm.Gorm `aperture:""`
//	}
//
//	func (r *userRepository) UserByID(
//		ctx context.Context,
//		id string,
//	) (userRow, error) {
//		var row userRow
//		err := r.db.Session(ctx).
//			Where(&userRow{UserID: id}).
//			First(&row).Error
//		return row, err
//	}
//
// # Transactions
//
// Transaction supplies one native transaction handle. Use that handle for
// every participating model so the transaction is not tied to a generic
// repository or a single row type:
//
//	err := db.Transaction(ctx, func(tx *gorm.DB) error {
//		if err := tx.Create(&user).Error; err != nil {
//			return err
//		}
//		return tx.Create(&token).Error
//	})
//
// # Schema setup
//
// AutoMigrate is a concrete GORM startup capability, not a framework port.
// GORM-backed repositories may call it from Setup:
//
//	func (r *userRepository) Setup() error {
//		return r.db.AutoMigrate(&userRow{})
//	}
//
// Persistence rows, GORM tags, table names, and domain conversion helpers
// belong in the concrete repository package. Domain models should not import
// GORM or carry GORM-only metadata.
//
// # Query guidance
//
// Prefer typed module repository methods and module-specific query values over
// exposing column names to services or delivery code. Simple equality queries
// can use struct or map conditions. Complex reads may use native GORM or
// parameterized SQL, but must remain inside the concrete repository adapter
// and return domain or read-model values.
package orm
