package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"arabica/internal/atproto"
	"arabica/internal/bff"
	"arabica/internal/database"
	"arabica/internal/feed"
	"arabica/internal/models"
)

// Config holds handler configuration options
type Config struct {
	// SecureCookies sets the Secure flag on authentication cookies
	// Should be true in production (HTTPS), false for local development (HTTP)
	SecureCookies bool
}

type Handler struct {
	oauth         *atproto.OAuthManager
	atprotoClient *atproto.Client
	config        Config
	feedService   *feed.Service
	feedRegistry  *feed.Registry
}

func NewHandler() *Handler {
	return &Handler{}
}

// SetConfig sets the handler configuration
func (h *Handler) SetConfig(config Config) {
	h.config = config
}

// SetOAuthManager sets the OAuth manager for authentication
func (h *Handler) SetOAuthManager(oauth *atproto.OAuthManager) {
	h.oauth = oauth
}

// SetAtprotoClient sets the atproto client for record operations
func (h *Handler) SetAtprotoClient(client *atproto.Client) {
	h.atprotoClient = client
}

// SetFeedService sets the feed service for social feed
func (h *Handler) SetFeedService(service *feed.Service) {
	h.feedService = service
}

// SetFeedRegistry sets the feed registry for tracking users
func (h *Handler) SetFeedRegistry(registry *feed.Registry) {
	h.feedRegistry = registry
}

// getAtprotoStore creates a user-scoped atproto store from the request context
// Returns the store and true if authenticated, or nil and false if not authenticated
func (h *Handler) getAtprotoStore(r *http.Request) (database.Store, bool) {
	// Get authenticated DID from context
	didStr, err := atproto.GetAuthenticatedDID(r.Context())
	if err != nil {
		return nil, false
	}

	// Parse DID string to syntax.DID
	did, err := atproto.ParseDID(didStr)
	if err != nil {
		return nil, false
	}

	// Get session ID from context
	sessionID, err := atproto.GetSessionIDFromContext(r.Context())
	if err != nil {
		return nil, false
	}

	// Create user-scoped atproto store
	store := atproto.NewAtprotoStore(r.Context(), h.atprotoClient, did, sessionID)
	return store, true
}

// Home page
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	didStr, err := atproto.GetAuthenticatedDID(r.Context())
	isAuthenticated := err == nil && didStr != ""

	// Fetch feed items (if feed service is configured)
	var feedItems []*feed.FeedItem
	if h.feedService != nil {
		feedItems, _ = h.feedService.GetRecentBrews(r.Context(), 20)
	}

	if err := bff.RenderHome(w, isAuthenticated, didStr, feedItems); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// List all brews
