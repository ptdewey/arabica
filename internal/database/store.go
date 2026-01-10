package database

import "arabica/internal/models"

// Store defines the interface for all database operations
// This abstraction allows swapping SQLite for ATProto or other backends
type Store interface {
	// Brew operations
	// Note: userID parameter is deprecated for ATProto (user is implicit from DID)
	// It remains for SQLite compatibility but should not be relied upon
	CreateBrew(brew *models.CreateBrewRequest, userID int) (*models.Brew, error)
	GetBrewByRKey(rkey string) (*models.Brew, error)
	ListBrews(userID int) ([]*models.Brew, error)
	UpdateBrewByRKey(rkey string, brew *models.CreateBrewRequest) error
	DeleteBrewByRKey(rkey string) error

	// Bean operations
	CreateBean(bean *models.CreateBeanRequest) (*models.Bean, error)
	GetBeanByRKey(rkey string) (*models.Bean, error)
	ListBeans() ([]*models.Bean, error)
	UpdateBeanByRKey(rkey string, bean *models.UpdateBeanRequest) error
	DeleteBeanByRKey(rkey string) error

	// Roaster operations
	CreateRoaster(roaster *models.CreateRoasterRequest) (*models.Roaster, error)
	GetRoasterByRKey(rkey string) (*models.Roaster, error)
	ListRoasters() ([]*models.Roaster, error)
	UpdateRoasterByRKey(rkey string, roaster *models.UpdateRoasterRequest) error
	DeleteRoasterByRKey(rkey string) error

	// Grinder operations
	CreateGrinder(grinder *models.CreateGrinderRequest) (*models.Grinder, error)
	GetGrinderByRKey(rkey string) (*models.Grinder, error)
	ListGrinders() ([]*models.Grinder, error)
	UpdateGrinderByRKey(rkey string, grinder *models.UpdateGrinderRequest) error
	DeleteGrinderByRKey(rkey string) error

	// Brewer operations
	CreateBrewer(brewer *models.CreateBrewerRequest) (*models.Brewer, error)
	GetBrewerByRKey(rkey string) (*models.Brewer, error)
	ListBrewers() ([]*models.Brewer, error)
	UpdateBrewerByRKey(rkey string, brewer *models.UpdateBrewerRequest) error
	DeleteBrewerByRKey(rkey string) error

	// Close the database connection
	Close() error
}
