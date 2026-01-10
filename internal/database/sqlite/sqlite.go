package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

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
	// Create migrations table if it doesn't exist
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// List of migration files in order
	// Note: Add new migrations to the end of this list
	migrations := []string{
		"migrations/001_initial.sql",
		"migrations/002_add_brewers_table.sql",
		"migrations/003_update_grinders_schema.sql",
		"migrations/004_update_brews_add_grinder_brewer_ids.sql",
		"migrations/005_add_water_amount_to_brews.sql",
		"migrations/006_add_pours_table.sql",
		// Future migrations go here
	}

	for i, migrationPath := range migrations {
		version := i + 1

		// Check if migration already applied
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			// Migration already applied, skip
			continue
		}

		// Read and execute migration
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationPath, err)
		}

		if _, err := s.db.Exec(string(migration)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationPath, err)
		}

		// Record that migration was applied
		_, err = s.db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return fmt.Errorf("failed to record migration: %w", err)
		}

		fmt.Printf("Applied migration %d: %s\n", version, migrationPath)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Helper functions for RKey <-> ID conversion
func rkeyToID(rkey string) (int, error) {
	id, err := strconv.Atoi(rkey)
	if err != nil {
		return 0, fmt.Errorf("invalid rkey: %w", err)
	}
	return id, nil
}

func idToRKey(id int) string {
	return strconv.Itoa(id)
}

func optionalIDToRKey(id *int) string {
	if id == nil || *id == 0 {
		return ""
	}
	return strconv.Itoa(*id)
}

func rkeyToOptionalID(rkey string) *int {
	if rkey == "" {
		return nil
	}
	id, err := strconv.Atoi(rkey)
	if err != nil {
		return nil
	}
	return &id
}

// ========== Brew Operations ==========

