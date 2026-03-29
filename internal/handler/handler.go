package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/youruser/geo-api/internal/model"
	"github.com/youruser/geo-api/internal/repository"
)

type Handler struct {
	repo *repository.GeoRepository
}

func New(repo *repository.GeoRepository) *Handler {
	return &Handler{repo: repo}
}

// ─── Router ───────────────────────────────────────────────────────────────────

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /stats", h.Stats)

	// Countries & States
	mux.HandleFunc("GET /countries", h.GetCountries)
	mux.HandleFunc("GET /countries/{iso2}", h.GetCountry)
	mux.HandleFunc("GET /countries/{iso2}/states", h.GetStatesByCountry)
	mux.HandleFunc("GET /countries/{iso2}/cities", h.GetCitiesByCountry)

	// States
	mux.HandleFunc("GET /states/{id}", h.GetState)
	mux.HandleFunc("GET /states/{id}/cities", h.GetCitiesByState)

	// Cities
	mux.HandleFunc("GET /cities/{id}", h.GetCity)

	// Documentation
	mux.HandleFunc("GET /openapi.yaml", h.ServeOpenAPI)
	mux.HandleFunc("GET /docs", h.ServeDocs)
}

func (h *Handler) RegisterAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /admin/cities", BasicAuth(h.GetAllCities))
	mux.HandleFunc("GET /admin/cities/export", BasicAuth(h.ExportCitiesCSV))
	mux.HandleFunc("GET /admin/metadata/download", BasicAuth(h.DownloadMetadata))
	mux.HandleFunc("PUT /admin/cities/{id}/metadata", BasicAuth(h.UpdateCityMetadata))
	mux.HandleFunc("POST /admin/cities/bulk-metadata", BasicAuth(h.BulkUpdateMetadata))
}

// ─── Health & Stats ───────────────────────────────────────────────────────────

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.repo.GetStats())
}

// ─── Documentation ────────────────────────────────────────────────────────────

func (h *Handler) ServeOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "openapi.yaml")
}

func (h *Handler) ServeDocs(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/docs.html")
}

// ─── Countries ────────────────────────────────────────────────────────────────

// GET /countries?search=col&page=1&limit=20
func (h *Handler) GetCountries(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	countries := h.repo.GetCountries(search)
	writeJSON(w, http.StatusOK, paginate(r, countries))
}

// GET /countries/{iso2}
func (h *Handler) GetCountry(w http.ResponseWriter, r *http.Request) {
	iso2 := r.PathValue("iso2")
	country, ok := h.repo.GetCountryByISO2(iso2)
	if !ok {
		writeError(w, http.StatusNotFound, "country not found")
		return
	}
	writeJSON(w, http.StatusOK, country)
}

// GET /countries/{iso2}/states?search=ant
func (h *Handler) GetStatesByCountry(w http.ResponseWriter, r *http.Request) {
	iso2 := r.PathValue("iso2")
	search := r.URL.Query().Get("search")
	states, ok := h.repo.GetStatesByCountry(iso2, search)
	if !ok {
		writeError(w, http.StatusNotFound, "country not found")
		return
	}
	writeJSON(w, http.StatusOK, paginate(r, states))
}

// GET /countries/{iso2}/cities?search=med
func (h *Handler) GetCitiesByCountry(w http.ResponseWriter, r *http.Request) {
	iso2 := r.PathValue("iso2")
	search := r.URL.Query().Get("search")
	cities, ok := h.repo.GetCitiesByCountry(iso2, search)
	if !ok {
		writeError(w, http.StatusNotFound, "country not found")
		return
	}
	writeJSON(w, http.StatusOK, paginate(r, cities))
}

// ─── States ───────────────────────────────────────────────────────────────────

// GET /states/{id}
func (h *Handler) GetState(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid state id")
		return
	}
	state, ok := h.repo.GetStateByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "state not found")
		return
	}
	writeJSON(w, http.StatusOK, state)
}

// GET /states/{id}/cities?search=med
func (h *Handler) GetCitiesByState(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid state id")
		return
	}
	search := r.URL.Query().Get("search")
	cities, ok := h.repo.GetCitiesByState(id, search)
	if !ok {
		writeError(w, http.StatusNotFound, "state not found")
		return
	}
	writeJSON(w, http.StatusOK, paginate(r, cities))
}

// ─── Cities ───────────────────────────────────────────────────────────────────

func (h *Handler) GetCity(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid city id")
		return
	}
	city, ok := h.repo.GetCityByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "city not found")
		return
	}
	writeJSON(w, http.StatusOK, city)
}

// ─── Admin ────────────────────────────────────────────────────────────────────

func (h *Handler) GetAllCities(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	country := r.URL.Query().Get("country")
	state := queryInt(r, "state", 0)
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 50)

	cities, total := h.repo.GetAllCities(search, country, state, page, limit)

	writeJSON(w, http.StatusOK, model.PaginatedResponse{
		Data:  cities,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *Handler) UpdateCityMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid city id")
		return
	}

	var metadata map[string]any
	if err := json.NewDecoder(r.Body).Decode(&metadata); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.repo.UpdateCityMetadata(id, metadata); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) ExportCitiesCSV(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	country := r.URL.Query().Get("country")
	state := queryInt(r, "state", 0)

	cities, _ := h.repo.GetAllCities(search, country, state, 1, 1000000)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=cities.csv")

	w.Write([]byte("id,name,country_code,state_code,latitude,longitude\n"))
	for _, city := range cities {
		w.Write([]byte(fmt.Sprintf("%d,%s,%s,%s,%s,%s\n",
			city.ID,
			escapeCSV(city.Name),
			city.CountryCode,
			city.StateCode,
			city.Latitude,
			city.Longitude,
		)))
	}
}

func (h *Handler) BulkUpdateMetadata(w http.ResponseWriter, r *http.Request) {
	var req model.BulkMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if len(req.Updates) == 0 {
		writeError(w, http.StatusBadRequest, "no updates provided")
		return
	}

	result := h.repo.BulkUpdateMetadata(req.Updates)
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) DownloadMetadata(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.repo.GetMetadataPath())
}

func escapeCSV(s string) string {
	if s == "" {
		return s
	}
	needsQuotes := false
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}
	if !needsQuotes {
		return s
	}
	result := "\""
	for _, c := range s {
		if c == '"' {
			result += "\"\""
		} else {
			result += string(c)
		}
	}
	result += "\""
	return result
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, model.ErrorResponse{Error: msg})
}

// paginate applies ?page= and ?limit= to any slice using generics.
func paginate[T any](r *http.Request, items []T) model.PaginatedResponse {
	page := queryInt(r, "page", 1)
	limit := queryInt(r, "limit", 50)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 500 {
		limit = 50
	}

	total := len(items)
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		items = []T{}
	} else {
		if end > total {
			end = total
		}
		items = items[start:end]
	}

	return model.PaginatedResponse{
		Data:  items,
		Total: total,
		Page:  page,
		Limit: limit,
	}
}

func queryInt(r *http.Request, key string, def int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
