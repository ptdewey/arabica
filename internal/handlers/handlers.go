package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"arabica/internal/database"
	"arabica/internal/models"
	"arabica/internal/templates"
)

type Handler struct {
	store database.Store
}

func NewHandler(store database.Store) *Handler {
	return &Handler{store: store}
}

// Home page
func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	templates.Home().Render(r.Context(), w)
}

// List all brews
func (h *Handler) HandleBrewList(w http.ResponseWriter, r *http.Request) {
	brews, err := h.store.ListBrews(1) // Default user ID = 1
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.BrewList(brews).Render(r.Context(), w)
}

// Show new brew form
func (h *Handler) HandleBrewNew(w http.ResponseWriter, r *http.Request) {
	beans, err := h.store.ListBeans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinders, err := h.store.ListGrinders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewers, err := h.store.ListBrewers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.BrewForm(beans, grinders, brewers, nil).Render(r.Context(), w)
}

// Create new brew
func (h *Handler) HandleBrewCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	beanID, _ := strconv.Atoi(r.FormValue("bean_id"))
	temperature, _ := strconv.ParseFloat(r.FormValue("temperature"), 64)
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

	req := &models.CreateBrewRequest{
		BeanID:       beanID,
		Method:       r.FormValue("method"),
		Temperature:  temperature,
		TimeSeconds:  timeSeconds,
		GrindSize:    r.FormValue("grind_size"),
		Grinder:      r.FormValue("grinder"),
		GrinderID:    grinderID,
		BrewerID:     brewerID,
		TastingNotes: r.FormValue("tasting_notes"),
		Rating:       rating,
	}

	_, err := h.store.CreateBrew(req, 1) // Default user ID = 1
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

	bean, err := h.store.CreateBean(&req)
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

	roaster, err := h.store.CreateRoaster(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roaster)
}

// Manage page
func (h *Handler) HandleManage(w http.ResponseWriter, r *http.Request) {
	beans, err := h.store.ListBeans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roasters, err := h.store.ListRoasters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	grinders, err := h.store.ListGrinders()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brewers, err := h.store.ListBrewers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.ManagePage(beans, roasters, grinders, brewers).Render(r.Context(), w)
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

	grinder, err := h.store.CreateGrinder(&req)
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

	brewer, err := h.store.CreateBrewer(&req)
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
