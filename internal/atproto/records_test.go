package atproto

import (
	"testing"
	"time"

	"arabica/internal/models"
)

func TestBrewToRecord(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("full brew with all fields", func(t *testing.T) {
		brew := &models.Brew{
			Method:       "V60",
			Temperature:  93.5,
			WaterAmount:  300,
			TimeSeconds:  180,
			GrindSize:    "Medium",
			TastingNotes: "Fruity and bright",
			Rating:       8,
			CreatedAt:    createdAt,
			Pours: []*models.Pour{
				{WaterAmount: 50, TimeSeconds: 30},
				{WaterAmount: 100, TimeSeconds: 60},
			},
		}

		beanURI := "at://did:plc:test/social.arabica.alpha.bean/bean123"
		grinderURI := "at://did:plc:test/social.arabica.alpha.grinder/grinder123"
		brewerURI := "at://did:plc:test/social.arabica.alpha.brewer/brewer123"

		record, err := BrewToRecord(brew, beanURI, grinderURI, brewerURI)
		if err != nil {
			t.Fatalf("BrewToRecord() error = %v", err)
		}

		// Check required fields
		if record["$type"] != NSIDBrew {
			t.Errorf("$type = %v, want %v", record["$type"], NSIDBrew)
		}
		if record["beanRef"] != beanURI {
			t.Errorf("beanRef = %v, want %v", record["beanRef"], beanURI)
		}
		if record["createdAt"] != "2025-01-10T12:00:00Z" {
			t.Errorf("createdAt = %v, want %v", record["createdAt"], "2025-01-10T12:00:00Z")
		}

		// Check optional fields
		if record["method"] != "V60" {
			t.Errorf("method = %v, want %v", record["method"], "V60")
		}
		// Temperature should be converted to tenths (93.5 -> 935)
		if record["temperature"] != 935 {
			t.Errorf("temperature = %v, want %v", record["temperature"], 935)
		}
		if record["waterAmount"] != 300 {
			t.Errorf("waterAmount = %v, want %v", record["waterAmount"], 300)
		}
		if record["timeSeconds"] != 180 {
			t.Errorf("timeSeconds = %v, want %v", record["timeSeconds"], 180)
		}
		if record["grindSize"] != "Medium" {
			t.Errorf("grindSize = %v, want %v", record["grindSize"], "Medium")
		}
		if record["grinderRef"] != grinderURI {
			t.Errorf("grinderRef = %v, want %v", record["grinderRef"], grinderURI)
		}
		if record["brewerRef"] != brewerURI {
			t.Errorf("brewerRef = %v, want %v", record["brewerRef"], brewerURI)
		}
		if record["tastingNotes"] != "Fruity and bright" {
			t.Errorf("tastingNotes = %v, want %v", record["tastingNotes"], "Fruity and bright")
		}
		if record["rating"] != 8 {
			t.Errorf("rating = %v, want %v", record["rating"], 8)
		}

		// Check pours
		pours, ok := record["pours"].([]map[string]interface{})
		if !ok {
			t.Fatalf("pours is not []map[string]interface{}")
		}
		if len(pours) != 2 {
			t.Errorf("len(pours) = %v, want %v", len(pours), 2)
		}
		if pours[0]["waterAmount"] != 50 {
			t.Errorf("pours[0].waterAmount = %v, want %v", pours[0]["waterAmount"], 50)
		}
	})

	t.Run("minimal brew", func(t *testing.T) {
		brew := &models.Brew{
			CreatedAt: createdAt,
		}

		beanURI := "at://did:plc:test/social.arabica.alpha.bean/bean123"

		record, err := BrewToRecord(brew, beanURI, "", "")
		if err != nil {
			t.Fatalf("BrewToRecord() error = %v", err)
		}

		// Optional fields should be omitted
		if _, ok := record["method"]; ok {
			t.Error("method should be omitted when empty")
		}
		if _, ok := record["temperature"]; ok {
			t.Error("temperature should be omitted when zero")
		}
		if _, ok := record["grinderRef"]; ok {
			t.Error("grinderRef should be omitted when empty")
		}
		if _, ok := record["brewerRef"]; ok {
			t.Error("brewerRef should be omitted when empty")
		}
		if _, ok := record["pours"]; ok {
			t.Error("pours should be omitted when empty")
		}
	})

	t.Run("error without beanURI", func(t *testing.T) {
		brew := &models.Brew{
			CreatedAt: createdAt,
		}

		_, err := BrewToRecord(brew, "", "", "")
		if err == nil {
			t.Error("BrewToRecord() should error without beanURI")
		}
	})
}

