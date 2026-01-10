package sqlite

import (
	"os"
	"path/filepath"
	"testing"

	"arabica/internal/models"
)

// TestMain changes to the project root directory so migrations can be found
func TestMain(m *testing.M) {
	// Find and change to the project root (where migrations/ dir exists)
	// This is needed because migrations are loaded relative to cwd
	wd, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(wd, "migrations")); err == nil {
			os.Chdir(wd)
			break
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			// Reached root, migrations not found
			break
		}
		wd = parent
	}
	os.Exit(m.Run())
}

// newTestStore creates an in-memory SQLite store for testing
func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() {
		store.Close()
	})
	return store
}

// ========== Helper Function Tests ==========

func TestRkeyToID(t *testing.T) {
	tests := []struct {
		name    string
		rkey    string
		want    int
		wantErr bool
	}{
		{"valid positive", "123", 123, false},
		{"valid zero", "0", 0, false},
		{"valid large", "999999", 999999, false},
		{"invalid empty", "", 0, true},
		{"invalid letters", "abc", 0, true},
		{"invalid mixed", "12a3", 0, true},
		{"invalid float", "12.3", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rkeyToID(tt.rkey)
			if (err != nil) != tt.wantErr {
				t.Errorf("rkeyToID(%q) error = %v, wantErr %v", tt.rkey, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("rkeyToID(%q) = %v, want %v", tt.rkey, got, tt.want)
			}
		})
	}
}