func (s *SQLiteStore) CreateBrew(req *models.CreateBrewRequest, userID int) (*models.Brew, error) {
	beanID, err := rkeyToID(req.BeanRKey)
	if err != nil {
		return nil, fmt.Errorf("invalid bean_rkey: %w", err)
	}

	grinderID := rkeyToOptionalID(req.GrinderRKey)
	brewerID := rkeyToOptionalID(req.BrewerRKey)

	result, err := s.db.Exec(`
		INSERT INTO brews (user_id, bean_id, method, temperature, water_amount, time_seconds, 
			grind_size, grinder, grinder_id, brewer_id, tasting_notes, rating)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, beanID, req.Method, req.Temperature, req.WaterAmount, req.TimeSeconds,
		req.GrindSize, "", grinderID, brewerID, req.TastingNotes, req.Rating)

	if err != nil {
		return nil, fmt.Errorf("failed to create brew: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	// Create pours if provided
	if len(req.Pours) > 0 {
		if err := s.CreatePours(int(id), req.Pours); err != nil {
			return nil, fmt.Errorf("failed to create pours: %w", err)
		}
	}

	return s.getBrew(int(id))
}

func (s *SQLiteStore) GetBrewByRKey(rkey string) (*models.Brew, error) {
	id, err := rkeyToID(rkey)
	if err != nil {
		return nil, err
	}
	return s.getBrew(id)
}

func (s *SQLiteStore) getBrew(id int) (*models.Brew, error) {
	brew := &models.Brew{
		Bean: &models.Bean{
			Roaster: &models.Roaster{},
		},
		GrinderObj: &models.Grinder{},
		BrewerObj:  &models.Brewer{},
	}

	var brewID, beanID int
	var grinderID, brewerID sql.NullInt64
	var beanRoasterID sql.NullInt64
	var grinderObjID, brewerObjID, roasterObjID int

	err := s.db.QueryRow(`
		SELECT 
			b.id, b.bean_id, b.method, b.temperature, b.water_amount,
			b.time_seconds, b.grind_size, b.grinder_id, b.brewer_id, b.tasting_notes, b.rating, b.created_at,
			bn.id, bn.name, bn.origin, bn.roast_level, bn.process, bn.description, bn.roaster_id,
			COALESCE(r.id, 0), COALESCE(r.name, ''), COALESCE(r.location, ''), COALESCE(r.website, ''),
			COALESCE(g.id, 0), COALESCE(g.name, ''), COALESCE(g.grinder_type, ''), COALESCE(g.burr_type, ''), COALESCE(g.notes, ''),
			COALESCE(br.id, 0), COALESCE(br.name, ''), COALESCE(br.description, '')
		FROM brews b
		JOIN beans bn ON b.bean_id = bn.id
		LEFT JOIN roasters r ON bn.roaster_id = r.id
		LEFT JOIN grinders g ON b.grinder_id = g.id
		LEFT JOIN brewers br ON b.brewer_id = br.id
		WHERE b.id = ?
	`, id).Scan(
		&brewID, &beanID, &brew.Method, &brew.Temperature, &brew.WaterAmount,
		&brew.TimeSeconds, &brew.GrindSize, &grinderID, &brewerID, &brew.TastingNotes, &brew.Rating, &brew.CreatedAt,
		&beanID, &brew.Bean.Name, &brew.Bean.Origin, &brew.Bean.RoastLevel, &brew.Bean.Process, &brew.Bean.Description, &beanRoasterID,
		&roasterObjID, &brew.Bean.Roaster.Name, &brew.Bean.Roaster.Location, &brew.Bean.Roaster.Website,
		&grinderObjID, &brew.GrinderObj.Name, &brew.GrinderObj.GrinderType, &brew.GrinderObj.BurrType, &brew.GrinderObj.Notes,
		&brewerObjID, &brew.BrewerObj.Name, &brew.BrewerObj.Description,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get brew: %w", err)
	}

	// Convert IDs to RKeys
	brew.RKey = idToRKey(brewID)
	brew.BeanRKey = idToRKey(beanID)
	brew.Bean.RKey = idToRKey(beanID)

	if grinderID.Valid && grinderID.Int64 > 0 {
		brew.GrinderRKey = idToRKey(int(grinderID.Int64))
		brew.GrinderObj.RKey = brew.GrinderRKey
	}
	if brewerID.Valid && brewerID.Int64 > 0 {
		brew.BrewerRKey = idToRKey(int(brewerID.Int64))
		brew.BrewerObj.RKey = brew.BrewerRKey
	}
	if beanRoasterID.Valid && beanRoasterID.Int64 > 0 {
		brew.Bean.RoasterRKey = idToRKey(int(beanRoasterID.Int64))
		brew.Bean.Roaster.RKey = brew.Bean.RoasterRKey
	}

	// Load pours for this brew
	pours, err := s.ListPours(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pours: %w", err)
	}
	brew.Pours = pours

	return brew, nil
}

func (s *SQLiteStore) ListBrews(userID int) ([]*models.Brew, error) {
	rows, err := s.db.Query(`
		SELECT 
			b.id, b.bean_id, b.method, b.temperature, b.water_amount,
			b.time_seconds, b.grind_size, b.grinder_id, b.brewer_id, b.tasting_notes, b.rating, b.created_at,
			bn.id, bn.name, bn.origin, bn.roast_level, bn.process, bn.description, bn.roaster_id,
			COALESCE(r.id, 0), COALESCE(r.name, ''), COALESCE(r.location, ''), COALESCE(r.website, ''),
			COALESCE(g.id, 0), COALESCE(g.name, ''), COALESCE(g.grinder_type, ''), COALESCE(g.burr_type, ''), COALESCE(g.notes, ''),
			COALESCE(br.id, 0), COALESCE(br.name, ''), COALESCE(br.description, '')
		FROM brews b
		JOIN beans bn ON b.bean_id = bn.id
		LEFT JOIN roasters r ON bn.roaster_id = r.id
		LEFT JOIN grinders g ON b.grinder_id = g.id
		LEFT JOIN brewers br ON b.brewer_id = br.id
		WHERE b.user_id = ?
		ORDER BY b.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list brews: %w", err)
	}
	defer rows.Close()

	// First pass: collect brew data and IDs
	type brewWithID struct {
		brew   *models.Brew
		brewID int
	}
	var brewsWithIDs []brewWithID

	for rows.Next() {
		brew := &models.Brew{
			Bean: &models.Bean{
				Roaster: &models.Roaster{},
			},
			GrinderObj: &models.Grinder{},
			BrewerObj:  &models.Brewer{},
		}

		var brewID, beanID int
		var grinderID, brewerID sql.NullInt64
		var beanRoasterID sql.NullInt64
		var grinderObjID, brewerObjID, roasterObjID int

		err := rows.Scan(
			&brewID, &beanID, &brew.Method, &brew.Temperature, &brew.WaterAmount,
			&brew.TimeSeconds, &brew.GrindSize, &grinderID, &brewerID, &brew.TastingNotes, &brew.Rating, &brew.CreatedAt,
			&beanID, &brew.Bean.Name, &brew.Bean.Origin, &brew.Bean.RoastLevel, &brew.Bean.Process, &brew.Bean.Description, &beanRoasterID,
			&roasterObjID, &brew.Bean.Roaster.Name, &brew.Bean.Roaster.Location, &brew.Bean.Roaster.Website,
			&grinderObjID, &brew.GrinderObj.Name, &brew.GrinderObj.GrinderType, &brew.GrinderObj.BurrType, &brew.GrinderObj.Notes,
			&brewerObjID, &brew.BrewerObj.Name, &brew.BrewerObj.Description,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan brew: %w", err)
		}

		// Convert IDs to RKeys
		brew.RKey = idToRKey(brewID)
		brew.BeanRKey = idToRKey(beanID)
		brew.Bean.RKey = idToRKey(beanID)

		if grinderID.Valid && grinderID.Int64 > 0 {
			brew.GrinderRKey = idToRKey(int(grinderID.Int64))
			brew.GrinderObj.RKey = brew.GrinderRKey
		}
		if brewerID.Valid && brewerID.Int64 > 0 {
			brew.BrewerRKey = idToRKey(int(brewerID.Int64))
			brew.BrewerObj.RKey = brew.BrewerRKey
		}
		if beanRoasterID.Valid && beanRoasterID.Int64 > 0 {
			brew.Bean.RoasterRKey = idToRKey(int(beanRoasterID.Int64))
			brew.Bean.Roaster.RKey = brew.Bean.RoasterRKey
		}

		brewsWithIDs = append(brewsWithIDs, brewWithID{brew: brew, brewID: brewID})
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating brews: %w", err)
	}

	// Close rows before making additional queries
	rows.Close()

	// Second pass: load pours for each brew (now that rows is closed)
	brews := make([]*models.Brew, len(brewsWithIDs))
	for i, bwid := range brewsWithIDs {
		pours, err := s.ListPours(bwid.brewID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pours: %w", err)
		}
		bwid.brew.Pours = pours
		brews[i] = bwid.brew
	}

	return brews, nil
}