func TestRecordToBrew(t *testing.T) {
	t.Run("full record", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":        NSIDBrew,
			"beanRef":      "at://did:plc:test/social.arabica.alpha.bean/bean123",
			"createdAt":    "2025-01-10T12:00:00Z",
			"method":       "V60",
			"temperature":  float64(935), // tenths
			"waterAmount":  float64(300),
			"timeSeconds":  float64(180),
			"grindSize":    "Medium",
			"grinderRef":   "at://did:plc:test/social.arabica.alpha.grinder/grinder123",
			"brewerRef":    "at://did:plc:test/social.arabica.alpha.brewer/brewer123",
			"tastingNotes": "Fruity",
			"rating":       float64(8),
			"pours": []interface{}{
				map[string]interface{}{"waterAmount": float64(50), "timeSeconds": float64(30)},
				map[string]interface{}{"waterAmount": float64(100), "timeSeconds": float64(60)},
			},
		}

		atURI := "at://did:plc:test/social.arabica.alpha.brew/brew123"
		brew, err := RecordToBrew(record, atURI)
		if err != nil {
			t.Fatalf("RecordToBrew() error = %v", err)
		}

		if brew.RKey != "brew123" {
			t.Errorf("RKey = %v, want %v", brew.RKey, "brew123")
		}
		if brew.Method != "V60" {
			t.Errorf("Method = %v, want %v", brew.Method, "V60")
		}
		// Temperature should be converted from tenths (935 -> 93.5)
		if brew.Temperature != 93.5 {
			t.Errorf("Temperature = %v, want %v", brew.Temperature, 93.5)
		}
		if brew.WaterAmount != 300 {
			t.Errorf("WaterAmount = %v, want %v", brew.WaterAmount, 300)
		}
		if brew.TimeSeconds != 180 {
			t.Errorf("TimeSeconds = %v, want %v", brew.TimeSeconds, 180)
		}
		if brew.GrindSize != "Medium" {
			t.Errorf("GrindSize = %v, want %v", brew.GrindSize, "Medium")
		}
		if brew.TastingNotes != "Fruity" {
			t.Errorf("TastingNotes = %v, want %v", brew.TastingNotes, "Fruity")
		}
		if brew.Rating != 8 {
			t.Errorf("Rating = %v, want %v", brew.Rating, 8)
		}

		if len(brew.Pours) != 2 {
			t.Fatalf("len(Pours) = %v, want %v", len(brew.Pours), 2)
		}
		if brew.Pours[0].WaterAmount != 50 {
			t.Errorf("Pours[0].WaterAmount = %v, want %v", brew.Pours[0].WaterAmount, 50)
		}
		if brew.Pours[0].PourNumber != 1 {
			t.Errorf("Pours[0].PourNumber = %v, want %v", brew.Pours[0].PourNumber, 1)
		}
	})

	t.Run("error without beanRef", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":     NSIDBrew,
			"createdAt": "2025-01-10T12:00:00Z",
		}

		_, err := RecordToBrew(record, "at://did:plc:test/social.arabica.alpha.brew/brew123")
		if err == nil {
			t.Error("RecordToBrew() should error without beanRef")
		}
	})

	t.Run("error without createdAt", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":   NSIDBrew,
			"beanRef": "at://did:plc:test/social.arabica.alpha.bean/bean123",
		}

		_, err := RecordToBrew(record, "at://did:plc:test/social.arabica.alpha.brew/brew123")
		if err == nil {
			t.Error("RecordToBrew() should error without createdAt")
		}
	})

	t.Run("error with invalid AT-URI", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":     NSIDBrew,
			"beanRef":   "at://did:plc:test/social.arabica.alpha.bean/bean123",
			"createdAt": "2025-01-10T12:00:00Z",
		}

		_, err := RecordToBrew(record, "invalid-uri")
		if err == nil {
			t.Error("RecordToBrew() should error with invalid AT-URI")
		}
	})
}

