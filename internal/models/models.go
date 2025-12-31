package models

import "time"

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Bean struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Origin      string    `json:"origin"`
	RoastLevel  string    `json:"roast_level"`
	Process     string    `json:"process"`
	Description string    `json:"description"`
	RoasterID   *int      `json:"roaster_id"`
	CreatedAt   time.Time `json:"created_at"`

	// Joined data for display
	Roaster *Roaster `json:"roaster,omitempty"`
}

type Roaster struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Website   string    `json:"website"`
	CreatedAt time.Time `json:"created_at"`
}

type Grinder struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	GrinderType string    `json:"grinder_type"` // Hand, Electric, Electric Hand
	BurrType    string    `json:"burr_type"`    // Conical, Flat, or empty
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

type Brewer struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Pour struct {
	ID          int       `json:"id"`
	BrewID      int       `json:"brew_id"`
	PourNumber  int       `json:"pour_number"`
	WaterAmount int       `json:"water_amount"`
	TimeSeconds int       `json:"time_seconds"`
	CreatedAt   time.Time `json:"created_at"`
}

type Brew struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	BeanID       int       `json:"bean_id"`
	Method       string    `json:"method,omitempty"`
	Temperature  float64   `json:"temperature"`
	WaterAmount  int       `json:"water_amount"`
	TimeSeconds  int       `json:"time_seconds"`
	GrindSize    string    `json:"grind_size"`
	Grinder      string    `json:"grinder,omitempty"`
	GrinderID    *int      `json:"grinder_id"`
	BrewerID     *int      `json:"brewer_id"`
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
	BeanID       int              `json:"bean_id"`
	Method       string           `json:"method"`
	Temperature  float64          `json:"temperature"`
	WaterAmount  int              `json:"water_amount"`
	TimeSeconds  int              `json:"time_seconds"`
	GrindSize    string           `json:"grind_size"`
	Grinder      string           `json:"grinder"`
	GrinderID    *int             `json:"grinder_id"`
	BrewerID     *int             `json:"brewer_id"`
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
	RoasterID   *int   `json:"roaster_id"`
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
	RoasterID   *int   `json:"roaster_id"`
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
