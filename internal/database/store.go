package database

import "arabica/internal/models"

// Store defines the interface for all database operations
// This abstraction allows swapping SQLite for PostgreSQL or other databases later
type Store interface {
	// Brew operations
	CreateBrew(brew *models.CreateBrewRequest, userID int) (*models.Brew, error)
	GetBrew(id int) (*models.Brew, error)
	ListBrews(userID int) ([]*models.Brew, error)
	UpdateBrew(id int, brew *models.CreateBrewRequest) error
	DeleteBrew(id int) error

	// Bean operations
	CreateBean(bean *models.CreateBeanRequest) (*models.Bean, error)
	GetBean(id int) (*models.Bean, error)
	ListBeans() ([]*models.Bean, error)
	UpdateBean(id int, bean *models.UpdateBeanRequest) error
	DeleteBean(id int) error

	// Roaster operations
	CreateRoaster(roaster *models.CreateRoasterRequest) (*models.Roaster, error)
	GetRoaster(id int) (*models.Roaster, error)
	ListRoasters() ([]*models.Roaster, error)
	UpdateRoaster(id int, roaster *models.UpdateRoasterRequest) error
	DeleteRoaster(id int) error

	// Grinder operations
	CreateGrinder(grinder *models.CreateGrinderRequest) (*models.Grinder, error)
	GetGrinder(id int) (*models.Grinder, error)
	ListGrinders() ([]*models.Grinder, error)
	UpdateGrinder(id int, grinder *models.UpdateGrinderRequest) error
	DeleteGrinder(id int) error

	// Close the database connection
	Close() error
}
