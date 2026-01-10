package atproto

import (
	"fmt"
	"time"

	"arabica/internal/models"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

// ========== Brew Conversions ==========

// BrewToRecord converts a models.Brew to an atproto record map
// Note: References (beanRef, grinderRef, brewerRef) must be AT-URIs
func BrewToRecord(brew *models.Brew, beanURI, grinderURI, brewerURI string) (map[string]interface{}, error) {
	if beanURI == "" {
		return nil, fmt.Errorf("beanRef (AT-URI) is required")
	}

	record := map[string]interface{}{
		"$type":     NSIDBrew,
		"beanRef":   beanURI,
		"createdAt": brew.CreatedAt.Format(time.RFC3339),
	}

	// Optional fields
	if brew.Method != "" {
		record["method"] = brew.Method
	}
	if brew.Temperature > 0 {
		// Convert float to tenths (93.5 -> 935)
		record["temperature"] = int(brew.Temperature * 10)
	}
	if brew.WaterAmount > 0 {
		record["waterAmount"] = brew.WaterAmount
	}
	if brew.CoffeeAmount > 0 {
		record["coffeeAmount"] = brew.CoffeeAmount
	}
	if brew.TimeSeconds > 0 {
		record["timeSeconds"] = brew.TimeSeconds
	}
	if brew.GrindSize != "" {
		record["grindSize"] = brew.GrindSize
	}
	if grinderURI != "" {
		record["grinderRef"] = grinderURI
	}
	if brewerURI != "" {
		record["brewerRef"] = brewerURI
	}
	if brew.TastingNotes != "" {
		record["tastingNotes"] = brew.TastingNotes
	}
	if brew.Rating > 0 {
		record["rating"] = brew.Rating
	}

	// Convert pours to embedded array
	if len(brew.Pours) > 0 {
		pours := make([]map[string]interface{}, len(brew.Pours))
		for i, pour := range brew.Pours {
			pours[i] = map[string]interface{}{
				"waterAmount": pour.WaterAmount,
				"timeSeconds": pour.TimeSeconds,
			}
		}
		record["pours"] = pours
	}

	return record, nil
}

// RecordToBrew converts an atproto record map to a models.Brew
// The atURI parameter should be the full AT-URI of this brew record
func RecordToBrew(record map[string]interface{}, atURI string) (*models.Brew, error) {
	brew := &models.Brew{}

	// Extract rkey from AT-URI
	if atURI != "" {
		parsedURI, err := syntax.ParseATURI(atURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AT-URI: %w", err)
		}
		brew.RKey = parsedURI.RecordKey().String()
	}

	// Required field: beanRef
	beanRef, ok := record["beanRef"].(string)
	if !ok || beanRef == "" {
		return nil, fmt.Errorf("beanRef is required")
	}
	// Store the beanRef for later resolution
	// For now, we'll just note it exists but won't resolve it here

	// Required field: createdAt
	createdAtStr, ok := record["createdAt"].(string)
	if !ok {
		return nil, fmt.Errorf("createdAt is required")
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %w", err)
	}
	brew.CreatedAt = createdAt

	// Optional fields
	if method, ok := record["method"].(string); ok {
		brew.Method = method
	}
	if temp, ok := record["temperature"].(float64); ok {
		// Convert from tenths to float (935 -> 93.5)
		brew.Temperature = temp / 10.0
	}
	if waterAmount, ok := record["waterAmount"].(float64); ok {
		brew.WaterAmount = int(waterAmount)
	}
	if coffeeAmount, ok := record["coffeeAmount"].(float64); ok {
		brew.CoffeeAmount = int(coffeeAmount)
	}
	if timeSeconds, ok := record["timeSeconds"].(float64); ok {
		brew.TimeSeconds = int(timeSeconds)
	}
	if grindSize, ok := record["grindSize"].(string); ok {
		brew.GrindSize = grindSize
	}
	if tastingNotes, ok := record["tastingNotes"].(string); ok {
		brew.TastingNotes = tastingNotes
	}
	if rating, ok := record["rating"].(float64); ok {
		brew.Rating = int(rating)
	}

	// Convert pours from embedded array
	if poursRaw, ok := record["pours"].([]interface{}); ok {
		brew.Pours = make([]*models.Pour, len(poursRaw))
		for i, pourRaw := range poursRaw {
			pourMap, ok := pourRaw.(map[string]interface{})
			if !ok {
				continue
			}
			pour := &models.Pour{}
			if waterAmount, ok := pourMap["waterAmount"].(float64); ok {
				pour.WaterAmount = int(waterAmount)
			}
			if timeSeconds, ok := pourMap["timeSeconds"].(float64); ok {
				pour.TimeSeconds = int(timeSeconds)
			}
			pour.PourNumber = i + 1 // Sequential numbering
			brew.Pours[i] = pour
		}
	}

	return brew, nil
}

// ========== Bean Conversions ==========

// BeanToRecord converts a models.Bean to an atproto record map
func BeanToRecord(bean *models.Bean, roasterURI string) (map[string]interface{}, error) {
	record := map[string]interface{}{
		"$type":     NSIDBean,
		"name":      bean.Name,
		"createdAt": bean.CreatedAt.Format(time.RFC3339),
	}

	// Optional fields
	if bean.Origin != "" {
		record["origin"] = bean.Origin
	}
	if bean.RoastLevel != "" {
		record["roastLevel"] = bean.RoastLevel
	}
	if bean.Process != "" {
		record["process"] = bean.Process
	}
	if bean.Description != "" {
		record["description"] = bean.Description
	}
	if roasterURI != "" {
		record["roasterRef"] = roasterURI
	}

	return record, nil
}

// RecordToBean converts an atproto record map to a models.Bean
func RecordToBean(record map[string]interface{}, atURI string) (*models.Bean, error) {
	bean := &models.Bean{}

	// Extract rkey from AT-URI
	if atURI != "" {
		parsedURI, err := syntax.ParseATURI(atURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AT-URI: %w", err)
		}
		bean.RKey = parsedURI.RecordKey().String()
	}

	// Required field: name
	name, ok := record["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}
	bean.Name = name

	// Required field: createdAt
	createdAtStr, ok := record["createdAt"].(string)
	if !ok {
		return nil, fmt.Errorf("createdAt is required")
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %w", err)
	}
	bean.CreatedAt = createdAt

	// Optional fields
	if origin, ok := record["origin"].(string); ok {
		bean.Origin = origin
	}
	if roastLevel, ok := record["roastLevel"].(string); ok {
		bean.RoastLevel = roastLevel
	}
	if process, ok := record["process"].(string); ok {
		bean.Process = process
	}
	if description, ok := record["description"].(string); ok {
		bean.Description = description
	}

	return bean, nil
}

// ========== Roaster Conversions ==========

// RoasterToRecord converts a models.Roaster to an atproto record map
func RoasterToRecord(roaster *models.Roaster) (map[string]interface{}, error) {
	record := map[string]interface{}{
		"$type":     NSIDRoaster,
		"name":      roaster.Name,
		"createdAt": roaster.CreatedAt.Format(time.RFC3339),
	}

	// Optional fields
	if roaster.Location != "" {
		record["location"] = roaster.Location
	}
	if roaster.Website != "" {
		record["website"] = roaster.Website
	}

	return record, nil
}

// RecordToRoaster converts an atproto record map to a models.Roaster
func RecordToRoaster(record map[string]interface{}, atURI string) (*models.Roaster, error) {
	roaster := &models.Roaster{}

	// Extract rkey from AT-URI
	if atURI != "" {
		parsedURI, err := syntax.ParseATURI(atURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AT-URI: %w", err)
		}
		roaster.RKey = parsedURI.RecordKey().String()
	}

	// Required field: name
	name, ok := record["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}
	roaster.Name = name

	// Required field: createdAt
	createdAtStr, ok := record["createdAt"].(string)
	if !ok {
		return nil, fmt.Errorf("createdAt is required")
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %w", err)
	}
	roaster.CreatedAt = createdAt

	// Optional fields
	if location, ok := record["location"].(string); ok {
		roaster.Location = location
	}
	if website, ok := record["website"].(string); ok {
		roaster.Website = website
	}

	return roaster, nil
}

// ========== Grinder Conversions ==========

// GrinderToRecord converts a models.Grinder to an atproto record map
func GrinderToRecord(grinder *models.Grinder) (map[string]interface{}, error) {
	record := map[string]interface{}{
		"$type":     NSIDGrinder,
		"name":      grinder.Name,
		"createdAt": grinder.CreatedAt.Format(time.RFC3339),
	}

	// Optional fields
	if grinder.GrinderType != "" {
		record["grinderType"] = grinder.GrinderType
	}
	if grinder.BurrType != "" {
		record["burrType"] = grinder.BurrType
	}
	if grinder.Notes != "" {
		record["notes"] = grinder.Notes
	}

	return record, nil
}

// RecordToGrinder converts an atproto record map to a models.Grinder
func RecordToGrinder(record map[string]interface{}, atURI string) (*models.Grinder, error) {
	grinder := &models.Grinder{}

	// Extract rkey from AT-URI
	if atURI != "" {
		parsedURI, err := syntax.ParseATURI(atURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AT-URI: %w", err)
		}
		grinder.RKey = parsedURI.RecordKey().String()
	}

	// Required field: name
	name, ok := record["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}
	grinder.Name = name

	// Required field: createdAt
	createdAtStr, ok := record["createdAt"].(string)
	if !ok {
		return nil, fmt.Errorf("createdAt is required")
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %w", err)
	}
	grinder.CreatedAt = createdAt

	// Optional fields
	if grinderType, ok := record["grinderType"].(string); ok {
		grinder.GrinderType = grinderType
	}
	if burrType, ok := record["burrType"].(string); ok {
		grinder.BurrType = burrType
	}
	if notes, ok := record["notes"].(string); ok {
		grinder.Notes = notes
	}

	return grinder, nil
}

// ========== Brewer Conversions ==========

// BrewerToRecord converts a models.Brewer to an atproto record map
func BrewerToRecord(brewer *models.Brewer) (map[string]interface{}, error) {
	record := map[string]interface{}{
		"$type":     NSIDBrewer,
		"name":      brewer.Name,
		"createdAt": brewer.CreatedAt.Format(time.RFC3339),
	}

	// Optional fields
	if brewer.Description != "" {
		record["description"] = brewer.Description
	}

	return record, nil
}

// RecordToBrewer converts an atproto record map to a models.Brewer
func RecordToBrewer(record map[string]interface{}, atURI string) (*models.Brewer, error) {
	brewer := &models.Brewer{}

	// Extract rkey from AT-URI
	if atURI != "" {
		parsedURI, err := syntax.ParseATURI(atURI)
		if err != nil {
			return nil, fmt.Errorf("invalid AT-URI: %w", err)
		}
		brewer.RKey = parsedURI.RecordKey().String()
	}

	// Required field: name
	name, ok := record["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("name is required")
	}
	brewer.Name = name

	// Required field: createdAt
	createdAtStr, ok := record["createdAt"].(string)
	if !ok {
		return nil, fmt.Errorf("createdAt is required")
	}
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %w", err)
	}
	brewer.CreatedAt = createdAt

	// Optional fields
	if description, ok := record["description"].(string); ok {
		brewer.Description = description
	}

	return brewer, nil
}
