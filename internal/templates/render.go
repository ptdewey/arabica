package templates

import (
	"html/template"
	"net/http"
	"os"
	"sync"

	"arabica/internal/models"
)

var (
	templates     *template.Template
	templatesOnce sync.Once
	templatesErr  error
)

// loadTemplates initializes templates lazily - only when first needed
func loadTemplates() error {
	templatesOnce.Do(func() {
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

		// Try to find templates relative to working directory
		// This supports both running from project root and from package directory
		paths := []string{
			"internal/templates/*.tmpl",
			"../../internal/templates/*.tmpl", // for when tests run from package dir
		}

		var err error
		for _, path := range paths {
			if _, statErr := os.Stat(path[:len(path)-6]); statErr == nil || os.IsExist(statErr) {
				templates, err = templates.ParseGlob(path)
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			templatesErr = err
			return
		}

		// Parse partials
		partialPaths := []string{
			"internal/templates/partials/*.tmpl",
			"../../internal/templates/partials/*.tmpl",
		}

		for _, path := range partialPaths {
			if _, statErr := os.Stat(path[:len(path)-6]); statErr == nil || os.IsExist(statErr) {
				templates, err = templates.ParseGlob(path)
				if err == nil {
					break
				}
			}
		}
		if err != nil {
			templatesErr = err
		}
	})
	return templatesErr
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
	if err := loadTemplates(); err != nil {
		return err
	}
	// Execute the layout template which calls the content template
	return templates.ExecuteTemplate(w, "layout", data)
}

// RenderHome renders the home page
func RenderHome(w http.ResponseWriter, isAuthenticated bool, userDID string) error {
	if err := loadTemplates(); err != nil {
		return err
	}
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
	if err := loadTemplates(); err != nil {
		return err
	}
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
	if err := loadTemplates(); err != nil {
		return err
	}
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
	if err := loadTemplates(); err != nil {
		return err
	}
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
