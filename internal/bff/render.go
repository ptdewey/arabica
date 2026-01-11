package bff

import (
	"html/template"
	"net/http"
	"os"
	"sync"

	"arabica/internal/atproto"
	"arabica/internal/feed"
	"arabica/internal/models"
)

var (
	templateFuncs template.FuncMap
	funcsOnce     sync.Once
	templateDir   string
	templateDirMu sync.Once
)

// getTemplateFuncs returns the function map used by all templates
func getTemplateFuncs() template.FuncMap {
	funcsOnce.Do(func() {
		templateFuncs = template.FuncMap{
			"formatTemp":       FormatTemp,
			"formatTime":       FormatTime,
			"formatRating":     FormatRating,
			"formatID":         FormatID,
			"formatInt":        FormatInt,
			"formatRoasterID":  FormatRoasterID,
			"poursToJSON":      PoursToJSON,
			"ptrEquals":        PtrEquals[int],
			"ptrValue":         PtrValue[int],
			"iterate":          Iterate,
			"iterateRemaining": IterateRemaining,
			"hasTemp":          HasTemp,
			"hasValue":         HasValue,
			"safeAvatarURL":    SafeAvatarURL,
			"safeWebsiteURL":   SafeWebsiteURL,
		}
	})
	return templateFuncs
}

// getTemplateDir finds the template directory
func getTemplateDir() string {
	templateDirMu.Do(func() {
		dirs := []string{
			"templates",
			"../../templates",
			"../../../templates",
		}
		for _, dir := range dirs {
			if _, err := os.Stat(dir); err == nil {
				templateDir = dir
				return
			}
		}
		templateDir = "templates" // fallback
	})
	return templateDir
}

