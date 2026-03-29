package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/youruser/geo-api/internal/model"
)

// GeoRepository holds all geo data in memory and exposes query methods.
type GeoRepository struct {
	countries    []model.CountryRaw
	metadataPath string
	mu           sync.RWMutex

	// fast-lookup indexes
	countryByISO2 map[string]*model.CountryRaw // key: uppercase ISO2
	countryByID   map[int]*model.CountryRaw
	stateByID     map[int]*model.StateRaw
	cityByID      map[int]*model.CityRaw
}

func (r *GeoRepository) GetMetadataPath() string {
	return r.metadataPath
}

// New loads the dataset from the given JSON file path and builds indexes.
func New(dataPath, metadataPath string) (*GeoRepository, error) {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, fmt.Errorf("reading geo data: %w", err)
	}

	var countries []model.CountryRaw
	if err := json.Unmarshal(data, &countries); err != nil {
		return nil, fmt.Errorf("parsing geo data: %w", err)
	}

	repo := &GeoRepository{
		countries:     countries,
		metadataPath:  metadataPath,
		countryByISO2: make(map[string]*model.CountryRaw),
		countryByID:   make(map[int]*model.CountryRaw),
		stateByID:     make(map[int]*model.StateRaw),
		cityByID:      make(map[int]*model.CityRaw),
	}

	// Build indexes
	for i := range repo.countries {
		c := &repo.countries[i]
		repo.countryByISO2[strings.ToUpper(c.ISO2)] = c
		repo.countryByID[c.ID] = c
		for j := range c.States {
			s := &c.States[j]
			repo.stateByID[s.ID] = s

			// Fill state with country info if missing
			if s.CountryCode == "" {
				s.CountryCode = c.ISO2
			}
			if s.CountryID == 0 {
				s.CountryID = c.ID
			}

			for k := range s.Cities {
				city := &s.Cities[k]

				// CRITICAL: Propagate parent info to the city
				// (Often missing in nested JSON structures)
				if city.CountryCode == "" {
					city.CountryCode = c.ISO2
				}
				if city.CountryID == 0 {
					city.CountryID = c.ID
				}
				if city.StateID == 0 {
					city.StateID = s.ID
				}
				if city.StateCode == "" {
					city.StateCode = s.StateCode
				}

				repo.cityByID[city.ID] = city
			}
		}
	}

	// Load existing metadata
	if err := repo.loadMetadata(); err != nil {
		log.Printf("Warning: could not load metadata: %v", err)
	}

	return repo, nil
}

func (r *GeoRepository) loadMetadata() error {
	data, err := os.ReadFile(r.metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metadata map[int]map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for id, m := range metadata {
		if city, ok := r.cityByID[id]; ok {
			city.Metadata = m
		}
	}
	return nil
}

func (r *GeoRepository) saveMetadata() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.saveMetadataLocked()
}

// UpdateCityMetadata updates metadata for a specific city (merges with existing).
func (r *GeoRepository) UpdateCityMetadata(id int, metadata map[string]any) error {
	r.mu.Lock()
	city, ok := r.cityByID[id]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("city with ID %d not found", id)
	}
	if city.Metadata == nil {
		city.Metadata = make(map[string]any)
	}
	for k, v := range metadata {
		city.Metadata[k] = v
	}
	r.mu.Unlock()

	return r.saveMetadata()
}

type BulkUpdateResult struct {
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	Errors  []struct {
		ID    int    `json:"id"`
		Error string `json:"error"`
	} `json:"errors,omitempty"`
}

func (r *GeoRepository) BulkUpdateMetadata(updates []model.BulkMetadataUpdate) BulkUpdateResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := BulkUpdateResult{}
	for _, update := range updates {
		city, ok := r.cityByID[update.ID]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, struct {
				ID    int    `json:"id"`
				Error string `json:"error"`
			}{update.ID, "city not found"})
			continue
		}
		if city.Metadata == nil {
			city.Metadata = make(map[string]any)
		}
		for k, v := range update.Metadata {
			city.Metadata[k] = v
		}
		result.Updated++
	}

	if err := r.saveMetadataLocked(); err != nil {
		result.Errors = append(result.Errors, struct {
			ID    int    `json:"id"`
			Error string `json:"error"`
		}{0, fmt.Sprintf("failed to save: %v", err)})
	}

	return result
}

func (r *GeoRepository) saveMetadataLocked() error {
	metadata := make(map[int]map[string]any)
	for id, city := range r.cityByID {
		if len(city.Metadata) > 0 {
			metadata[id] = city.Metadata
		}
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.metadataPath, data, 0644)
}

// ─── Countries ────────────────────────────────────────────────────────────────

func (r *GeoRepository) GetCountries(search string) []model.CountrySummary {
	search = strings.ToLower(search)
	result := make([]model.CountrySummary, 0, len(r.countries))
	for _, c := range r.countries {
		if search != "" && !strings.Contains(strings.ToLower(c.Name), search) {
			continue
		}
		result = append(result, toCountrySummary(c))
	}
	return result
}

