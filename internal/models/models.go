package models

import "time"

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