func (s *SQLiteStore) UpdateBrewByRKey(rkey string, req *models.CreateBrewRequest) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	beanID, err := rkeyToID(req.BeanRKey)
	if err != nil {
		return fmt.Errorf("invalid bean_rkey: %w", err)
	}

	grinderID := rkeyToOptionalID(req.GrinderRKey)
	brewerID := rkeyToOptionalID(req.BrewerRKey)

	_, err = s.db.Exec(`
		UPDATE brews 
		SET bean_id = ?, method = ?, temperature = ?, water_amount = ?, time_seconds = ?,
			grind_size = ?, grinder_id = ?, brewer_id = ?, tasting_notes = ?, rating = ?
		WHERE id = ?
	`, beanID, req.Method, req.Temperature, req.WaterAmount, req.TimeSeconds,
		req.GrindSize, grinderID, brewerID, req.TastingNotes, req.Rating, id)

	if err != nil {
		return fmt.Errorf("failed to update brew: %w", err)
	}

	// Update pours: delete existing and recreate
	if err := s.DeletePoursForBrew(id); err != nil {
		return fmt.Errorf("failed to delete existing pours: %w", err)
	}
	if len(req.Pours) > 0 {
		if err := s.CreatePours(id, req.Pours); err != nil {
			return fmt.Errorf("failed to create pours: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStore) DeleteBrewByRKey(rkey string) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	// Delete pours first (foreign key)
	if err := s.DeletePoursForBrew(id); err != nil {
		return fmt.Errorf("failed to delete pours: %w", err)
	}

	_, err = s.db.Exec("DELETE FROM brews WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete brew: %w", err)
	}
	return nil
}

// ========== Bean Operations ==========

func (s *SQLiteStore) CreateBean(req *models.CreateBeanRequest) (*models.Bean, error) {
	roasterID := rkeyToOptionalID(req.RoasterRKey)

	result, err := s.db.Exec(`
		INSERT INTO beans (name, origin, roast_level, process, description, roaster_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Name, req.Origin, req.RoastLevel, req.Process, req.Description, roasterID)

	if err != nil {
		return nil, fmt.Errorf("failed to create bean: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.getBean(int(id))
}

func (s *SQLiteStore) GetBeanByRKey(rkey string) (*models.Bean, error) {
	id, err := rkeyToID(rkey)
	if err != nil {
		return nil, err
	}
	return s.getBean(id)
}

func (s *SQLiteStore) getBean(id int) (*models.Bean, error) {
	bean := &models.Bean{
		Roaster: &models.Roaster{},
	}

	var beanID int
	var roasterID sql.NullInt64
	var roasterObjID int

	err := s.db.QueryRow(`
		SELECT b.id, b.name, b.origin, b.roast_level, b.process, b.description, b.roaster_id, b.created_at,
			COALESCE(r.id, 0), COALESCE(r.name, ''), COALESCE(r.location, ''), COALESCE(r.website, '')
		FROM beans b
		LEFT JOIN roasters r ON b.roaster_id = r.id
		WHERE b.id = ?
	`, id).Scan(&beanID, &bean.Name, &bean.Origin, &bean.RoastLevel, &bean.Process, &bean.Description, &roasterID, &bean.CreatedAt,
		&roasterObjID, &bean.Roaster.Name, &bean.Roaster.Location, &bean.Roaster.Website)

	if err != nil {
		return nil, fmt.Errorf("failed to get bean: %w", err)
	}

	bean.RKey = idToRKey(beanID)
	if roasterID.Valid && roasterID.Int64 > 0 {
		bean.RoasterRKey = idToRKey(int(roasterID.Int64))
		bean.Roaster.RKey = bean.RoasterRKey
	}

	return bean, nil
}

func (s *SQLiteStore) ListBeans() ([]*models.Bean, error) {
	rows, err := s.db.Query(`
		SELECT b.id, b.name, b.origin, b.roast_level, b.process, b.description, b.roaster_id, b.created_at,
			COALESCE(r.id, 0), COALESCE(r.name, ''), COALESCE(r.location, ''), COALESCE(r.website, '')
		FROM beans b
		LEFT JOIN roasters r ON b.roaster_id = r.id
		ORDER BY b.created_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list beans: %w", err)
	}
	defer rows.Close()

	var beans []*models.Bean
	for rows.Next() {
		bean := &models.Bean{
			Roaster: &models.Roaster{},
		}

		var beanID int
		var roasterID sql.NullInt64
		var roasterObjID int

		err := rows.Scan(&beanID, &bean.Name, &bean.Origin, &bean.RoastLevel, &bean.Process, &bean.Description, &roasterID, &bean.CreatedAt,
			&roasterObjID, &bean.Roaster.Name, &bean.Roaster.Location, &bean.Roaster.Website)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bean: %w", err)
		}

		bean.RKey = idToRKey(beanID)
		if roasterID.Valid && roasterID.Int64 > 0 {
			bean.RoasterRKey = idToRKey(int(roasterID.Int64))
			bean.Roaster.RKey = bean.RoasterRKey
		}

		beans = append(beans, bean)
	}

	return beans, nil
}