func TestBeanToRecord(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("full bean", func(t *testing.T) {
		bean := &models.Bean{
			Name:        "Ethiopian Yirgacheffe",
			Origin:      "Ethiopia",
			RoastLevel:  "Light",
			Process:     "Washed",
			Description: "Fruity and floral notes",
			CreatedAt:   createdAt,
		}

		roasterURI := "at://did:plc:test/social.arabica.alpha.roaster/roaster123"

		record, err := BeanToRecord(bean, roasterURI)
		if err != nil {
			t.Fatalf("BeanToRecord() error = %v", err)
		}

		if record["$type"] != NSIDBean {
			t.Errorf("$type = %v, want %v", record["$type"], NSIDBean)
		}
		if record["name"] != "Ethiopian Yirgacheffe" {
			t.Errorf("name = %v, want %v", record["name"], "Ethiopian Yirgacheffe")
		}
		if record["origin"] != "Ethiopia" {
			t.Errorf("origin = %v, want %v", record["origin"], "Ethiopia")
		}
		if record["roastLevel"] != "Light" {
			t.Errorf("roastLevel = %v, want %v", record["roastLevel"], "Light")
		}
		if record["process"] != "Washed" {
			t.Errorf("process = %v, want %v", record["process"], "Washed")
		}
		if record["description"] != "Fruity and floral notes" {
			t.Errorf("description = %v, want %v", record["description"], "Fruity and floral notes")
		}
		if record["roasterRef"] != roasterURI {
			t.Errorf("roasterRef = %v, want %v", record["roasterRef"], roasterURI)
		}
	})

	t.Run("bean without roaster", func(t *testing.T) {
		bean := &models.Bean{
			Name:      "Generic Coffee",
			CreatedAt: createdAt,
		}

		record, err := BeanToRecord(bean, "")
		if err != nil {
			t.Fatalf("BeanToRecord() error = %v", err)
		}

		if _, ok := record["roasterRef"]; ok {
			t.Error("roasterRef should be omitted when empty")
		}
	})
}

func TestRecordToBean(t *testing.T) {
	t.Run("full record", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":       NSIDBean,
			"name":        "Ethiopian Yirgacheffe",
			"origin":      "Ethiopia",
			"roastLevel":  "Light",
			"process":     "Washed",
			"description": "Fruity notes",
			"createdAt":   "2025-01-10T12:00:00Z",
		}

		atURI := "at://did:plc:test/social.arabica.alpha.bean/bean123"
		bean, err := RecordToBean(record, atURI)
		if err != nil {
			t.Fatalf("RecordToBean() error = %v", err)
		}

		if bean.RKey != "bean123" {
			t.Errorf("RKey = %v, want %v", bean.RKey, "bean123")
		}
		if bean.Name != "Ethiopian Yirgacheffe" {
			t.Errorf("Name = %v, want %v", bean.Name, "Ethiopian Yirgacheffe")
		}
		if bean.Origin != "Ethiopia" {
			t.Errorf("Origin = %v, want %v", bean.Origin, "Ethiopia")
		}
		if bean.RoastLevel != "Light" {
			t.Errorf("RoastLevel = %v, want %v", bean.RoastLevel, "Light")
		}
		if bean.Process != "Washed" {
			t.Errorf("Process = %v, want %v", bean.Process, "Washed")
		}
		if bean.Description != "Fruity notes" {
			t.Errorf("Description = %v, want %v", bean.Description, "Fruity notes")
		}
	})

	t.Run("error without name", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":     NSIDBean,
			"createdAt": "2025-01-10T12:00:00Z",
		}

		_, err := RecordToBean(record, "at://did:plc:test/social.arabica.alpha.bean/bean123")
		if err == nil {
			t.Error("RecordToBean() should error without name")
		}
	})
}