func TestIdToRKey(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want string
	}{
		{"positive", 123, "123"},
		{"zero", 0, "0"},
		{"large", 999999, "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := idToRKey(tt.id); got != tt.want {
				t.Errorf("idToRKey(%d) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestOptionalIDToRKey(t *testing.T) {
	intPtr := func(i int) *int { return &i }

	tests := []struct {
		name string
		id   *int
		want string
	}{
		{"nil", nil, ""},
		{"zero", intPtr(0), ""},
		{"positive", intPtr(123), "123"},
		{"large", intPtr(999999), "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := optionalIDToRKey(tt.id); got != tt.want {
				t.Errorf("optionalIDToRKey(%v) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestRkeyToOptionalID(t *testing.T) {
	tests := []struct {
		name    string
		rkey    string
		wantNil bool
		wantVal int
	}{
		{"empty", "", true, 0},
		{"invalid", "abc", true, 0},
		{"valid", "123", false, 123},
		{"zero", "0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rkeyToOptionalID(tt.rkey)
			if tt.wantNil {
				if got != nil {
					t.Errorf("rkeyToOptionalID(%q) = %v, want nil", tt.rkey, *got)
				}
			} else {
				if got == nil {
					t.Errorf("rkeyToOptionalID(%q) = nil, want %d", tt.rkey, tt.wantVal)
				} else if *got != tt.wantVal {
					t.Errorf("rkeyToOptionalID(%q) = %d, want %d", tt.rkey, *got, tt.wantVal)
				}
			}
		})
	}
}

// ========== Roaster CRUD Tests ==========

func TestRoasterCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create
	req := &models.CreateRoasterRequest{
		Name:     "Test Roaster",
		Location: "Portland, OR",
		Website:  "https://example.com",
	}

	roaster, err := store.CreateRoaster(req)
	if err != nil {
		t.Fatalf("CreateRoaster failed: %v", err)
	}

	if roaster.Name != req.Name {
		t.Errorf("CreateRoaster name = %q, want %q", roaster.Name, req.Name)
	}
	if roaster.Location != req.Location {
		t.Errorf("CreateRoaster location = %q, want %q", roaster.Location, req.Location)
	}
	if roaster.Website != req.Website {
		t.Errorf("CreateRoaster website = %q, want %q", roaster.Website, req.Website)
	}
	if roaster.RKey == "" {
		t.Error("CreateRoaster RKey is empty")
	}

	// Read
	got, err := store.GetRoasterByRKey(roaster.RKey)
	if err != nil {
		t.Fatalf("GetRoasterByRKey failed: %v", err)
	}
	if got.Name != roaster.Name {
		t.Errorf("GetRoasterByRKey name = %q, want %q", got.Name, roaster.Name)
	}

	// List
	list, err := store.ListRoasters()
	if err != nil {
		t.Fatalf("ListRoasters failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListRoasters len = %d, want 1", len(list))
	}

	// Update
	updateReq := &models.UpdateRoasterRequest{
		Name:     "Updated Roaster",
		Location: "Seattle, WA",
		Website:  "https://updated.com",
	}
	err = store.UpdateRoasterByRKey(roaster.RKey, updateReq)
	if err != nil {
		t.Fatalf("UpdateRoasterByRKey failed: %v", err)
	}

	updated, _ := store.GetRoasterByRKey(roaster.RKey)
	if updated.Name != updateReq.Name {
		t.Errorf("UpdateRoasterByRKey name = %q, want %q", updated.Name, updateReq.Name)
	}
	if updated.Location != updateReq.Location {
		t.Errorf("UpdateRoasterByRKey location = %q, want %q", updated.Location, updateReq.Location)
	}

	// Delete
	err = store.DeleteRoasterByRKey(roaster.RKey)
	if err != nil {
		t.Fatalf("DeleteRoasterByRKey failed: %v", err)
	}

	list, _ = store.ListRoasters()
	if len(list) != 0 {
		t.Errorf("After delete, ListRoasters len = %d, want 0", len(list))
	}
}

func TestGetRoasterByRKey_NotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetRoasterByRKey("999")
	if err == nil {
		t.Error("GetRoasterByRKey should return error for non-existent roaster")
	}
}

func TestGetRoasterByRKey_InvalidRKey(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetRoasterByRKey("invalid")
	if err == nil {
		t.Error("GetRoasterByRKey should return error for invalid rkey")
	}
}

// ========== Grinder CRUD Tests ==========

func TestGrinderCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create
	req := &models.CreateGrinderRequest{
		Name:        "Comandante C40",
		GrinderType: "Hand",
		BurrType:    "Conical",
		Notes:       "Great for pour over",
	}

	grinder, err := store.CreateGrinder(req)
	if err != nil {
		t.Fatalf("CreateGrinder failed: %v", err)
	}

	if grinder.Name != req.Name {
		t.Errorf("CreateGrinder name = %q, want %q", grinder.Name, req.Name)
	}
	if grinder.GrinderType != req.GrinderType {
		t.Errorf("CreateGrinder type = %q, want %q", grinder.GrinderType, req.GrinderType)
	}
	if grinder.BurrType != req.BurrType {
		t.Errorf("CreateGrinder burr_type = %q, want %q", grinder.BurrType, req.BurrType)
	}
	if grinder.Notes != req.Notes {
		t.Errorf("CreateGrinder notes = %q, want %q", grinder.Notes, req.Notes)
	}

	// Read
	got, err := store.GetGrinderByRKey(grinder.RKey)
	if err != nil {
		t.Fatalf("GetGrinderByRKey failed: %v", err)
	}
	if got.Name != grinder.Name {
		t.Errorf("GetGrinderByRKey name = %q, want %q", got.Name, grinder.Name)
	}

	// List
	list, err := store.ListGrinders()
	if err != nil {
		t.Fatalf("ListGrinders failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListGrinders len = %d, want 1", len(list))
	}

	// Update
	updateReq := &models.UpdateGrinderRequest{
		Name:        "Updated Grinder",
		GrinderType: "Electric",
		BurrType:    "Flat",
		Notes:       "Updated notes",
	}
	err = store.UpdateGrinderByRKey(grinder.RKey, updateReq)
	if err != nil {
		t.Fatalf("UpdateGrinderByRKey failed: %v", err)
	}

	updated, _ := store.GetGrinderByRKey(grinder.RKey)
	if updated.Name != updateReq.Name {
		t.Errorf("UpdateGrinderByRKey name = %q, want %q", updated.Name, updateReq.Name)
	}
	if updated.GrinderType != updateReq.GrinderType {
		t.Errorf("UpdateGrinderByRKey type = %q, want %q", updated.GrinderType, updateReq.GrinderType)
	}

	// Delete
	err = store.DeleteGrinderByRKey(grinder.RKey)
	if err != nil {
		t.Fatalf("DeleteGrinderByRKey failed: %v", err)
	}

	list, _ = store.ListGrinders()
	if len(list) != 0 {
		t.Errorf("After delete, ListGrinders len = %d, want 0", len(list))
	}
}

// ========== Brewer CRUD Tests ==========

func TestBrewerCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create
	req := &models.CreateBrewerRequest{
		Name:        "V60",
		Description: "Hario V60 pour over dripper",
	}

	brewer, err := store.CreateBrewer(req)
	if err != nil {
		t.Fatalf("CreateBrewer failed: %v", err)
	}

	if brewer.Name != req.Name {
		t.Errorf("CreateBrewer name = %q, want %q", brewer.Name, req.Name)
	}
	if brewer.Description != req.Description {
		t.Errorf("CreateBrewer description = %q, want %q", brewer.Description, req.Description)
	}

	// Read
	got, err := store.GetBrewerByRKey(brewer.RKey)
	if err != nil {
		t.Fatalf("GetBrewerByRKey failed: %v", err)
	}
	if got.Name != brewer.Name {
		t.Errorf("GetBrewerByRKey name = %q, want %q", got.Name, brewer.Name)
	}

	// List
	list, err := store.ListBrewers()
	if err != nil {
		t.Fatalf("ListBrewers failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListBrewers len = %d, want 1", len(list))
	}

	// Update
	updateReq := &models.UpdateBrewerRequest{
		Name:        "Chemex",
		Description: "Classic Chemex brewer",
	}
	err = store.UpdateBrewerByRKey(brewer.RKey, updateReq)
	if err != nil {
		t.Fatalf("UpdateBrewerByRKey failed: %v", err)
	}

	updated, _ := store.GetBrewerByRKey(brewer.RKey)
	if updated.Name != updateReq.Name {
		t.Errorf("UpdateBrewerByRKey name = %q, want %q", updated.Name, updateReq.Name)
	}
	if updated.Description != updateReq.Description {
		t.Errorf("UpdateBrewerByRKey description = %q, want %q", updated.Description, updateReq.Description)
	}

	// Delete
	err = store.DeleteBrewerByRKey(brewer.RKey)
	if err != nil {
		t.Fatalf("DeleteBrewerByRKey failed: %v", err)
	}

	list, _ = store.ListBrewers()
	if len(list) != 0 {
		t.Errorf("After delete, ListBrewers len = %d, want 0", len(list))
	}
}