func (s *SQLiteStore) UpdateBeanByRKey(rkey string, req *models.UpdateBeanRequest) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	roasterID := rkeyToOptionalID(req.RoasterRKey)

	_, err = s.db.Exec(`
		UPDATE beans 
		SET name = ?, origin = ?, roast_level = ?, process = ?, description = ?, roaster_id = ?
		WHERE id = ?
	`, req.Name, req.Origin, req.RoastLevel, req.Process, req.Description, roasterID, id)

	if err != nil {
		return fmt.Errorf("failed to update bean: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteBeanByRKey(rkey string) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM beans WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bean: %w", err)
	}
	return nil
}

// ========== Roaster Operations ==========

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

	return s.getRoaster(int(id))
}

func (s *SQLiteStore) GetRoasterByRKey(rkey string) (*models.Roaster, error) {
	id, err := rkeyToID(rkey)
	if err != nil {
		return nil, err
	}
	return s.getRoaster(id)
}

func (s *SQLiteStore) getRoaster(id int) (*models.Roaster, error) {
	roaster := &models.Roaster{}
	var roasterID int

	err := s.db.QueryRow(`
		SELECT id, name, location, website, created_at
		FROM roasters WHERE id = ?
	`, id).Scan(&roasterID, &roaster.Name, &roaster.Location, &roaster.Website, &roaster.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get roaster: %w", err)
	}

	roaster.RKey = idToRKey(roasterID)
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
		var roasterID int

		err := rows.Scan(&roasterID, &roaster.Name, &roaster.Location, &roaster.Website, &roaster.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan roaster: %w", err)
		}

		roaster.RKey = idToRKey(roasterID)
		roasters = append(roasters, roaster)
	}

	return roasters, nil
}

