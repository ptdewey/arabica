package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"arabica/internal/atproto"
	"arabica/internal/database"
	"arabica/internal/models"
	"arabica/internal/templates"
)

type Handler struct {
	store         database.Store
	oauth         *atproto.OAuthManager
	atprotoClient *atproto.Client
}

func NewHandler(store database.Store) *Handler {
	return &Handler{store: store}
}

// SetOAuthManager sets the OAuth manager for authentication
func (h *Handler) SetOAuthManager(oauth *atproto.OAuthManager) {
	h.oauth = oauth
}

// SetAtprotoClient sets the atproto client for record operations
func (h *Handler) SetAtprotoClient(client *atproto.Client) {
	h.atprotoClient = client
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

	if err := templates.RenderHome(w, isAuthenticated, didStr); err != nil {
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

	if err := templates.RenderBrewList(w, brews, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show new brew form
func (h *Handler) HandleBrewNew(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	beans, err := store.ListBeans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roasters, err := store.ListRoasters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinders, err := store.ListGrinders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewers, err := store.ListBrewers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.RenderBrewForm(w, beans, roasters, grinders, brewers, nil, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Show edit brew form
func (h *Handler) HandleBrewEdit(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	didStr, _ := atproto.GetAuthenticatedDID(r.Context())

	brew, err := store.GetBrew(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	beans, err := store.ListBeans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roasters, err := store.ListRoasters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinders, err := store.ListGrinders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewers, err := store.ListBrewers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.RenderBrewForm(w, beans, roasters, grinders, brewers, brew, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Create new brew
func (h *Handler) HandleBrewCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	beanID, _ := strconv.Atoi(r.FormValue("bean_id"))
	temperature, _ := strconv.ParseFloat(r.FormValue("temperature"), 64)
	waterAmount, _ := strconv.Atoi(r.FormValue("water_amount"))
	timeSeconds, _ := strconv.Atoi(r.FormValue("time_seconds"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	var grinderID *int
	if gIDStr := r.FormValue("grinder_id"); gIDStr != "" {
		gID, _ := strconv.Atoi(gIDStr)
		grinderID = &gID
	}

	var brewerID *int
	if bIDStr := r.FormValue("brewer_id"); bIDStr != "" {
		bID, _ := strconv.Atoi(bIDStr)
		brewerID = &bID
	}

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
		BeanID:       beanID,
		Method:       r.FormValue("method"),
		Temperature:  temperature,
		WaterAmount:  waterAmount,
		TimeSeconds:  timeSeconds,
		GrindSize:    r.FormValue("grind_size"),
		Grinder:      r.FormValue("grinder"),
		GrinderID:    grinderID,
		BrewerID:     brewerID,
		TastingNotes: r.FormValue("tasting_notes"),
		Rating:       rating,
		Pours:        pours,
	}

	// Require authentication
	store, authenticated := h.getAtprotoStore(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	brew, err := store.CreateBrew(req, 1) // User ID not used with atproto
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create pours if any
	if len(pours) > 0 {
		if err := store.CreatePours(brew.ID, pours); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect to brew list
	w.Header().Set("HX-Redirect", "/brews")
	w.WriteHeader(http.StatusOK)
}

// Update existing brew
func (h *Handler) HandleBrewUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	beanID, _ := strconv.Atoi(r.FormValue("bean_id"))
	temperature, _ := strconv.ParseFloat(r.FormValue("temperature"), 64)
	waterAmount, _ := strconv.Atoi(r.FormValue("water_amount"))
	timeSeconds, _ := strconv.Atoi(r.FormValue("time_seconds"))
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	var grinderID *int
	if gIDStr := r.FormValue("grinder_id"); gIDStr != "" {
		gID, _ := strconv.Atoi(gIDStr)
		grinderID = &gID
	}

	var brewerID *int
	if bIDStr := r.FormValue("brewer_id"); bIDStr != "" {
		bID, _ := strconv.Atoi(bIDStr)
		brewerID = &bID
	}

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
		BeanID:       beanID,
		Method:       r.FormValue("method"),
		Temperature:  temperature,
		WaterAmount:  waterAmount,
		TimeSeconds:  timeSeconds,
		GrindSize:    r.FormValue("grind_size"),
		Grinder:      r.FormValue("grinder"),
		GrinderID:    grinderID,
		BrewerID:     brewerID,
		TastingNotes: r.FormValue("tasting_notes"),
		Rating:       rating,
		Pours:        pours,
	}

	err = h.store.UpdateBrew(id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete existing pours and create new ones
	if err := h.store.DeletePoursForBrew(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pours) > 0 {
		if err := h.store.CreatePours(id, pours); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Redirect to brew list
	w.Header().Set("HX-Redirect", "/brews")
	w.WriteHeader(http.StatusOK)
}

// Delete brew
func (h *Handler) HandleBrewDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteBrew(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Export brews as JSON
func (h *Handler) HandleBrewExport(w http.ResponseWriter, r *http.Request) {
	brews, err := h.store.ListBrews(1) // Default user ID = 1
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

	beans, err := store.ListBeans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roasters, err := store.ListRoasters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinders, err := store.ListGrinders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewers, err := store.ListBrewers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.RenderManage(w, beans, roasters, grinders, brewers, authenticated, didStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Bean update/delete handlers
func (h *Handler) HandleBeanUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateBeanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateBean(id, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bean, err := h.store.GetBean(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bean)
}

func (h *Handler) HandleBeanDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteBean(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Roaster update/delete handlers
func (h *Handler) HandleRoasterUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateRoasterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateRoaster(id, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roaster, err := h.store.GetRoaster(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roaster)
}

func (h *Handler) HandleRoasterDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteRoaster(id); err != nil {
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
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateGrinderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateGrinder(id, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinder, err := h.store.GetGrinder(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grinder)
}

func (h *Handler) HandleGrinderDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteGrinder(id); err != nil {
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
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateBrewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateBrewer(id, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewer, err := h.store.GetBrewer(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(brewer)
}

func (h *Handler) HandleBrewerDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteBrewer(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