func (r *GeoRepository) GetCountryByISO2(iso2 string) (*model.CountrySummary, bool) {
	c, ok := r.countryByISO2[strings.ToUpper(iso2)]
	if !ok {
		return nil, false
	}
	s := toCountrySummary(*c)
	return &s, true
}

// ─── States ───────────────────────────────────────────────────────────────────

func (r *GeoRepository) GetStatesByCountry(iso2, search string) ([]model.StateSummary, bool) {
	c, ok := r.countryByISO2[strings.ToUpper(iso2)]
	if !ok {
		return nil, false
	}
	search = strings.ToLower(search)
	result := make([]model.StateSummary, 0, len(c.States))
	for _, s := range c.States {
		if search != "" && !strings.Contains(strings.ToLower(s.Name), search) {
			continue
		}
		result = append(result, toStateSummary(s))
	}
	return result, true
}

func (r *GeoRepository) GetStateByID(id int) (*model.StateSummary, bool) {
	s, ok := r.stateByID[id]
	if !ok {
		return nil, false
	}
	ss := toStateSummary(*s)
	return &ss, true
}

// ─── Cities ───────────────────────────────────────────────────────────────────

func (r *GeoRepository) GetAllCities(search, country string, state int, page, limit int) ([]model.CitySummary, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	search = strings.ToLower(search)
	country = strings.ToUpper(country)
	var filtered []model.CitySummary

	for _, city := range r.cityByID {
		// Filter by search name
		if search != "" && !strings.Contains(strings.ToLower(city.Name), search) {
			continue
		}
		// Filter by country code (case-insensitive)
		if country != "" && strings.ToUpper(city.CountryCode) != country {
			continue
		}
		// Filter by state ID
		if state > 0 && city.StateID != state {
			continue
		}

		filtered = append(filtered, toCitySummary(*city))
	}

	total := len(filtered)

	// Sort by name, then by ID to ensure stable pagination
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Name != filtered[j].Name {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].ID < filtered[j].ID
	})

	// Simple pagination
	start := (page - 1) * limit
	if start < 0 {
		start = 0
	}
	if start >= total {
		return []model.CitySummary{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return filtered[start:end], total
}

func (r *GeoRepository) GetCitiesByState(stateID int, search string) ([]model.CitySummary, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.stateByID[stateID]
	if !ok {
		return nil, false
	}
	search = strings.ToLower(search)
	result := make([]model.CitySummary, 0, len(s.Cities))
	for _, city := range s.Cities {
		if search != "" && !strings.Contains(strings.ToLower(city.Name), search) {
			continue
		}
		result = append(result, toCitySummary(city))
	}
	return result, true
}

func (r *GeoRepository) GetCitiesByCountry(iso2, search string) ([]model.CitySummary, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.countryByISO2[strings.ToUpper(iso2)]
	if !ok {
		return nil, false
	}
	search = strings.ToLower(search)
	var result []model.CitySummary
	for _, s := range c.States {
		for _, city := range s.Cities {
			if search != "" && !strings.Contains(strings.ToLower(city.Name), search) {
				continue
			}
			result = append(result, toCitySummary(city))
		}
	}
	return result, true
}

func (r *GeoRepository) GetCityByID(id int) (*model.CitySummary, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	city, ok := r.cityByID[id]
	if !ok {
		return nil, false
	}
	s := toCitySummary(*city)
	return &s, true
}

// ─── Stats ────────────────────────────────────────────────────────────────────

type Stats struct {
	Countries int `json:"countries"`
	States    int `json:"states"`
	Cities    int `json:"cities"`
}

func (r *GeoRepository) GetStats() Stats {
	var states, cities int
	for _, c := range r.countries {
		states += len(c.States)
		for _, s := range c.States {
			cities += len(s.Cities)
		}
	}
	return Stats{
		Countries: len(r.countries),
		States:    states,
		Cities:    cities,
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func toCountrySummary(c model.CountryRaw) model.CountrySummary {
	return model.CountrySummary{
		ID: c.ID, Name: c.Name, ISO2: c.ISO2, ISO3: c.ISO3,
		PhoneCode: c.PhoneCode, Capital: c.Capital, Currency: c.Currency,
		Region: c.Region, Subregion: c.Subregion, Emoji: c.Emoji,
	}
}

func toStateSummary(s model.StateRaw) model.StateSummary {
	return model.StateSummary{
		ID: s.ID, Name: s.Name, StateCode: s.StateCode,
		CountryID: s.CountryID, CountryCode: s.CountryCode,
		Latitude: s.Latitude, Longitude: s.Longitude,
	}
}

func toCitySummary(c model.CityRaw) model.CitySummary {
	return model.CitySummary{
		ID: c.ID, Name: c.Name, StateID: c.StateID, StateCode: c.StateCode,
		CountryID: c.CountryID, CountryCode: c.CountryCode,
		Latitude: c.Latitude, Longitude: c.Longitude,
		Metadata: c.Metadata,
	}
}