func (s *SQLiteStore) UpdateRoasterByRKey(rkey string, req *models.UpdateRoasterRequest) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE roasters 
		SET name = ?, location = ?, website = ?
		WHERE id = ?
	`, req.Name, req.Location, req.Website, id)

	if err != nil {
		return fmt.Errorf("failed to update roaster: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteRoasterByRKey(rkey string) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM roasters WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete roaster: %w", err)
	}
	return nil
}

// ========== Grinder Operations ==========

func (s *SQLiteStore) CreateGrinder(req *models.CreateGrinderRequest) (*models.Grinder, error) {
	result, err := s.db.Exec(`
		INSERT INTO grinders (name, grinder_type, burr_type, notes) VALUES (?, ?, ?, ?)
	`, req.Name, req.GrinderType, req.BurrType, req.Notes)

	if err != nil {
		return nil, fmt.Errorf("failed to create grinder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.getGrinder(int(id))
}

func (s *SQLiteStore) GetGrinderByRKey(rkey string) (*models.Grinder, error) {
	id, err := rkeyToID(rkey)
	if err != nil {
		return nil, err
	}
	return s.getGrinder(id)
}

func (s *SQLiteStore) getGrinder(id int) (*models.Grinder, error) {
	grinder := &models.Grinder{}
	var grinderID int

	err := s.db.QueryRow(`
		SELECT id, name, grinder_type, burr_type, notes, created_at
		FROM grinders WHERE id = ?
	`, id).Scan(&grinderID, &grinder.Name, &grinder.GrinderType, &grinder.BurrType, &grinder.Notes, &grinder.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get grinder: %w", err)
	}

	grinder.RKey = idToRKey(grinderID)
	return grinder, nil
}

func (s *SQLiteStore) ListGrinders() ([]*models.Grinder, error) {
	rows, err := s.db.Query(`
		SELECT id, name, grinder_type, burr_type, notes, created_at
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
		var grinderID int

		err := rows.Scan(&grinderID, &grinder.Name, &grinder.GrinderType, &grinder.BurrType, &grinder.Notes, &grinder.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grinder: %w", err)
		}

		grinder.RKey = idToRKey(grinderID)
		grinders = append(grinders, grinder)
	}

	return grinders, nil
}