// ========== Bean CRUD Tests ==========

func TestBeanCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create a roaster first for reference
	roaster, err := store.CreateRoaster(&models.CreateRoasterRequest{
		Name:     "Test Roaster",
		Location: "Portland, OR",
	})
	if err != nil {
		t.Fatalf("CreateRoaster failed: %v", err)
	}

	// Create bean
	req := &models.CreateBeanRequest{
		Name:        "Ethiopia Yirgacheffe",
		Origin:      "Ethiopia",
		RoastLevel:  "Light",
		Process:     "Washed",
		Description: "Floral and citrus notes",
		RoasterRKey: roaster.RKey,
	}

	bean, err := store.CreateBean(req)
	if err != nil {
		t.Fatalf("CreateBean failed: %v", err)
	}

	if bean.Name != req.Name {
		t.Errorf("CreateBean name = %q, want %q", bean.Name, req.Name)
	}
	if bean.Origin != req.Origin {
		t.Errorf("CreateBean origin = %q, want %q", bean.Origin, req.Origin)
	}
	if bean.RoasterRKey != roaster.RKey {
		t.Errorf("CreateBean roaster_rkey = %q, want %q", bean.RoasterRKey, roaster.RKey)
	}
	if bean.Roaster == nil || bean.Roaster.Name != roaster.Name {
		t.Error("CreateBean roaster object not populated correctly")
	}

	// Read
	got, err := store.GetBeanByRKey(bean.RKey)
	if err != nil {
		t.Fatalf("GetBeanByRKey failed: %v", err)
	}
	if got.Name != bean.Name {
		t.Errorf("GetBeanByRKey name = %q, want %q", got.Name, bean.Name)
	}
	if got.Roaster == nil || got.Roaster.Name != roaster.Name {
		t.Error("GetBeanByRKey roaster object not populated correctly")
	}

	// List
	list, err := store.ListBeans()
	if err != nil {
		t.Fatalf("ListBeans failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListBeans len = %d, want 1", len(list))
	}

	// Update
	updateReq := &models.UpdateBeanRequest{
		Name:        "Colombia Huila",
		Origin:      "Colombia",
		RoastLevel:  "Medium",
		Process:     "Natural",
		Description: "Chocolate and cherry",
		RoasterRKey: roaster.RKey,
	}
	err = store.UpdateBeanByRKey(bean.RKey, updateReq)
	if err != nil {
		t.Fatalf("UpdateBeanByRKey failed: %v", err)
	}

	updated, _ := store.GetBeanByRKey(bean.RKey)
	if updated.Name != updateReq.Name {
		t.Errorf("UpdateBeanByRKey name = %q, want %q", updated.Name, updateReq.Name)
	}
	if updated.Origin != updateReq.Origin {
		t.Errorf("UpdateBeanByRKey origin = %q, want %q", updated.Origin, updateReq.Origin)
	}

	// Delete
	err = store.DeleteBeanByRKey(bean.RKey)
	if err != nil {
		t.Fatalf("DeleteBeanByRKey failed: %v", err)
	}

	list, _ = store.ListBeans()
	if len(list) != 0 {
		t.Errorf("After delete, ListBeans len = %d, want 0", len(list))
	}
}

