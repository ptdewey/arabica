package atproto

import (
	"context"
	"fmt"
	"time"

	"arabica/internal/database"
	"arabica/internal/models"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

// AtprotoStore implements the database.Store interface using atproto records
type AtprotoStore struct {
	client    *Client
	did       syntax.DID
	sessionID string
	ctx       context.Context
}

// NewAtprotoStore creates a new atproto store for a specific user session
func NewAtprotoStore(ctx context.Context, client *Client, did syntax.DID, sessionID string) database.Store {
	return &AtprotoStore{
		client:    client,
		did:       did,
		sessionID: sessionID,
		ctx:       ctx,
	}
}

// ========== Brew Operations ==========

func (s *AtprotoStore) CreateBrew(brew *models.CreateBrewRequest, userID int) (*models.Brew, error) {
	// Build AT-URI references from rkeys
	if brew.BeanRKey == "" {
		return nil, fmt.Errorf("bean_rkey is required")
	}

	beanURI := fmt.Sprintf("at://%s/com.arabica.bean/%s", s.did.String(), brew.BeanRKey)

	var grinderURI, brewerURI string
	if brew.GrinderRKey != "" {
		grinderURI = fmt.Sprintf("at://%s/com.arabica.grinder/%s", s.did.String(), brew.GrinderRKey)
	}
	if brew.BrewerRKey != "" {
		brewerURI = fmt.Sprintf("at://%s/com.arabica.brewer/%s", s.did.String(), brew.BrewerRKey)
	}

	// Convert to models.Brew for record conversion
	brewModel := &models.Brew{
		BeanRKey:     brew.BeanRKey,
		GrinderRKey:  brew.GrinderRKey,
		BrewerRKey:   brew.BrewerRKey,
		Method:       brew.Method,
		Temperature:  brew.Temperature,
		WaterAmount:  brew.WaterAmount,
		TimeSeconds:  brew.TimeSeconds,
		GrindSize:    brew.GrindSize,
		TastingNotes: brew.TastingNotes,
		Rating:       brew.Rating,
		CreatedAt:    time.Now(),
	}

	// Convert pours
	if len(brew.Pours) > 0 {
		brewModel.Pours = make([]*models.Pour, len(brew.Pours))
		for i, pour := range brew.Pours {
			brewModel.Pours[i] = &models.Pour{
				WaterAmount: pour.WaterAmount,
				TimeSeconds: pour.TimeSeconds,
			}
		}
	}

	// Convert to atproto record
	record, err := BrewToRecord(brewModel, beanURI, grinderURI, brewerURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert brew to record: %w", err)
	}

	// Create record in PDS
	output, err := s.client.CreateRecord(s.ctx, s.did, s.sessionID, &CreateRecordInput{
		Collection: "com.arabica.brew",
		Record:     record,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create brew record: %w", err)
	}

	// Parse the returned AT-URI to get the rkey
	atURI, err := syntax.ParseATURI(output.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse returned AT-URI: %w", err)
	}

	// Store the rkey in the model
	rkey := atURI.RecordKey().String()
	brewModel.RKey = rkey
	brewModel.ID = 0 // ID no longer used for atproto records

	// Fetch and resolve references to populate Bean, Grinder, Brewer
	err = ResolveBrewRefs(s.ctx, s.client, brewModel, beanURI, grinderURI, brewerURI, s.sessionID)
	if err != nil {
		// Non-fatal: return the brew even if we can't resolve refs
		fmt.Printf("Warning: failed to resolve brew references: %v\n", err)
	}

	return brewModel, nil
}

func (s *AtprotoStore) GetBrew(id int) (*models.Brew, error) {
	// Convert ID to rkey
	// TODO: Implement proper ID -> rkey mapping
	rkey := fmt.Sprintf("%d", id)

	output, err := s.client.GetRecord(s.ctx, s.did, s.sessionID, &GetRecordInput{
		Collection: "com.arabica.brew",
		RKey:       rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get brew record: %w", err)
	}

	// Build the AT-URI for this brew
	atURI := fmt.Sprintf("at://%s/com.arabica.brew/%s", s.did.String(), rkey)

	// Convert to models.Brew
	brew, err := RecordToBrew(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert brew record: %w", err)
	}

	// Extract and resolve references
	beanRef, _ := output.Value["beanRef"].(string)
	grinderRef, _ := output.Value["grinderRef"].(string)
	brewerRef, _ := output.Value["brewerRef"].(string)

	err = ResolveBrewRefs(s.ctx, s.client, brew, beanRef, grinderRef, brewerRef, s.sessionID)
	if err != nil {
		fmt.Printf("Warning: failed to resolve brew references: %v\n", err)
	}

	return brew, nil
}

func (s *AtprotoStore) ListBrews(userID int) ([]*models.Brew, error) {
	output, err := s.client.ListRecords(s.ctx, s.did, s.sessionID, &ListRecordsInput{
		Collection: "com.arabica.brew",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list brew records: %w", err)
	}

	brews := make([]*models.Brew, 0, len(output.Records))

	for _, rec := range output.Records {
		brew, err := RecordToBrew(rec.Value, rec.URI)
		if err != nil {
			fmt.Printf("Warning: failed to convert brew record %s: %v\n", rec.URI, err)
			continue
		}

		// Extract and resolve references
		beanRef, _ := rec.Value["beanRef"].(string)
		grinderRef, _ := rec.Value["grinderRef"].(string)
		brewerRef, _ := rec.Value["brewerRef"].(string)

		err = ResolveBrewRefs(s.ctx, s.client, brew, beanRef, grinderRef, brewerRef, s.sessionID)
		if err != nil {
			fmt.Printf("Warning: failed to resolve brew references for %s: %v\n", rec.URI, err)
		}

		brews = append(brews, brew)
	}

	return brews, nil
}

func (s *AtprotoStore) UpdateBrew(id int, brew *models.CreateBrewRequest) error {
	// Convert ID to rkey
	rkey := fmt.Sprintf("%d", id)

	// Build AT-URI references
	beanURI := fmt.Sprintf("at://%s/com.arabica.bean/%d", s.did.String(), brew.BeanID)

	var grinderURI, brewerURI string
	if brew.GrinderID != nil {
		grinderURI = fmt.Sprintf("at://%s/com.arabica.grinder/%d", s.did.String(), *brew.GrinderID)
	}
	if brew.BrewerID != nil {
		brewerURI = fmt.Sprintf("at://%s/com.arabica.brewer/%d", s.did.String(), *brew.BrewerID)
	}

	// Convert to models.Brew
	brewModel := &models.Brew{
		BeanID:       brew.BeanID,
		Method:       brew.Method,
		Temperature:  brew.Temperature,
		WaterAmount:  brew.WaterAmount,
		TimeSeconds:  brew.TimeSeconds,
		GrindSize:    brew.GrindSize,
		GrinderID:    brew.GrinderID,
		BrewerID:     brew.BrewerID,
		TastingNotes: brew.TastingNotes,
		Rating:       brew.Rating,
		CreatedAt:    time.Now(), // Keep original creation time in production
	}

	// Convert pours
	if len(brew.Pours) > 0 {
		brewModel.Pours = make([]*models.Pour, len(brew.Pours))
		for i, pour := range brew.Pours {
			brewModel.Pours[i] = &models.Pour{
				WaterAmount: pour.WaterAmount,
				TimeSeconds: pour.TimeSeconds,
			}
		}
	}

	// Convert to atproto record
	record, err := BrewToRecord(brewModel, beanURI, grinderURI, brewerURI)
	if err != nil {
		return fmt.Errorf("failed to convert brew to record: %w", err)
	}

	// Update record in PDS
	err = s.client.PutRecord(s.ctx, s.did, s.sessionID, &PutRecordInput{
		Collection: "com.arabica.brew",
		RKey:       rkey,
		Record:     record,
	})
	if err != nil {
		return fmt.Errorf("failed to update brew record: %w", err)
	}

	return nil
}

func (s *AtprotoStore) DeleteBrew(id int) error {
	rkey := fmt.Sprintf("%d", id)

	err := s.client.DeleteRecord(s.ctx, s.did, s.sessionID, &DeleteRecordInput{
		Collection: "com.arabica.brew",
		RKey:       rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete brew record: %w", err)
	}

	return nil
}

// ========== Bean Operations ==========

func (s *AtprotoStore) CreateBean(bean *models.CreateBeanRequest) (*models.Bean, error) {
	var roasterURI string
	if bean.RoasterRKey != "" {
		roasterURI = fmt.Sprintf("at://%s/com.arabica.roaster/%s", s.did.String(), bean.RoasterRKey)
	}

	beanModel := &models.Bean{
		Name:        bean.Name,
		Origin:      bean.Origin,
		RoastLevel:  bean.RoastLevel,
		Process:     bean.Process,
		Description: bean.Description,
		RoasterRKey: bean.RoasterRKey,
		CreatedAt:   time.Now(),
	}

	record, err := BeanToRecord(beanModel, roasterURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bean to record: %w", err)
	}

	output, err := s.client.CreateRecord(s.ctx, s.did, s.sessionID, &CreateRecordInput{
		Collection: "com.arabica.bean",
		Record:     record,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bean record: %w", err)
	}

	atURI, err := syntax.ParseATURI(output.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse returned AT-URI: %w", err)
	}

	// Store the rkey in the model
	rkey := atURI.RecordKey().String()
	beanModel.RKey = rkey
	beanModel.ID = 0 // ID no longer used for atproto records

	return beanModel, nil
}

func (s *AtprotoStore) GetBean(id int) (*models.Bean, error) {
	rkey := fmt.Sprintf("%d", id)

	output, err := s.client.GetRecord(s.ctx, s.did, s.sessionID, &GetRecordInput{
		Collection: "com.arabica.bean",
		RKey:       rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bean record: %w", err)
	}

	atURI := fmt.Sprintf("at://%s/com.arabica.bean/%s", s.did.String(), rkey)
	bean, err := RecordToBean(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bean record: %w", err)
	}

	// Resolve roaster reference if present
	if roasterRef, ok := output.Value["roasterRef"].(string); ok && roasterRef != "" {
		// Only try to resolve if it looks like a valid AT-URI (not just "0" or a numeric ID)
		if len(roasterRef) > 10 && (roasterRef[:5] == "at://" || roasterRef[:4] == "did:") {
			bean.Roaster, err = ResolveRoasterRef(s.ctx, s.client, roasterRef, s.sessionID)
			if err != nil {
				fmt.Printf("Warning: failed to resolve roaster reference: %v\n", err)
			}
		}
	}

	return bean, nil
}

func (s *AtprotoStore) ListBeans() ([]*models.Bean, error) {
	output, err := s.client.ListRecords(s.ctx, s.did, s.sessionID, &ListRecordsInput{
		Collection: "com.arabica.bean",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list bean records: %w", err)
	}

	beans := make([]*models.Bean, 0, len(output.Records))

	for _, rec := range output.Records {
		bean, err := RecordToBean(rec.Value, rec.URI)
		if err != nil {
			fmt.Printf("Warning: failed to convert bean record %s: %v\n", rec.URI, err)
			continue
		}

		// Resolve roaster reference if present
		if roasterRef, ok := rec.Value["roasterRef"].(string); ok && roasterRef != "" {
			// Only try to resolve if it looks like a valid AT-URI
			if len(roasterRef) > 10 && (roasterRef[:5] == "at://" || roasterRef[:4] == "did:") {
				bean.Roaster, err = ResolveRoasterRef(s.ctx, s.client, roasterRef, s.sessionID)
				if err != nil {
					fmt.Printf("Warning: failed to resolve roaster reference: %v\n", err)
				}
			}
		}

		beans = append(beans, bean)
	}

	return beans, nil
}

func (s *AtprotoStore) UpdateBean(id int, bean *models.UpdateBeanRequest) error {
	rkey := fmt.Sprintf("%d", id)

	var roasterURI string
	if bean.RoasterID != nil {
		roasterURI = fmt.Sprintf("at://%s/com.arabica.roaster/%d", s.did.String(), *bean.RoasterID)
	}

	beanModel := &models.Bean{
		Name:        bean.Name,
		Origin:      bean.Origin,
		RoastLevel:  bean.RoastLevel,
		Process:     bean.Process,
		Description: bean.Description,
		RoasterID:   bean.RoasterID,
		CreatedAt:   time.Now(),
	}

	record, err := BeanToRecord(beanModel, roasterURI)
	if err != nil {
		return fmt.Errorf("failed to convert bean to record: %w", err)
	}

	err = s.client.PutRecord(s.ctx, s.did, s.sessionID, &PutRecordInput{
		Collection: "com.arabica.bean",
		RKey:       rkey,
		Record:     record,
	})
	if err != nil {
		return fmt.Errorf("failed to update bean record: %w", err)
	}

	return nil
}

func (s *AtprotoStore) DeleteBean(id int) error {
	rkey := fmt.Sprintf("%d", id)

	err := s.client.DeleteRecord(s.ctx, s.did, s.sessionID, &DeleteRecordInput{
		Collection: "com.arabica.bean",
		RKey:       rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete bean record: %w", err)
	}

	return nil
}

// ========== Roaster Operations ==========

func (s *AtprotoStore) CreateRoaster(roaster *models.CreateRoasterRequest) (*models.Roaster, error) {
	roasterModel := &models.Roaster{
		Name:      roaster.Name,
		Location:  roaster.Location,
		Website:   roaster.Website,
		CreatedAt: time.Now(),
	}

	record, err := RoasterToRecord(roasterModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert roaster to record: %w", err)
	}

	output, err := s.client.CreateRecord(s.ctx, s.did, s.sessionID, &CreateRecordInput{
		Collection: "com.arabica.roaster",
		Record:     record,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create roaster record: %w", err)
	}

	atURI, err := syntax.ParseATURI(output.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse returned AT-URI: %w", err)
	}

	// Store the rkey in the model
	rkey := atURI.RecordKey().String()
	roasterModel.RKey = rkey
	roasterModel.ID = 0 // ID no longer used for atproto records

	return roasterModel, nil
}

func (s *AtprotoStore) GetRoaster(id int) (*models.Roaster, error) {
	rkey := fmt.Sprintf("%d", id)

	output, err := s.client.GetRecord(s.ctx, s.did, s.sessionID, &GetRecordInput{
		Collection: "com.arabica.roaster",
		RKey:       rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get roaster record: %w", err)
	}

	atURI := fmt.Sprintf("at://%s/com.arabica.roaster/%s", s.did.String(), rkey)
	roaster, err := RecordToRoaster(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert roaster record: %w", err)
	}

	return roaster, nil
}

func (s *AtprotoStore) ListRoasters() ([]*models.Roaster, error) {
	output, err := s.client.ListRecords(s.ctx, s.did, s.sessionID, &ListRecordsInput{
		Collection: "com.arabica.roaster",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list roaster records: %w", err)
	}

	roasters := make([]*models.Roaster, 0, len(output.Records))

	for _, rec := range output.Records {
		roaster, err := RecordToRoaster(rec.Value, rec.URI)
		if err != nil {
			fmt.Printf("Warning: failed to convert roaster record %s: %v\n", rec.URI, err)
			continue
		}
		roasters = append(roasters, roaster)
	}

	return roasters, nil
}

func (s *AtprotoStore) UpdateRoaster(id int, roaster *models.UpdateRoasterRequest) error {
	rkey := fmt.Sprintf("%d", id)

	roasterModel := &models.Roaster{
		Name:      roaster.Name,
		Location:  roaster.Location,
		Website:   roaster.Website,
		CreatedAt: time.Now(),
	}

	record, err := RoasterToRecord(roasterModel)
	if err != nil {
		return fmt.Errorf("failed to convert roaster to record: %w", err)
	}

	err = s.client.PutRecord(s.ctx, s.did, s.sessionID, &PutRecordInput{
		Collection: "com.arabica.roaster",
		RKey:       rkey,
		Record:     record,
	})
	if err != nil {
		return fmt.Errorf("failed to update roaster record: %w", err)
	}

	return nil
}

func (s *AtprotoStore) DeleteRoaster(id int) error {
	rkey := fmt.Sprintf("%d", id)

	err := s.client.DeleteRecord(s.ctx, s.did, s.sessionID, &DeleteRecordInput{
		Collection: "com.arabica.roaster",
		RKey:       rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete roaster record: %w", err)
	}

	return nil
}

// ========== Grinder Operations ==========

func (s *AtprotoStore) CreateGrinder(grinder *models.CreateGrinderRequest) (*models.Grinder, error) {
	grinderModel := &models.Grinder{
		Name:        grinder.Name,
		GrinderType: grinder.GrinderType,
		BurrType:    grinder.BurrType,
		Notes:       grinder.Notes,
		CreatedAt:   time.Now(),
	}

	record, err := GrinderToRecord(grinderModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert grinder to record: %w", err)
	}

	output, err := s.client.CreateRecord(s.ctx, s.did, s.sessionID, &CreateRecordInput{
		Collection: "com.arabica.grinder",
		Record:     record,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create grinder record: %w", err)
	}

	atURI, err := syntax.ParseATURI(output.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse returned AT-URI: %w", err)
	}

	// Store the rkey in the model
	rkey := atURI.RecordKey().String()
	grinderModel.RKey = rkey
	grinderModel.ID = 0 // ID no longer used for atproto records

	return grinderModel, nil
}

func (s *AtprotoStore) GetGrinder(id int) (*models.Grinder, error) {
	rkey := fmt.Sprintf("%d", id)

	output, err := s.client.GetRecord(s.ctx, s.did, s.sessionID, &GetRecordInput{
		Collection: "com.arabica.grinder",
		RKey:       rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get grinder record: %w", err)
	}

	atURI := fmt.Sprintf("at://%s/com.arabica.grinder/%s", s.did.String(), rkey)
	grinder, err := RecordToGrinder(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert grinder record: %w", err)
	}

	return grinder, nil
}

func (s *AtprotoStore) ListGrinders() ([]*models.Grinder, error) {
	output, err := s.client.ListRecords(s.ctx, s.did, s.sessionID, &ListRecordsInput{
		Collection: "com.arabica.grinder",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list grinder records: %w", err)
	}

	grinders := make([]*models.Grinder, 0, len(output.Records))

	for _, rec := range output.Records {
		grinder, err := RecordToGrinder(rec.Value, rec.URI)
		if err != nil {
			fmt.Printf("Warning: failed to convert grinder record %s: %v\n", rec.URI, err)
			continue
		}
		grinders = append(grinders, grinder)
	}

	return grinders, nil
}

func (s *AtprotoStore) UpdateGrinder(id int, grinder *models.UpdateGrinderRequest) error {
	rkey := fmt.Sprintf("%d", id)

	grinderModel := &models.Grinder{
		Name:        grinder.Name,
		GrinderType: grinder.GrinderType,
		BurrType:    grinder.BurrType,
		Notes:       grinder.Notes,
		CreatedAt:   time.Now(),
	}

	record, err := GrinderToRecord(grinderModel)
	if err != nil {
		return fmt.Errorf("failed to convert grinder to record: %w", err)
	}

	err = s.client.PutRecord(s.ctx, s.did, s.sessionID, &PutRecordInput{
		Collection: "com.arabica.grinder",
		RKey:       rkey,
		Record:     record,
	})
	if err != nil {
		return fmt.Errorf("failed to update grinder record: %w", err)
	}

	return nil
}

func (s *AtprotoStore) DeleteGrinder(id int) error {
	rkey := fmt.Sprintf("%d", id)

	err := s.client.DeleteRecord(s.ctx, s.did, s.sessionID, &DeleteRecordInput{
		Collection: "com.arabica.grinder",
		RKey:       rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete grinder record: %w", err)
	}

	return nil
}

// ========== Brewer Operations ==========

func (s *AtprotoStore) CreateBrewer(brewer *models.CreateBrewerRequest) (*models.Brewer, error) {
	brewerModel := &models.Brewer{
		Name:        brewer.Name,
		Description: brewer.Description,
		CreatedAt:   time.Now(),
	}

	record, err := BrewerToRecord(brewerModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert brewer to record: %w", err)
	}

	output, err := s.client.CreateRecord(s.ctx, s.did, s.sessionID, &CreateRecordInput{
		Collection: "com.arabica.brewer",
		Record:     record,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create brewer record: %w", err)
	}

	atURI, err := syntax.ParseATURI(output.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse returned AT-URI: %w", err)
	}

	// Store the rkey in the model
	rkey := atURI.RecordKey().String()
	brewerModel.RKey = rkey
	brewerModel.ID = 0 // ID no longer used for atproto records

	return brewerModel, nil
}

func (s *AtprotoStore) GetBrewer(id int) (*models.Brewer, error) {
	rkey := fmt.Sprintf("%d", id)

	output, err := s.client.GetRecord(s.ctx, s.did, s.sessionID, &GetRecordInput{
		Collection: "com.arabica.brewer",
		RKey:       rkey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get brewer record: %w", err)
	}

	atURI := fmt.Sprintf("at://%s/com.arabica.brewer/%s", s.did.String(), rkey)
	brewer, err := RecordToBrewer(output.Value, atURI)
	if err != nil {
		return nil, fmt.Errorf("failed to convert brewer record: %w", err)
	}

	return brewer, nil
}

func (s *AtprotoStore) ListBrewers() ([]*models.Brewer, error) {
	output, err := s.client.ListRecords(s.ctx, s.did, s.sessionID, &ListRecordsInput{
		Collection: "com.arabica.brewer",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list brewer records: %w", err)
	}

	brewers := make([]*models.Brewer, 0, len(output.Records))

	for _, rec := range output.Records {
		brewer, err := RecordToBrewer(rec.Value, rec.URI)
		if err != nil {
			fmt.Printf("Warning: failed to convert brewer record %s: %v\n", rec.URI, err)
			continue
		}
		brewers = append(brewers, brewer)
	}

	return brewers, nil
}

func (s *AtprotoStore) UpdateBrewer(id int, brewer *models.UpdateBrewerRequest) error {
	rkey := fmt.Sprintf("%d", id)

	brewerModel := &models.Brewer{
		Name:        brewer.Name,
		Description: brewer.Description,
		CreatedAt:   time.Now(),
	}

	record, err := BrewerToRecord(brewerModel)
	if err != nil {
		return fmt.Errorf("failed to convert brewer to record: %w", err)
	}

	err = s.client.PutRecord(s.ctx, s.did, s.sessionID, &PutRecordInput{
		Collection: "com.arabica.brewer",
		RKey:       rkey,
		Record:     record,
	})
	if err != nil {
		return fmt.Errorf("failed to update brewer record: %w", err)
	}

	return nil
}

func (s *AtprotoStore) DeleteBrewer(id int) error {
	rkey := fmt.Sprintf("%d", id)

	err := s.client.DeleteRecord(s.ctx, s.did, s.sessionID, &DeleteRecordInput{
		Collection: "com.arabica.brewer",
		RKey:       rkey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete brewer record: %w", err)
	}

	return nil
}

// ========== Pour Operations ==========

// Note: Pours are embedded in brew records, not separate
// These operations modify the parent brew record

func (s *AtprotoStore) CreatePours(brewID int, pours []models.CreatePourData) error {
	// Get the existing brew
	brew, err := s.GetBrew(brewID)
	if err != nil {
		return fmt.Errorf("failed to get brew: %w", err)
	}

	// Add the pours to the brew
	brew.Pours = make([]*models.Pour, len(pours))
	for i, pour := range pours {
		brew.Pours[i] = &models.Pour{
			WaterAmount: pour.WaterAmount,
			TimeSeconds: pour.TimeSeconds,
			PourNumber:  i + 1,
		}
	}

	// Update the brew record with the pours
	// This is a bit awkward with the current interface design
	// In production, we might need a better approach
	return fmt.Errorf("CreatePours not yet fully implemented for atproto")
}

func (s *AtprotoStore) ListPours(brewID int) ([]*models.Pour, error) {
	// Get the brew and return its pours
	brew, err := s.GetBrew(brewID)
	if err != nil {
		return nil, fmt.Errorf("failed to get brew: %w", err)
	}

	return brew.Pours, nil
}

func (s *AtprotoStore) DeletePoursForBrew(brewID int) error {
	// Get the existing brew
	brew, err := s.GetBrew(brewID)
	if err != nil {
		return fmt.Errorf("failed to get brew: %w", err)
	}

	// Clear the pours
	brew.Pours = nil

	// Update the brew record
	// This requires re-implementing the update logic
	return fmt.Errorf("DeletePoursForBrew not yet fully implemented for atproto")
}

func (s *AtprotoStore) Close() error {
	// No persistent connection to close for atproto
	return nil
}
