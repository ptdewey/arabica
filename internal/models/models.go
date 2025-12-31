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
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Roaster struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Website   string    `json:"website"`
	CreatedAt time.Time `json:"created_at"`
}

type Grinder struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type Brew struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	BeanID       int       `json:"bean_id"`
	RoasterID    int       `json:"roaster_id"`
	Method       string    `json:"method"`
	Temperature  float64   `json:"temperature"`
	TimeSeconds  int       `json:"time_seconds"`
	GrindSize    string    `json:"grind_size"`
	Grinder      string    `json:"grinder"`
	TastingNotes string    `json:"tasting_notes"`
	Rating       int       `json:"rating"`
	CreatedAt    time.Time `json:"created_at"`

	// Joined data for display
	Bean    *Bean    `json:"bean,omitempty"`
	Roaster *Roaster `json:"roaster,omitempty"`
}

type CreateBrewRequest struct {
	BeanID       int     `json:"bean_id"`
	RoasterID    int     `json:"roaster_id"`
	Method       string  `json:"method"`
	Temperature  float64 `json:"temperature"`
	TimeSeconds  int     `json:"time_seconds"`
	GrindSize    string  `json:"grind_size"`
	Grinder      string  `json:"grinder"`
	TastingNotes string  `json:"tasting_notes"`
	Rating       int     `json:"rating"`
}

type CreateBeanRequest struct {
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	Description string `json:"description"`
}

type CreateRoasterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

type CreateGrinderRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Notes string `json:"notes"`
}

type UpdateBeanRequest struct {
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	RoastLevel  string `json:"roast_level"`
	Description string `json:"description"`
}

type UpdateRoasterRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	Website  string `json:"website"`
}

type UpdateGrinderRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Notes string `json:"notes"`
}