func TestBeanWithoutRoaster(t *testing.T) {
	store := newTestStore(t)

	// Create bean without roaster
	req := &models.CreateBeanRequest{
		Name:       "Unroasted Bean",
		Origin:     "Kenya",
		RoastLevel: "Light",
	}

	bean, err := store.CreateBean(req)
	if err != nil {
		t.Fatalf("CreateBean failed: %v", err)
	}

	if bean.RoasterRKey != "" {
		t.Errorf("CreateBean roaster_rkey = %q, want empty", bean.RoasterRKey)
	}
}

// ========== Brew CRUD Tests ==========

func TestBrewCRUD(t *testing.T) {
	store := newTestStore(t)

	// Create prerequisites
	roaster, _ := store.CreateRoaster(&models.CreateRoasterRequest{
		Name: "Test Roaster",
	})

	bean, _ := store.CreateBean(&models.CreateBeanRequest{
		Name:        "Test Bean",
		Origin:      "Ethiopia",
		RoasterRKey: roaster.RKey,
	})

	grinder, _ := store.CreateGrinder(&models.CreateGrinderRequest{
		Name:        "Test Grinder",
		GrinderType: "Hand",
	})

	brewer, _ := store.CreateBrewer(&models.CreateBrewerRequest{
		Name: "V60",
	})

	// Create brew
	req := &models.CreateBrewRequest{
		BeanRKey:     bean.RKey,
		Method:       "Pour Over",
		Temperature:  93.5,
		WaterAmount:  250,
		TimeSeconds:  180,
		GrindSize:    "Medium-Fine",
		GrinderRKey:  grinder.RKey,
		BrewerRKey:   brewer.RKey,
		TastingNotes: "Bright and fruity",
		Rating:       8,
		Pours: []models.CreatePourData{
			{WaterAmount: 50, TimeSeconds: 30},
			{WaterAmount: 100, TimeSeconds: 60},
			{WaterAmount: 100, TimeSeconds: 90},
		},
	}

	brew, err := store.CreateBrew(req, 1) // userID = 1
	if err != nil {
		t.Fatalf("CreateBrew failed: %v", err)
	}

	// Verify basic fields
	if brew.Method != req.Method {
		t.Errorf("CreateBrew method = %q, want %q", brew.Method, req.Method)
	}
	if brew.Temperature != req.Temperature {
		t.Errorf("CreateBrew temperature = %v, want %v", brew.Temperature, req.Temperature)
	}
	if brew.WaterAmount != req.WaterAmount {
		t.Errorf("CreateBrew water_amount = %d, want %d", brew.WaterAmount, req.WaterAmount)
	}
	if brew.TimeSeconds != req.TimeSeconds {
		t.Errorf("CreateBrew time_seconds = %d, want %d", brew.TimeSeconds, req.TimeSeconds)
	}
	if brew.Rating != req.Rating {
		t.Errorf("CreateBrew rating = %d, want %d", brew.Rating, req.Rating)
	}

	// Verify relationships
	if brew.BeanRKey != bean.RKey {
		t.Errorf("CreateBrew bean_rkey = %q, want %q", brew.BeanRKey, bean.RKey)
	}
	if brew.GrinderRKey != grinder.RKey {
		t.Errorf("CreateBrew grinder_rkey = %q, want %q", brew.GrinderRKey, grinder.RKey)
	}
	if brew.BrewerRKey != brewer.RKey {
		t.Errorf("CreateBrew brewer_rkey = %q, want %q", brew.BrewerRKey, brewer.RKey)
	}

	// Verify joined objects
	if brew.Bean == nil || brew.Bean.Name != "Test Bean" {
		t.Error("CreateBrew bean object not populated correctly")
	}
	if brew.GrinderObj == nil || brew.GrinderObj.Name != "Test Grinder" {
		t.Error("CreateBrew grinder object not populated correctly")
	}
	if brew.BrewerObj == nil || brew.BrewerObj.Name != "V60" {
		t.Error("CreateBrew brewer object not populated correctly")
	}

	// Verify pours
	if len(brew.Pours) != 3 {
		t.Errorf("CreateBrew pours len = %d, want 3", len(brew.Pours))
	}
	if brew.Pours[0].PourNumber != 1 {
		t.Errorf("CreateBrew pour 1 number = %d, want 1", brew.Pours[0].PourNumber)
	}
	if brew.Pours[0].WaterAmount != 50 {
		t.Errorf("CreateBrew pour 1 water = %d, want 50", brew.Pours[0].WaterAmount)
	}

	// Read
	got, err := store.GetBrewByRKey(brew.RKey)
	if err != nil {
		t.Fatalf("GetBrewByRKey failed: %v", err)
	}
	if got.Method != brew.Method {
		t.Errorf("GetBrewByRKey method = %q, want %q", got.Method, brew.Method)
	}
	if len(got.Pours) != 3 {
		t.Errorf("GetBrewByRKey pours len = %d, want 3", len(got.Pours))
	}

	// List
	list, err := store.ListBrews(1)
	if err != nil {
		t.Fatalf("ListBrews failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("ListBrews len = %d, want 1", len(list))
	}

	// Update
	updateReq := &models.CreateBrewRequest{
		BeanRKey:     bean.RKey,
		Method:       "French Press",
		Temperature:  95.0,
		WaterAmount:  300,
		TimeSeconds:  240,
		GrindSize:    "Coarse",
		GrinderRKey:  grinder.RKey,
		BrewerRKey:   brewer.RKey,
		TastingNotes: "Bold and full-bodied",
		Rating:       9,
		Pours:        []models.CreatePourData{}, // No pours for French press
	}
	err = store.UpdateBrewByRKey(brew.RKey, updateReq)
	if err != nil {
		t.Fatalf("UpdateBrewByRKey failed: %v", err)
	}

	updated, _ := store.GetBrewByRKey(brew.RKey)
	if updated.Method != updateReq.Method {
		t.Errorf("UpdateBrewByRKey method = %q, want %q", updated.Method, updateReq.Method)
	}
	if updated.Temperature != updateReq.Temperature {
		t.Errorf("UpdateBrewByRKey temperature = %v, want %v", updated.Temperature, updateReq.Temperature)
	}
	if len(updated.Pours) != 0 {
		t.Errorf("UpdateBrewByRKey pours len = %d, want 0 (pours deleted)", len(updated.Pours))
	}

	// Delete
	err = store.DeleteBrewByRKey(brew.RKey)
	if err != nil {
		t.Fatalf("DeleteBrewByRKey failed: %v", err)
	}

	list, _ = store.ListBrews(1)
	if len(list) != 0 {
		t.Errorf("After delete, ListBrews len = %d, want 0", len(list))
	}
}

func TestBrewWithoutOptionalFields(t *testing.T) {
	store := newTestStore(t)

	// Create only required entities
	bean, _ := store.CreateBean(&models.CreateBeanRequest{
		Name:   "Test Bean",
		Origin: "Ethiopia",
	})

	// Create brew without grinder or brewer
	req := &models.CreateBrewRequest{
		BeanRKey:    bean.RKey,
		Method:      "Immersion",
		Temperature: 90.0,
		TimeSeconds: 180,
		Rating:      7,
	}

	brew, err := store.CreateBrew(req, 1)
	if err != nil {
		t.Fatalf("CreateBrew failed: %v", err)
	}

	if brew.GrinderRKey != "" {
		t.Errorf("CreateBrew grinder_rkey = %q, want empty", brew.GrinderRKey)
	}
	if brew.BrewerRKey != "" {
		t.Errorf("CreateBrew brewer_rkey = %q, want empty", brew.BrewerRKey)
	}
}

func TestBrewInvalidBeanRKey(t *testing.T) {
	store := newTestStore(t)

	req := &models.CreateBrewRequest{
		BeanRKey: "invalid",
		Method:   "Pour Over",
	}

	_, err := store.CreateBrew(req, 1)
	if err == nil {
		t.Error("CreateBrew should fail with invalid bean_rkey")
	}
}

func TestListBrewsByUser(t *testing.T) {
	store := newTestStore(t)

	// Insert a second test user (user 1 'default' is created by migration)
	_, err := store.db.Exec("INSERT INTO users (id, username) VALUES (2, 'testuser2')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	bean, err := store.CreateBean(&models.CreateBeanRequest{
		Name:   "Test Bean",
		Origin: "Ethiopia",
	})
	if err != nil {
		t.Fatalf("CreateBean failed: %v", err)
	}

	// Create brews for user 1
	for i := 0; i < 3; i++ {
		_, err := store.CreateBrew(&models.CreateBrewRequest{
			BeanRKey: bean.RKey,
			Method:   "Pour Over",
			Rating:   5,
		}, 1)
		if err != nil {
			t.Fatalf("CreateBrew for user 1 failed: %v", err)
		}
	}

	// Create brews for user 2
	for i := 0; i < 2; i++ {
		_, err := store.CreateBrew(&models.CreateBrewRequest{
			BeanRKey: bean.RKey,
			Method:   "Pour Over",
			Rating:   5,
		}, 2)
		if err != nil {
			t.Fatalf("CreateBrew for user 2 failed: %v", err)
		}
	}

	// List should filter by user
	list1, err := store.ListBrews(1)
	if err != nil {
		t.Fatalf("ListBrews(1) failed: %v", err)
	}
	if len(list1) != 3 {
		t.Errorf("ListBrews(1) len = %d, want 3", len(list1))
	}

	list2, err := store.ListBrews(2)
	if err != nil {
		t.Fatalf("ListBrews(2) failed: %v", err)
	}
	if len(list2) != 2 {
		t.Errorf("ListBrews(2) len = %d, want 2", len(list2))
	}
}