// parsePageTemplate parses a complete page template with layout and partials
func parsePageTemplate(pageName string) (*template.Template, error) {
	dir := getTemplateDir()
	t := template.New("").Funcs(getTemplateFuncs())

	// Parse layout first
	t, err := t.ParseFiles(dir + "/layout.tmpl")
	if err != nil {
		return nil, err
	}

	// Parse all partials
	t, err = t.ParseGlob(dir + "/partials/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Parse the specific page template
	t, err = t.ParseFiles(dir + "/" + pageName)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// parsePartialTemplate parses just the partials (for partial-only renders)
func parsePartialTemplate() (*template.Template, error) {
	dir := getTemplateDir()
	t := template.New("").Funcs(getTemplateFuncs())

	// Parse all partials
	t, err := t.ParseGlob(dir + "/partials/*.tmpl")
	if err != nil {
		return nil, err
	}

	return t, nil
}

// PageData contains data for rendering pages
type PageData struct {
	Title           string
	Beans           []*models.Bean
	Roasters        []*models.Roaster
	Grinders        []*models.Grinder
	Brewers         []*models.Brewer
	Brew            *BrewData
	Brews           []*BrewListData
	FeedItems       []*feed.FeedItem
	IsAuthenticated bool
	UserDID         string
}

// BrewData wraps a brew with pre-serialized JSON for pours
type BrewData struct {
	*models.Brew
	PoursJSON string
}

// BrewListData wraps a brew with pre-formatted display values
type BrewListData struct {
	*models.Brew
	TempFormatted   string
	TimeFormatted   string
	RatingFormatted string
}

// RenderTemplate renders a template with layout
func RenderTemplate(w http.ResponseWriter, tmpl string, data *PageData) error {
	t, err := parsePageTemplate(tmpl)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderHome renders the home page
func RenderHome(w http.ResponseWriter, isAuthenticated bool, userDID string, feedItems []*feed.FeedItem) error {
	t, err := parsePageTemplate("home.tmpl")
	if err != nil {
		return err
	}
	data := &PageData{
		Title:           "Home",
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
		FeedItems:       feedItems,
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderBrewList renders the brew list page
func RenderBrewList(w http.ResponseWriter, brews []*models.Brew, isAuthenticated bool, userDID string) error {
	t, err := parsePageTemplate("brew_list.tmpl")
	if err != nil {
		return err
	}
	brewList := make([]*BrewListData, len(brews))
	for i, brew := range brews {
		brewList[i] = &BrewListData{
			Brew:            brew,
			TempFormatted:   FormatTemp(brew.Temperature),
			TimeFormatted:   FormatTime(brew.TimeSeconds),
			RatingFormatted: FormatRating(brew.Rating),
		}
	}

	data := &PageData{
		Title:           "All Brews",
		Brews:           brewList,
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderBrewForm renders the brew form page
func RenderBrewForm(w http.ResponseWriter, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer, brew *models.Brew, isAuthenticated bool, userDID string) error {
	t, err := parsePageTemplate("brew_form.tmpl")
	if err != nil {
		return err
	}
	var brewData *BrewData
	title := "New Brew"

	if brew != nil {
		title = "Edit Brew"
		brewData = &BrewData{
			Brew:      brew,
			PoursJSON: PoursToJSON(brew.Pours),
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
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderManage renders the manage page
func RenderManage(w http.ResponseWriter, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer, isAuthenticated bool, userDID string) error {
	t, err := parsePageTemplate("manage.tmpl")
	if err != nil {
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
	return t.ExecuteTemplate(w, "layout", data)
}

// RenderFeedPartial renders just the feed partial (for HTMX async loading)
func RenderFeedPartial(w http.ResponseWriter, feedItems []*feed.FeedItem) error {
	t, err := parsePartialTemplate()
	if err != nil {
		return err
	}
	data := &PageData{
		FeedItems: feedItems,
	}
	return t.ExecuteTemplate(w, "feed", data)
}

// RenderBrewListPartial renders just the brew list partial (for HTMX async loading)
func RenderBrewListPartial(w http.ResponseWriter, brews []*models.Brew) error {
	t, err := parsePartialTemplate()
	if err != nil {
		return err
	}
	brewList := make([]*BrewListData, len(brews))
	for i, brew := range brews {
		brewList[i] = &BrewListData{
			Brew:            brew,
			TempFormatted:   FormatTemp(brew.Temperature),
			TimeFormatted:   FormatTime(brew.TimeSeconds),
			RatingFormatted: FormatRating(brew.Rating),
		}
	}

	data := &PageData{
		Brews: brewList,
	}
	return t.ExecuteTemplate(w, "brew_list_content", data)
}

// RenderManagePartial renders just the manage partial (for HTMX async loading)
func RenderManagePartial(w http.ResponseWriter, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer) error {
	t, err := parsePartialTemplate()
	if err != nil {
		return err
	}
	data := &PageData{
		Beans:    beans,
		Roasters: roasters,
		Grinders: grinders,
		Brewers:  brewers,
	}
	return t.ExecuteTemplate(w, "manage_content", data)
}

// findTemplatePath finds the correct path to a template file
func findTemplatePath(name string) string {
	dir := getTemplateDir()
	return dir + "/" + name
}

// ProfilePageData contains data for rendering the profile page
type ProfilePageData struct {
	Title           string
	Profile         *atproto.Profile
	Brews           []*models.Brew
	Beans           []*models.Bean
	Roasters        []*models.Roaster
	Grinders        []*models.Grinder
	Brewers         []*models.Brewer
	IsAuthenticated bool
	UserDID         string
}

// RenderProfile renders a user's public profile page
func RenderProfile(w http.ResponseWriter, profile *atproto.Profile, brews []*models.Brew, beans []*models.Bean, roasters []*models.Roaster, grinders []*models.Grinder, brewers []*models.Brewer, isAuthenticated bool, userDID string) error {
	t, err := parsePageTemplate("profile.tmpl")
	if err != nil {
		return err
	}

	displayName := profile.Handle
	if profile.DisplayName != nil && *profile.DisplayName != "" {
		displayName = *profile.DisplayName
	}

	data := &ProfilePageData{
		Title:           displayName + "'s Profile",
		Profile:         profile,
		Brews:           brews,
		Beans:           beans,
		Roasters:        roasters,
		Grinders:        grinders,
		Brewers:         brewers,
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// Render404 renders the 404 not found page
func Render404(w http.ResponseWriter, isAuthenticated bool, userDID string) error {
	t, err := parsePageTemplate("404.tmpl")
	if err != nil {
		return err
	}
	data := &PageData{
		Title:           "Page Not Found",
		IsAuthenticated: isAuthenticated,
		UserDID:         userDID,
	}
	w.WriteHeader(http.StatusNotFound)
	return t.ExecuteTemplate(w, "layout", data)
}
