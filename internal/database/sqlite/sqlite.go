package sqlite

import (
	"database/sql"
	"fmt"
	"os"

	"arabica/internal/models"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store and runs migrations
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Run migrations
	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) runMigrations() error {
	migration, err := os.ReadFile("migrations/001_initial.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := s.db.Exec(string(migration)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Brew operations

func (s *SQLiteStore) CreateBrew(req *models.CreateBrewRequest, userID int) (*models.Brew, error) {
	result, err := s.db.Exec(`
		INSERT INTO brews (user_id, bean_id, roaster_id, method, temperature, time_seconds, 
			grind_size, grinder, tasting_notes, rating)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, req.BeanID, req.RoasterID, req.Method, req.Temperature, req.TimeSeconds,
		req.GrindSize, req.Grinder, req.TastingNotes, req.Rating)

	if err != nil {
		return nil, fmt.Errorf("failed to create brew: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetBrew(int(id))
}

func (s *SQLiteStore) GetBrew(id int) (*models.Brew, error) {
	brew := &models.Brew{
		Bean:    &models.Bean{},
		Roaster: &models.Roaster{},
	}

	err := s.db.QueryRow(`
		SELECT 
			b.id, b.user_id, b.bean_id, b.roaster_id, b.method, b.temperature, 
			b.time_seconds, b.grind_size, b.grinder, b.tasting_notes, b.rating, b.created_at,
			bn.id, bn.name, bn.origin, bn.roast_level, bn.description,
			r.id, r.name
		FROM brews b
		JOIN beans bn ON b.bean_id = bn.id
		JOIN roasters r ON b.roaster_id = r.id
		WHERE b.id = ?
	`, id).Scan(
		&brew.ID, &brew.UserID, &brew.BeanID, &brew.RoasterID, &brew.Method, &brew.Temperature,
		&brew.TimeSeconds, &brew.GrindSize, &brew.Grinder, &brew.TastingNotes, &brew.Rating, &brew.CreatedAt,
		&brew.Bean.ID, &brew.Bean.Name, &brew.Bean.Origin, &brew.Bean.RoastLevel, &brew.Bean.Description,
		&brew.Roaster.ID, &brew.Roaster.Name,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get brew: %w", err)
	}

	return brew, nil
}

func (s *SQLiteStore) ListBrews(userID int) ([]*models.Brew, error) {
	rows, err := s.db.Query(`
		SELECT 
			b.id, b.user_id, b.bean_id, b.roaster_id, b.method, b.temperature, 
			b.time_seconds, b.grind_size, b.grinder, b.tasting_notes, b.rating, b.created_at,
			bn.id, bn.name, bn.origin, bn.roast_level, bn.description,
			r.id, r.name
		FROM brews b
		JOIN beans bn ON b.bean_id = bn.id
		JOIN roasters r ON b.roaster_id = r.id
		WHERE b.user_id = ?
		ORDER BY b.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list brews: %w", err)
	}
	defer rows.Close()

	var brews []*models.Brew
	for rows.Next() {
		brew := &models.Brew{
			Bean:    &models.Bean{},
			Roaster: &models.Roaster{},
		}

		err := rows.Scan(
			&brew.ID, &brew.UserID, &brew.BeanID, &brew.RoasterID, &brew.Method, &brew.Temperature,
			&brew.TimeSeconds, &brew.GrindSize, &brew.Grinder, &brew.TastingNotes, &brew.Rating, &brew.CreatedAt,
			&brew.Bean.ID, &brew.Bean.Name, &brew.Bean.Origin, &brew.Bean.RoastLevel, &brew.Bean.Description,
			&brew.Roaster.ID, &brew.Roaster.Name,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan brew: %w", err)
		}

		brews = append(brews, brew)
	}

	return brews, nil
}

func (s *SQLiteStore) UpdateBrew(id int, req *models.CreateBrewRequest) error {
	_, err := s.db.Exec(`
		UPDATE brews 
		SET bean_id = ?, roaster_id = ?, method = ?, temperature = ?, time_seconds = ?,
			grind_size = ?, grinder = ?, tasting_notes = ?, rating = ?
		WHERE id = ?
	`, req.BeanID, req.RoasterID, req.Method, req.Temperature, req.TimeSeconds,
		req.GrindSize, req.Grinder, req.TastingNotes, req.Rating, id)

	if err != nil {
		return fmt.Errorf("failed to update brew: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteBrew(id int) error {
	_, err := s.db.Exec("DELETE FROM brews WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete brew: %w", err)
	}
	return nil
}

// Bean operations

func (s *SQLiteStore) CreateBean(req *models.CreateBeanRequest) (*models.Bean, error) {
	result, err := s.db.Exec(`
		INSERT INTO beans (name, origin, roast_level, description)
		VALUES (?, ?, ?, ?)
	`, req.Name, req.Origin, req.RoastLevel, req.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to create bean: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetBean(int(id))
}

func (s *SQLiteStore) GetBean(id int) (*models.Bean, error) {
	bean := &models.Bean{}
	err := s.db.QueryRow(`
		SELECT id, name, origin, roast_level, description, created_at
		FROM beans WHERE id = ?
	`, id).Scan(&bean.ID, &bean.Name, &bean.Origin, &bean.RoastLevel, &bean.Description, &bean.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get bean: %w", err)
	}

	return bean, nil
}

func (s *SQLiteStore) ListBeans() ([]*models.Bean, error) {
	rows, err := s.db.Query(`
		SELECT id, name, origin, roast_level, description, created_at
		FROM beans
		ORDER BY created_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list beans: %w", err)
	}
	defer rows.Close()

	var beans []*models.Bean
	for rows.Next() {
		bean := &models.Bean{}
		err := rows.Scan(&bean.ID, &bean.Name, &bean.Origin, &bean.RoastLevel, &bean.Description, &bean.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bean: %w", err)
		}
		beans = append(beans, bean)
	}

	return beans, nil
}

// Roaster operations

func (s *SQLiteStore) CreateRoaster(req *models.CreateRoasterRequest) (*models.Roaster, error) {
	result, err := s.db.Exec(`
		INSERT INTO roasters (name, location, website) VALUES (?, ?, ?)
	`, req.Name, req.Location, req.Website)

	if err != nil {
		return nil, fmt.Errorf("failed to create roaster: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetRoaster(int(id))
}

func (s *SQLiteStore) GetRoaster(id int) (*models.Roaster, error) {
	roaster := &models.Roaster{}
	err := s.db.QueryRow(`
		SELECT id, name, location, website, created_at
		FROM roasters WHERE id = ?
	`, id).Scan(&roaster.ID, &roaster.Name, &roaster.Location, &roaster.Website, &roaster.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get roaster: %w", err)
	}

	return roaster, nil
}

func (s *SQLiteStore) ListRoasters() ([]*models.Roaster, error) {
	rows, err := s.db.Query(`
		SELECT id, name, location, website, created_at
		FROM roasters
		ORDER BY name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list roasters: %w", err)
	}
	defer rows.Close()

	var roasters []*models.Roaster
	for rows.Next() {
		roaster := &models.Roaster{}
		err := rows.Scan(&roaster.ID, &roaster.Name, &roaster.Location, &roaster.Website, &roaster.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan roaster: %w", err)
		}
		roasters = append(roasters, roaster)
	}

	return roasters, nil
}

func (s *SQLiteStore) UpdateRoaster(id int, req *models.UpdateRoasterRequest) error {
	_, err := s.db.Exec(`
		UPDATE roasters 
		SET name = ?, location = ?, website = ?
		WHERE id = ?
	`, req.Name, req.Location, req.Website, id)

	if err != nil {
		return fmt.Errorf("failed to update roaster: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteRoaster(id int) error {
	_, err := s.db.Exec("DELETE FROM roasters WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete roaster: %w", err)
	}
	return nil
}

// Bean update/delete operations

func (s *SQLiteStore) UpdateBean(id int, req *models.UpdateBeanRequest) error {
	_, err := s.db.Exec(`
		UPDATE beans 
		SET name = ?, origin = ?, roast_level = ?, description = ?
		WHERE id = ?
	`, req.Name, req.Origin, req.RoastLevel, req.Description, id)

	if err != nil {
		return fmt.Errorf("failed to update bean: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteBean(id int) error {
	_, err := s.db.Exec("DELETE FROM beans WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bean: %w", err)
	}
	return nil
}

// Grinder operations

func (s *SQLiteStore) CreateGrinder(req *models.CreateGrinderRequest) (*models.Grinder, error) {
	result, err := s.db.Exec(`
		INSERT INTO grinders (name, type, notes) VALUES (?, ?, ?)
	`, req.Name, req.Type, req.Notes)

	if err != nil {
		return nil, fmt.Errorf("failed to create grinder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.GetGrinder(int(id))
}

func (s *SQLiteStore) GetGrinder(id int) (*models.Grinder, error) {
	grinder := &models.Grinder{}
	err := s.db.QueryRow(`
		SELECT id, name, type, notes, created_at
		FROM grinders WHERE id = ?
	`, id).Scan(&grinder.ID, &grinder.Name, &grinder.Type, &grinder.Notes, &grinder.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get grinder: %w", err)
	}

	return grinder, nil
}

func (s *SQLiteStore) ListGrinders() ([]*models.Grinder, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, notes, created_at
		FROM grinders
		ORDER BY name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list grinders: %w", err)
	}
	defer rows.Close()

	var grinders []*models.Grinder
	for rows.Next() {
		grinder := &models.Grinder{}
		err := rows.Scan(&grinder.ID, &grinder.Name, &grinder.Type, &grinder.Notes, &grinder.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grinder: %w", err)
		}
		grinders = append(grinders, grinder)
	}

	return grinders, nil
}

func (s *SQLiteStore) UpdateGrinder(id int, req *models.UpdateGrinderRequest) error {
	_, err := s.db.Exec(`
		UPDATE grinders 
		SET name = ?, type = ?, notes = ?
		WHERE id = ?
	`, req.Name, req.Type, req.Notes, id)

	if err != nil {
		return fmt.Errorf("failed to update grinder: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteGrinder(id int) error {
	_, err := s.db.Exec("DELETE FROM grinders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete grinder: %w", err)
	}
	return nil
}