// ========== Pour Tests ==========

func TestPourOperations(t *testing.T) {
	store := newTestStore(t)

	bean, _ := store.CreateBean(&models.CreateBeanRequest{
		Name:   "Test Bean",
		Origin: "Ethiopia",
	})

	brew, _ := store.CreateBrew(&models.CreateBrewRequest{
		BeanRKey: bean.RKey,
		Method:   "Pour Over",
		Rating:   5,
	}, 1)

	brewID, _ := rkeyToID(brew.RKey)

	// Create pours
	pours := []models.CreatePourData{
		{WaterAmount: 50, TimeSeconds: 30},
		{WaterAmount: 100, TimeSeconds: 60},
	}

	err := store.CreatePours(brewID, pours)
	if err != nil {
		t.Fatalf("CreatePours failed: %v", err)
	}

	// List pours
	list, err := store.ListPours(brewID)
	if err != nil {
		t.Fatalf("ListPours failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("ListPours len = %d, want 2", len(list))
	}
	if list[0].PourNumber != 1 {
		t.Errorf("Pour 1 number = %d, want 1", list[0].PourNumber)
	}
	if list[1].PourNumber != 2 {
		t.Errorf("Pour 2 number = %d, want 2", list[1].PourNumber)
	}

	// Delete pours
	err = store.DeletePoursForBrew(brewID)
	if err != nil {
		t.Fatalf("DeletePoursForBrew failed: %v", err)
	}

	list, _ = store.ListPours(brewID)
	if len(list) != 0 {
		t.Errorf("After delete, ListPours len = %d, want 0", len(list))
	}
}

func TestCreatePoursEmpty(t *testing.T) {
	store := newTestStore(t)

	// Should not error with empty slice
	err := store.CreatePours(1, []models.CreatePourData{})
	if err != nil {
		t.Errorf("CreatePours with empty slice failed: %v", err)
	}
}

// ========== Store Tests ==========

func TestNewSQLiteStore_InMemory(t *testing.T) {
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}
	defer store.Close()

	// Verify tables exist by running a query
	_, err = store.ListRoasters()
	if err != nil {
		t.Errorf("Store not properly initialized: %v", err)
	}
}

func TestStoreClose(t *testing.T) {
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteStore failed: %v", err)
	}

	err = store.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Operations should fail after close
	_, err = store.ListRoasters()
	if err == nil {
		t.Error("Operations should fail after Close")
	}
}