func (h *Handler) HandleBrewList(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	brews, err := store.ListBrews(1) // User ID is not used with atproto
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := bff.RenderBrewList(w, brews, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show new brew form
func (h *Handler) HandleBrewNew(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	_, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	// Don't fetch data from PDS - client will populate dropdowns from cache
	// This makes the page load much faster
	if err := bff.RenderBrewForm(w, nil, nil, nil, nil, nil, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show edit brew form
func (h *Handler) HandleBrewEdit(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	brew, err := store.GetBrewByRKey(rkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Don't fetch dropdown data from PDS - client will populate from cache
	// This makes the page load much faster
	if err := bff.RenderBrewForm(w, nil, nil, nil, nil, brew, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Create new brew
func (h *Handler) HandleBrewCreate(w http.ResponseWriter, r *http.Request) {
	// Require authentication first
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	temperature, _ := strconv.ParseFloat(r.FormValue("temperature"), 64)
	waterAmount, _ := strconv.Atoi(r.FormValue("water_amount"))
	coffeeAmount, _ := strconv.Atoi(r.FormValue("coffee_amount"))
	timeSeconds, _ := strconv.Atoi(r.FormValue("time_seconds"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	// Parse pours
	var pours []models.CreatePourData
	i := 0
	for {
		waterKey := "pour_water_" + strconv.Itoa(i)
		timeKey := "pour_time_" + strconv.Itoa(i)

		waterStr := r.FormValue(waterKey)
		timeStr := r.FormValue(timeKey)

		if waterStr == "" && timeStr == "" {
			break
		}

		water, _ := strconv.Atoi(waterStr)
		time, _ := strconv.Atoi(timeStr)

		if water > 0 && time >= 0 {
			pours = append(pours, models.CreatePourData{
				WaterAmount: water,
				TimeSeconds: time,
			})
		}
		i++
	}

	req := &models.CreateBrewRequest{
		BeanRKey:     r.FormValue("bean_rkey"),
		Method:       r.FormValue("method"),
		Temperature:  temperature,
		WaterAmount:  waterAmount,
		CoffeeAmount: coffeeAmount,
		TimeSeconds:  timeSeconds,
		GrindSize:    r.FormValue("grind_size"),
		GrinderRKey:  r.FormValue("grinder_rkey"),
		BrewerRKey:   r.FormValue("brewer_rkey"),
		TastingNotes: r.FormValue("tasting_notes"),
		Rating:       rating,
		Pours:        pours, // Pours are embedded in the brew record for ATProto
	}

	_, err := store.CreateBrew(req, 1) // User ID not used with atproto
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to brew list
	w.Header().Set("HX-Redirect", "/brews")
	w.WriteHeader(http.StatusOK)
}

// Update existing brew
func (h *Handler) HandleBrewUpdate(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	temperature, _ := strconv.ParseFloat(r.FormValue("temperature"), 64)
	waterAmount, _ := strconv.Atoi(r.FormValue("water_amount"))
	coffeeAmount, _ := strconv.Atoi(r.FormValue("coffee_amount"))
	timeSeconds, _ := strconv.Atoi(r.FormValue("time_seconds"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	// Parse pours
	var pours []models.CreatePourData
	i := 0
	for {
		waterKey := "pour_water_" + strconv.Itoa(i)
		timeKey := "pour_time_" + strconv.Itoa(i)

		waterStr := r.FormValue(waterKey)
		timeStr := r.FormValue(timeKey)

		if waterStr == "" && timeStr == "" {
			break
		}

		water, _ := strconv.Atoi(waterStr)
		time, _ := strconv.Atoi(timeStr)

		if water > 0 && time >= 0 {
			pours = append(pours, models.CreatePourData{
				WaterAmount: water,
				TimeSeconds: time,
			})
		}
		i++
	}

	req := &models.CreateBrewRequest{
		BeanRKey:     r.FormValue("bean_rkey"),
		Method:       r.FormValue("method"),
		Temperature:  temperature,
		WaterAmount:  waterAmount,
		CoffeeAmount: coffeeAmount,
		TimeSeconds:  timeSeconds,
		GrindSize:    r.FormValue("grind_size"),
		GrinderRKey:  r.FormValue("grinder_rkey"),
		BrewerRKey:   r.FormValue("brewer_rkey"),
		TastingNotes: r.FormValue("tasting_notes"),
		Rating:       rating,
		Pours:        pours,
	}

	err := store.UpdateBrewByRKey(rkey, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to brew list
	w.Header().Set("HX-Redirect", "/brews")
	w.WriteHeader(http.StatusOK)
}

// Delete brew
func (h *Handler) HandleBrewDelete(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	if err := store.DeleteBrewByRKey(rkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Export brews as JSON
func (h *Handler) HandleBrewExport(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	brews, err := store.ListBrews(1) // User ID is not used with atproto
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=arabica-brews.json")

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(brews)
}

// API endpoint to list all user data (beans, roasters, grinders, brewers, brews)
// Used by client-side cache for faster page loads
func (h *Handler) HandleAPIListAll(w http.ResponseWriter, r *http.Request) {
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Fetch all collections in parallel
	type result struct {
		beans    []*models.Bean
		roasters []*models.Roaster
		grinders []*models.Grinder
		brewers  []*models.Brewer
		brews    []*models.Brew
		err      error
		which    string
	}

	results := make(chan result, 5)

	go func() {
		beans, err := store.ListBeans()
		results <- result{beans: beans, err: err, which: "beans"}
	}()
	go func() {
		roasters, err := store.ListRoasters()
		results <- result{roasters: roasters, err: err, which: "roasters"}
	}()
	go func() {
		grinders, err := store.ListGrinders()
		results <- result{grinders: grinders, err: err, which: "grinders"}
	}()
	go func() {
		brewers, err := store.ListBrewers()
		results <- result{brewers: brewers, err: err, which: "brewers"}
	}()
	go func() {
		brews, err := store.ListBrews(1) // User ID not used with atproto
		results <- result{brews: brews, err: err, which: "brews"}
	}()

	var beans []*models.Bean
	var roasters []*models.Roaster
	var grinders []*models.Grinder
	var brewers []*models.Brewer
	var brews []*models.Brew

	for i := 0; i < 5; i++ {
		res := <-results
		if res.err != nil {
			http.Error(w, res.err.Error(), http.StatusInternalServerError)
			return
		}
		switch res.which {
		case "beans":
			beans = res.beans
		case "roasters":
			roasters = res.roasters
		case "grinders":
			grinders = res.grinders
		case "brewers":
			brewers = res.brewers
		case "brews":
			brews = res.brews
		}
	}

	// Link beans to roasters
	atproto.LinkBeansToRoasters(beans, roasters)

	response := map[string]interface{}{
		"beans":    beans,
		"roasters": roasters,
		"grinders": grinders,
		"brewers":  brewers,
		"brews":    brews,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// API endpoint to create bean
func (h *Handler) HandleBeanCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBeanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	bean, err := store.CreateBean(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bean)
}

// API endpoint to create roaster
func (h *Handler) HandleRoasterCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	roaster, err := store.CreateRoaster(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roaster)
}

// Manage page
func (h *Handler) HandleManage(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	// Fetch all collections in parallel for better performance
	type result struct {
		beans    []*models.Bean
		roasters []*models.Roaster
		grinders []*models.Grinder
		brewers  []*models.Brewer
		err      error
		which    string
	}

	results := make(chan result, 4)

	// Launch parallel fetches
	go func() {
		beans, err := store.ListBeans()
		results <- result{beans: beans, err: err, which: "beans"}
	}()
	go func() {
		roasters, err := store.ListRoasters()
		results <- result{roasters: roasters, err: err, which: "roasters"}
	}()
	go func() {
		grinders, err := store.ListGrinders()
		results <- result{grinders: grinders, err: err, which: "grinders"}
	}()
	go func() {
		brewers, err := store.ListBrewers()
		results <- result{brewers: brewers, err: err, which: "brewers"}
	}()

	// Collect results
	var beans []*models.Bean
	var roasters []*models.Roaster
	var grinders []*models.Grinder
	var brewers []*models.Brewer

	for i := 0; i < 4; i++ {
		res := <-results
		if res.err != nil {
			http.Error(w, res.err.Error(), http.StatusInternalServerError)
			return
		}
		switch res.which {
		case "beans":
			beans = res.beans
		case "roasters":
			roasters = res.roasters
		case "grinders":
			grinders = res.grinders
		case "brewers":
			brewers = res.brewers
		}
	}

	// Link beans to their roasters using the pre-fetched roasters
	// This avoids N+1 queries when using ATProto store
	atproto.LinkBeansToRoasters(beans, roasters)

	if err := bff.RenderManage(w, beans, roasters, grinders, brewers, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Bean update/delete handlers
func (h *Handler) HandleBeanUpdate(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.UpdateBeanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.UpdateBeanByRKey(rkey, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bean, err := store.GetBeanByRKey(rkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bean)
}

func (h *Handler) HandleBeanDelete(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	if err := store.DeleteBeanByRKey(rkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Roaster update/delete handlers
func (h *Handler) HandleRoasterUpdate(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.UpdateRoasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.UpdateRoasterByRKey(rkey, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roaster, err := store.GetRoasterByRKey(rkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roaster)
}

func (h *Handler) HandleRoasterDelete(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	if err := store.DeleteRoasterByRKey(rkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Grinder CRUD handlers
func (h *Handler) HandleGrinderCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateGrinderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	grinder, err := store.CreateGrinder(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grinder)
}

func (h *Handler) HandleGrinderUpdate(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.UpdateGrinderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.UpdateGrinderByRKey(rkey, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinder, err := store.GetGrinderByRKey(rkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grinder)
}

func (h *Handler) HandleGrinderDelete(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	if err := store.DeleteGrinderByRKey(rkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Brewer CRUD handlers
func (h *Handler) HandleBrewerCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBrewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	brewer, err := store.CreateBrewer(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brewer)
}

func (h *Handler) HandleBrewerUpdate(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var req models.UpdateBrewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.UpdateBrewerByRKey(rkey, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewer, err := store.GetBrewerByRKey(rkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brewer)
}

func (h *Handler) HandleBrewerDelete(w http.ResponseWriter, r *http.Request) {
	rkey := r.PathValue("id") // URL still uses "id" path param but value is now rkey

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	if err := store.DeleteBrewerByRKey(rkey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
