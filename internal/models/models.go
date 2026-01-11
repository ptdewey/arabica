package models

import (
	"errors"
	"time"
)

// Field length limits for validation
const (
	MaxNameLength        = 200
	MaxLocationLength    = 200
	MaxWebsiteLength     = 500
	MaxDescriptionLength = 2000
	MaxNotesLength       = 2000
	MaxOriginLength      = 200
	MaxRoastLevelLength  = 100
	MaxProcessLength     = 100
	MaxMethodLength      = 100
	MaxGrindSizeLength   = 100
	MaxGrinderTypeLength = 50
	MaxBurrTypeLength    = 50
)

// Validation errors
var (
	ErrNameRequired    = errors.New("name is required")
	ErrNameTooLong     = errors.New("name is too long")
	ErrLocationTooLong = errors.New("location is too long")
	ErrWebsiteTooLong  = errors.New("website is too long")
	ErrDescTooLong     = errors.New("description is too long")
	ErrNotesTooLong    = errors.New("notes is too long")
	ErrOriginTooLong   = errors.New("origin is too long")
	ErrFieldTooLong    = errors.New("field value is too long")
)

type Bean struct {
	RKey        string    `json:"rkey"` // Record key (AT Protocol or stringified ID for SQLite)
	Name        string    `json:"name"`
	Origin      string    `json:"origin"`
	RoastLevel  string    `json:"roast_level"`
	Process     string    `json:"process"`
	Description string    `json:"description"`
	RoasterRKey string    `json:"roaster_rkey"` // AT Protocol reference
	CreatedAt   time.Time `json:"created_at"`

	// Joined data for display
	Roaster *Roaster `json:"roaster,omitempty"`
}

type Roaster struct {
	RKey      string    `json:"rkey"` // Record key
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Website   string    `json:"website"`
	CreatedAt time.Time `json:"created_at"`
}

type Grinder struct {
	RKey        string    `json:"rkey"` // Record key
	Name        string    `json:"name"`
	GrinderType string    `json:"grinder_type"` // Hand, Electric, Portable Electric
	BurrType    string    `json:"burr_type"`    // Conical, Flat, Blade, or empty
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Brewer struct {
	RKey        string    `json:"rkey"` // Record key
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Pour struct {
	PourNumber  int       `json:"pour_number"`
	WaterAmount int       `json:"water_amount"`
	TimeSeconds int       `json:"time_seconds"`
	CreatedAt   time.Time `json:"created_at"`
}

type Brew struct {
	RKey         string    `json:"rkey"` // Record key
	BeanRKey     string    `json:"bean_rkey"`
	Method       string    `json:"method,omitempty"`
	Temperature  float64   `json:"temperature"`
	WaterAmount  int       `json:"water_amount"`
	CoffeeAmount int       `json:"coffee_amount"`
	TimeSeconds  int       `json:"time_seconds"`
	GrindSize    string    `json:"grind_size"`
	GrinderRKey  string    `json:"grinder_rkey"`
	BrewerRKey   string    `json:"brewer_rkey"`
	TastingNotes string    `json:"tasting_notes"`
	Rating       int       `json:"rating"`
	CreatedAt    time.Time `json:"created_at"`

	// Joined data for display
	Bean       *Bean    `json:"bean,omitempty"`
	GrinderObj *Grinder `json:"grinder_obj,omitempty"`
	BrewerObj  *Brewer  `json:"brewer_obj,omitempty"`
	Pours      []*Pour  `json:"pours,omitempty"`
}

type CreateBrewRequest struct {
	BeanRKey     string           `json:"bean_rkey"`
	Method       string           `json:"method"`
	Temperature  float64          `json:"temperature"`
	WaterAmount  int              `json:"water_amount"`
	CoffeeAmount int              `json:"coffee_amount"`
	TimeSeconds  int              `json:"time_seconds"`
	GrindSize    string           `json:"grind_size"`
	GrinderRKey  string           `json:"grinder_rkey"`
	BrewerRKey   string           `json:"brewer_rkey"`
	TastingNotes string           `json:"tasting_notes"`
	Rating       int              `json:"rating"`
	Pours        []CreatePourData `json:"pours"`
}

type CreatePourData struct {
	WaterAmount int `json:"water_amount"`
	TimeSeconds int `json:"time_seconds"`
}

type CreateBeanRequest struct {
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	Process     string `json:"process"`
	Description string `json:"description"`
	RoasterRKey string `json:"roaster_rkey"`
}

type CreateRoasterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

type CreateGrinderRequest struct {
	Name        string `json:"name"`
	GrinderType string `json:"grinder_type"`
	BurrType    string `json:"burr_type"`
	Notes       string `json:"notes"`
}

type CreateBrewerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateBeanRequest struct {
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	Process     string `json:"process"`
	Description string `json:"description"`
	RoasterRKey string `json:"roaster_rkey"`
}

type UpdateRoasterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

type UpdateGrinderRequest struct {
	Name        string `json:"name"`
	GrinderType string `json:"grinder_type"`
	BurrType    string `json:"burr_type"`
	Notes       string `json:"notes"`
}

type UpdateBrewerRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Validate checks that all fields are within acceptable limits
func (r *CreateBeanRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Origin) > MaxOriginLength {
		return ErrOriginTooLong
	}
	if len(r.RoastLevel) > MaxRoastLevelLength {
		return ErrFieldTooLong
	}
	if len(r.Process) > MaxProcessLength {
		return ErrFieldTooLong
	}
	if len(r.Description) > MaxDescriptionLength {
		return ErrDescTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *UpdateBeanRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Origin) > MaxOriginLength {
		return ErrOriginTooLong
	}
	if len(r.RoastLevel) > MaxRoastLevelLength {
		return ErrFieldTooLong
	}
	if len(r.Process) > MaxProcessLength {
		return ErrFieldTooLong
	}
	if len(r.Description) > MaxDescriptionLength {
		return ErrDescTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *CreateRoasterRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Location) > MaxLocationLength {
		return ErrLocationTooLong
	}
	if len(r.Website) > MaxWebsiteLength {
		return ErrWebsiteTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *UpdateRoasterRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Location) > MaxLocationLength {
		return ErrLocationTooLong
	}
	if len(r.Website) > MaxWebsiteLength {
		return ErrWebsiteTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *CreateGrinderRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.GrinderType) > MaxGrinderTypeLength {
		return ErrFieldTooLong
	}
	if len(r.BurrType) > MaxBurrTypeLength {
		return ErrFieldTooLong
	}
	if len(r.Notes) > MaxNotesLength {
		return ErrNotesTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *UpdateGrinderRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.GrinderType) > MaxGrinderTypeLength {
		return ErrFieldTooLong
	}
	if len(r.BurrType) > MaxBurrTypeLength {
		return ErrFieldTooLong
	}
	if len(r.Notes) > MaxNotesLength {
		return ErrNotesTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *CreateBrewerRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Description) > MaxDescriptionLength {
		return ErrDescTooLong
	}
	return nil
}

// Validate checks that all fields are within acceptable limits
func (r *UpdateBrewerRequest) Validate() error {
	if r.Name == "" {
		return ErrNameRequired
	}
	if len(r.Name) > MaxNameLength {
		return ErrNameTooLong
	}
	if len(r.Description) > MaxDescriptionLength {
		return ErrDescTooLong
	}
	return nil
}