func TestRoasterToRecord(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("full roaster", func(t *testing.T) {
		roaster := &models.Roaster{
			Name:      "Counter Culture",
			Location:  "Durham, NC",
			Website:   "https://counterculturecoffee.com",
			CreatedAt: createdAt,
		}

		record, err := RoasterToRecord(roaster)
		if err != nil {
			t.Fatalf("RoasterToRecord() error = %v", err)
		}

		if record["$type"] != NSIDRoaster {
			t.Errorf("$type = %v, want %v", record["$type"], NSIDRoaster)
		}
		if record["name"] != "Counter Culture" {
			t.Errorf("name = %v, want %v", record["name"], "Counter Culture")
		}
		if record["location"] != "Durham, NC" {
			t.Errorf("location = %v, want %v", record["location"], "Durham, NC")
		}
		if record["website"] != "https://counterculturecoffee.com" {
			t.Errorf("website = %v, want %v", record["website"], "https://counterculturecoffee.com")
		}
	})
}

func TestRecordToRoaster(t *testing.T) {
	t.Run("full record", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":     NSIDRoaster,
			"name":      "Counter Culture",
			"location":  "Durham, NC",
			"website":   "https://counterculturecoffee.com",
			"createdAt": "2025-01-10T12:00:00Z",
		}

		atURI := "at://did:plc:test/social.arabica.alpha.roaster/roaster123"
		roaster, err := RecordToRoaster(record, atURI)
		if err != nil {
			t.Fatalf("RecordToRoaster() error = %v", err)
		}

		if roaster.RKey != "roaster123" {
			t.Errorf("RKey = %v, want %v", roaster.RKey, "roaster123")
		}
		if roaster.Name != "Counter Culture" {
			t.Errorf("Name = %v, want %v", roaster.Name, "Counter Culture")
		}
		if roaster.Location != "Durham, NC" {
			t.Errorf("Location = %v, want %v", roaster.Location, "Durham, NC")
		}
		if roaster.Website != "https://counterculturecoffee.com" {
			t.Errorf("Website = %v, want %v", roaster.Website, "https://counterculturecoffee.com")
		}
	})

	t.Run("error without name", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":     NSIDRoaster,
			"createdAt": "2025-01-10T12:00:00Z",
		}

		_, err := RecordToRoaster(record, "at://did:plc:test/social.arabica.alpha.roaster/roaster123")
		if err == nil {
			t.Error("RecordToRoaster() should error without name")
		}
	})
}

func TestGrinderToRecord(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("full grinder", func(t *testing.T) {
		grinder := &models.Grinder{
			Name:        "Comandante C40",
			GrinderType: "Hand",
			BurrType:    "Conical",
			Notes:       "Great for travel",
			CreatedAt:   createdAt,
		}

		record, err := GrinderToRecord(grinder)
		if err != nil {
			t.Fatalf("GrinderToRecord() error = %v", err)
		}

		if record["$type"] != NSIDGrinder {
			t.Errorf("$type = %v, want %v", record["$type"], NSIDGrinder)
		}
		if record["name"] != "Comandante C40" {
			t.Errorf("name = %v, want %v", record["name"], "Comandante C40")
		}
		if record["grinderType"] != "Hand" {
			t.Errorf("grinderType = %v, want %v", record["grinderType"], "Hand")
		}
		if record["burrType"] != "Conical" {
			t.Errorf("burrType = %v, want %v", record["burrType"], "Conical")
		}
		if record["notes"] != "Great for travel" {
			t.Errorf("notes = %v, want %v", record["notes"], "Great for travel")
		}
	})
}

