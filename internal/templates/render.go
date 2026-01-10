package templates

import (
	"fmt"
	"html/template"
	"net/http"

	"arabica/internal/models"
)

var templates *template.Template

// Initialize loads all template files
func init() {
	var err error

	// Parse all template files including partials
	templates = template.New("")
	templates = templates.Funcs(template.FuncMap{
		"formatTemp":      formatTemp,
		"formatTime":      formatTime,
		"formatRating":    formatRating,
		"formatID":        formatID,
		"formatInt":       formatInt,
		"formatRoasterID": formatRoasterID,
		"poursToJSON":     poursToJSON,
		"ptrEquals":       ptrEquals[int],
		"ptrValue":        ptrValue[int],
	})

	// Parse all templates
	templates, err = templates.ParseGlob("internal/templates/*.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}

	// Parse partials
	templates, err = templates.ParseGlob("internal/templates/partials/*.tmpl")
	if err != nil {
		panic(fmt.Sprintf("Failed to parse partial templates: %v", err))
	}
}

// Data structures for templates
type PageData struct {
	Title           string
	Beans           []*models.Bean
	Roasters        []*models.Roaster
	Grinders        []*models.Grinder
	Brewers         []*models.Brewer
	Brew            *BrewData
	Brews           []*BrewListData
	IsAuthenticated bool
	UserDID         string
}

type BrewData struct {
	*models.Brew
	PoursJSON string
}

type BrewListData struct {
	*models.Brew
	TempFormatted   string
	TimeFormatted   string
	RatingFormatted string
}

// RenderTemplate renders a template with layout
func RenderTemplate(w http.ResponseWriter, tmpl string, data *PageData) error {
	// Execute the layout template which calls the content template
	return templates.ExecuteTemplate(w, "layout", data)
}

// RenderHome renders the home page
func RenderHome(w http.ResponseWriter, isAuthenticated bool, userDID string) error {
	data := &PageData{
		Title:           "Home",
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	// Need to execute layout with the home template
	t := template.Must(templates.Clone())
	t = template.Must(t.ParseFiles("internal/templates/home.tmpl"))
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderBrewList renders the brew list page
func RenderBrewList(w http.ResponseWriter, brews []*models.Brew, isAuthenticated bool, userDID string) error {
	brewList := make([]*BrewListData, len(brews))
	for i, brew := range brews {
		brewList[i] = &BrewListData{
			Brew:            brew,
			TempFormatted:   formatTemp(brew.Temperature),
			TimeFormatted:   formatTime(brew.TimeSeconds),
			RatingFormatted: formatRating(brew.Rating),
		}
	}

	data := &PageData{
		Title:           "All Brews",
		Brews:           brewList,
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	t := template.Must(templates.Clone())
	t = template.Must(t.ParseFiles("internal/templates/brew_list.tmpl"))
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderBrewForm renders the brew form page
func RenderBrewForm(w http.ResponseWriter, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer, brew *models.Brew, isAuthenticated bool, userDID string) error {
	var brewData *BrewData
	title := "New Brew"

	if brew != nil {
		title = "Edit Brew"
		brewData = &BrewData{
			Brew:      brew,
			PoursJSON: poursToJSON(brew.Pours),
		}
	}

	data := &PageData{
		Title:           title,
		Beans:           beans,
		Roasters:        roasters,
		Grinders:        grinders,
		Brewers:         brewers,
		Brew:            brewData,
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	t := template.Must(templates.Clone())
	t = template.Must(t.ParseFiles("internal/templates/brew_form.tmpl"))
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderManage renders the manage page
func RenderManage(w http.ResponseWriter, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer, isAuthenticated bool, userDID string) error {
	data := &PageData{
		Title:           "Manage",
		Beans:           beans,
		Roasters:        roasters,
		Grinders:        grinders,
		Brewers:         brewers,
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	t := template.Must(templates.Clone())
	t = template.Must(t.ParseFiles("internal/templates/manage.tmpl"))
	return t.ExecuteTemplate(w, "layout", data)
}