func (s *SQLiteStore) UpdateGrinderByRKey(rkey string, req *models.UpdateGrinderRequest) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE grinders 
		SET name = ?, grinder_type = ?, burr_type = ?, notes = ?
		WHERE id = ?
	`, req.Name, req.GrinderType, req.BurrType, req.Notes, id)

	if err != nil {
		return fmt.Errorf("failed to update grinder: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteGrinderByRKey(rkey string) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM grinders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete grinder: %w", err)
	}
	return nil
}

// ========== Brewer Operations ==========

func (s *SQLiteStore) CreateBrewer(req *models.CreateBrewerRequest) (*models.Brewer, error) {
	result, err := s.db.Exec(`
		INSERT INTO brewers (name, description) VALUES (?, ?)
	`, req.Name, req.Description)

	if err != nil {
		return nil, fmt.Errorf("failed to create brewer: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return s.getBrewer(int(id))
}

func (s *SQLiteStore) GetBrewerByRKey(rkey string) (*models.Brewer, error) {
	id, err := rkeyToID(rkey)
	if err != nil {
		return nil, err
	}
	return s.getBrewer(id)
}

func (s *SQLiteStore) getBrewer(id int) (*models.Brewer, error) {
	brewer := &models.Brewer{}
	var brewerID int

	err := s.db.QueryRow(`
		SELECT id, name, description, created_at
		FROM brewers WHERE id = ?
	`, id).Scan(&brewerID, &brewer.Name, &brewer.Description, &brewer.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to get brewer: %w", err)
	}

	brewer.RKey = idToRKey(brewerID)
	return brewer, nil
}

func (s *SQLiteStore) ListBrewers() ([]*models.Brewer, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at
		FROM brewers
		ORDER BY name ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to list brewers: %w", err)
	}
	defer rows.Close()

	var brewers []*models.Brewer
	for rows.Next() {
		brewer := &models.Brewer{}
		var brewerID int

		err := rows.Scan(&brewerID, &brewer.Name, &brewer.Description, &brewer.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brewer: %w", err)
		}

		brewer.RKey = idToRKey(brewerID)
		brewers = append(brewers, brewer)
	}

	return brewers, nil
}

func (s *SQLiteStore) UpdateBrewerByRKey(rkey string, req *models.UpdateBrewerRequest) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		UPDATE brewers 
		SET name = ?, description = ?
		WHERE id = ?
	`, req.Name, req.Description, id)

	if err != nil {
		return fmt.Errorf("failed to update brewer: %w", err)
	}

	return nil
}

func (s *SQLiteStore) DeleteBrewerByRKey(rkey string) error {
	id, err := rkeyToID(rkey)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("DELETE FROM brewers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete brewer: %w", err)
	}
	return nil
}

// ========== Pour Operations ==========

func (s *SQLiteStore) CreatePours(brewID int, pours []models.CreatePourData) error {
	if len(pours) == 0 {
		return nil
	}

	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO pours (brew_id, pour_number, water_amount, time_seconds)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, pour := range pours {
		_, err := stmt.Exec(brewID, i+1, pour.WaterAmount, pour.TimeSeconds)
		if err != nil {
			return fmt.Errorf("failed to insert pour: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *SQLiteStore) ListPours(brewID int) ([]*models.Pour, error) {
	rows, err := s.db.Query(`
		SELECT pour_number, water_amount, time_seconds, created_at
		FROM pours
		WHERE brew_id = ?
		ORDER BY pour_number ASC
	`, brewID)

	if err != nil {
		return nil, fmt.Errorf("failed to list pours: %w", err)
	}
	defer rows.Close()

	var pours []*models.Pour
	for rows.Next() {
		pour := &models.Pour{}
		err := rows.Scan(&pour.PourNumber, &pour.WaterAmount, &pour.TimeSeconds, &pour.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pour: %w", err)
		}
		pours = append(pours, pour)
	}

	return pours, nil
}

func (s *SQLiteStore) DeletePoursForBrew(brewID int) error {
	_, err := s.db.Exec("DELETE FROM pours WHERE brew_id = ?", brewID)
	if err != nil {
		return fmt.Errorf("failed to delete pours: %w", err)
	}
	return nil
}