func TestRecordToGrinder(t *testing.T) {
	t.Run("full record", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":       NSIDGrinder,
			"name":        "Comandante C40",
			"grinderType": "Hand",
			"burrType":    "Conical",
			"notes":       "Great for travel",
			"createdAt":   "2025-01-10T12:00:00Z",
		}

		atURI := "at://did:plc:test/social.arabica.alpha.grinder/grinder123"
		grinder, err := RecordToGrinder(record, atURI)
		if err != nil {
			t.Fatalf("RecordToGrinder() error = %v", err)
		}

		if grinder.RKey != "grinder123" {
			t.Errorf("RKey = %v, want %v", grinder.RKey, "grinder123")
		}
		if grinder.Name != "Comandante C40" {
			t.Errorf("Name = %v, want %v", grinder.Name, "Comandante C40")
		}
		if grinder.GrinderType != "Hand" {
			t.Errorf("GrinderType = %v, want %v", grinder.GrinderType, "Hand")
		}
		if grinder.BurrType != "Conical" {
			t.Errorf("BurrType = %v, want %v", grinder.BurrType, "Conical")
		}
		if grinder.Notes != "Great for travel" {
			t.Errorf("Notes = %v, want %v", grinder.Notes, "Great for travel")
		}
	})
}

func TestBrewerToRecord(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("full brewer", func(t *testing.T) {
		brewer := &models.Brewer{
			Name:        "Hario V60",
			Description: "Pour-over dripper",
			CreatedAt:   createdAt,
		}

		record, err := BrewerToRecord(brewer)
		if err != nil {
			t.Fatalf("BrewerToRecord() error = %v", err)
		}

		if record["$type"] != NSIDBrewer {
			t.Errorf("$type = %v, want %v", record["$type"], NSIDBrewer)
		}
		if record["name"] != "Hario V60" {
			t.Errorf("name = %v, want %v", record["name"], "Hario V60")
		}
		if record["description"] != "Pour-over dripper" {
			t.Errorf("description = %v, want %v", record["description"], "Pour-over dripper")
		}
	})
}

func TestRecordToBrewer(t *testing.T) {
	t.Run("full record", func(t *testing.T) {
		record := map[string]interface{}{
			"$type":       NSIDBrewer,
			"name":        "Hario V60",
			"description": "Pour-over dripper",
			"createdAt":   "2025-01-10T12:00:00Z",
		}

		atURI := "at://did:plc:test/social.arabica.alpha.brewer/brewer123"
		brewer, err := RecordToBrewer(record, atURI)
		if err != nil {
			t.Fatalf("RecordToBrewer() error = %v", err)
		}

		if brewer.RKey != "brewer123" {
			t.Errorf("RKey = %v, want %v", brewer.RKey, "brewer123")
		}
		if brewer.Name != "Hario V60" {
			t.Errorf("Name = %v, want %v", brewer.Name, "Hario V60")
		}
		if brewer.Description != "Pour-over dripper" {
			t.Errorf("Description = %v, want %v", brewer.Description, "Pour-over dripper")
		}
	})
}

// TestRoundTrip verifies that converting to record and back preserves data
func TestRoundTrip(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)

	t.Run("bean round trip", func(t *testing.T) {
		original := &models.Bean{
			Name:        "Ethiopian Yirgacheffe",
			Origin:      "Ethiopia",
			RoastLevel:  "Light",
			Process:     "Washed",
			Description: "Fruity notes",
			CreatedAt:   createdAt,
		}

		record, err := BeanToRecord(original, "")
		if err != nil {
			t.Fatalf("BeanToRecord() error = %v", err)
		}

		restored, err := RecordToBean(record, "at://did:plc:test/social.arabica.alpha.bean/bean123")
		if err != nil {
			t.Fatalf("RecordToBean() error = %v", err)
		}

		if restored.Name != original.Name {
			t.Errorf("Name = %v, want %v", restored.Name, original.Name)
		}
		if restored.Origin != original.Origin {
			t.Errorf("Origin = %v, want %v", restored.Origin, original.Origin)
		}
		if restored.RoastLevel != original.RoastLevel {
			t.Errorf("RoastLevel = %v, want %v", restored.RoastLevel, original.RoastLevel)
		}
	})

	t.Run("roaster round trip", func(t *testing.T) {
		original := &models.Roaster{
			Name:      "Counter Culture",
			Location:  "Durham, NC",
			Website:   "https://counterculturecoffee.com",
			CreatedAt: createdAt,
		}

		record, err := RoasterToRecord(original)
		if err != nil {
			t.Fatalf("RoasterToRecord() error = %v", err)
		}

		restored, err := RecordToRoaster(record, "at://did:plc:test/social.arabica.alpha.roaster/roaster123")
		if err != nil {
			t.Fatalf("RecordToRoaster() error = %v", err)
		}

		if restored.Name != original.Name {
			t.Errorf("Name = %v, want %v", restored.Name, original.Name)
		}
		if restored.Location != original.Location {
			t.Errorf("Location = %v, want %v", restored.Location, original.Location)
		}
		if restored.Website != original.Website {
			t.Errorf("Website = %v, want %v", restored.Website, original.Website)
		}
	})

	t.Run("grinder round trip", func(t *testing.T) {
		original := &models.Grinder{
			Name:        "Comandante C40",
			GrinderType: "Hand",
			BurrType:    "Conical",
			Notes:       "Great for travel",
			CreatedAt:   createdAt,
		}

		record, err := GrinderToRecord(original)
		if err != nil {
			t.Fatalf("GrinderToRecord() error = %v", err)
		}

		restored, err := RecordToGrinder(record, "at://did:plc:test/social.arabica.alpha.grinder/grinder123")
		if err != nil {
			t.Fatalf("RecordToGrinder() error = %v", err)
		}

		if restored.Name != original.Name {
			t.Errorf("Name = %v, want %v", restored.Name, original.Name)
		}
		if restored.GrinderType != original.GrinderType {
			t.Errorf("GrinderType = %v, want %v", restored.GrinderType, original.GrinderType)
		}
		if restored.BurrType != original.BurrType {
			t.Errorf("BurrType = %v, want %v", restored.BurrType, original.BurrType)
		}
	})

	t.Run("brewer round trip", func(t *testing.T) {
		original := &models.Brewer{
			Name:        "Hario V60",
			Description: "Pour-over dripper",
			CreatedAt:   createdAt,
		}

		record, err := BrewerToRecord(original)
		if err != nil {
			t.Fatalf("BrewerToRecord() error = %v", err)
		}

		restored, err := RecordToBrewer(record, "at://did:plc:test/social.arabica.alpha.brewer/brewer123")
		if err != nil {
			t.Fatalf("RecordToBrewer() error = %v", err)
		}

		if restored.Name != original.Name {
			t.Errorf("Name = %v, want %v", restored.Name, original.Name)
		}
		if restored.Description != original.Description {
			t.Errorf("Description = %v, want %v", restored.Description, original.Description)
		}
	})
}

func TestTemperatureConversion(t *testing.T) {
	// Test temperature encoding/decoding edge cases
	tests := []struct {
		name        string
		tempFloat   float64
		tempEncoded int
	}{
		{"zero", 0, 0},
		{"room temp", 20.0, 200},
		{"hot coffee", 93.5, 935},
		{"boiling", 100.0, 1000},
		{"fahrenheit range", 200.0, 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdAt := time.Now()
			brew := &models.Brew{
				Temperature: tt.tempFloat,
				CreatedAt:   createdAt,
			}

			record, err := BrewToRecord(brew, "at://did:plc:test/social.arabica.alpha.bean/bean123", "", "")
			if err != nil {
				t.Fatalf("BrewToRecord() error = %v", err)
			}

			// Check encoding
			if tt.tempFloat > 0 {
				encoded, ok := record["temperature"].(int)
				if !ok {
					t.Fatalf("temperature should be int, got %T", record["temperature"])
				}
				if encoded != tt.tempEncoded {
					t.Errorf("encoded temperature = %v, want %v", encoded, tt.tempEncoded)
				}
			}

			// Check decoding
			if tt.tempFloat > 0 {
				record["temperature"] = float64(tt.tempEncoded) // Simulate JSON unmarshaling
				restored, err := RecordToBrew(record, "at://did:plc:test/social.arabica.alpha.brew/brew123")
				if err != nil {
					t.Fatalf("RecordToBrew() error = %v", err)
				}
				if restored.Temperature != tt.tempFloat {
					t.Errorf("decoded temperature = %v, want %v", restored.Temperature, tt.tempFloat)
				}
			}
		})
	}
}
