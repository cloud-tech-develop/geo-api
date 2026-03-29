package repository

import (
	"os"
	"testing"

	"github.com/cloud-tech-develop/geo-api/internal/model"
)

func setupTestRepo(t *testing.T) *GeoRepository {
	tmpFile, err := os.CreateTemp("", "test-geo-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	metadataFile, err := os.CreateTemp("", "test-meta-*.json")
	if err != nil {
		t.Fatalf("Failed to create metadata temp file: %v", err)
	}
	defer os.Remove(metadataFile.Name())

	testData := `[
		{
			"id": 1,
			"name": "Colombia",
			"iso2": "CO",
			"iso3": "COL",
			"phone_code": "57",
			"capital": "Bogotá",
			"currency": "COP",
			"region": "Americas",
			"subregion": "South America",
			"emoji": "🇨🇴",
			"states": [
				{
					"id": 100,
					"name": "Antioquia",
					"country_id": 1,
					"country_code": "CO",
					"state_code": "ANT",
					"latitude": "6.2442",
					"longitude": "-75.5812",
					"cities": [
						{
							"id": 1000,
							"name": "Medellín",
							"state_id": 100,
							"state_code": "ANT",
							"country_id": 1,
							"country_code": "CO",
							"latitude": "6.2442",
							"longitude": "-75.5812"
						},
						{
							"id": 1001,
							"name": "Bello",
							"state_id": 100,
							"state_code": "ANT",
							"country_id": 1,
							"country_code": "CO",
							"latitude": "6.3374",
							"longitude": "-75.5579"
						}
					]
				}
			]
		},
		{
			"id": 2,
			"name": "Argentina",
			"iso2": "AR",
			"iso3": "ARG",
			"phone_code": "54",
			"capital": "Buenos Aires",
			"currency": "ARS",
			"region": "Americas",
			"subregion": "South America",
			"emoji": "🇦🇷",
			"states": []
		}
	]`

	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	tmpFile.Close()

	repo, err := New(tmpFile.Name(), metadataFile.Name())
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	return repo
}

func TestNew(t *testing.T) {
	repo := setupTestRepo(t)
	if repo == nil {
		t.Fatal("Expected repo to be created")
	}
}

func TestGetCountries(t *testing.T) {
	repo := setupTestRepo(t)

	tests := []struct {
		name     string
		search   string
		expected int
	}{
		{"all countries", "", 2},
		{"search colombia", "col", 1},
		{"search argentina", "arg", 1},
		{"search medellin", "", 2},
		{"no results", "xyz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := repo.GetCountries(tt.search)
			if len(result) != tt.expected {
				t.Errorf("GetCountries(%q) = %d, want %d", tt.search, len(result), tt.expected)
			}
		})
	}
}

func TestGetCountryByISO2(t *testing.T) {
	repo := setupTestRepo(t)

	tests := []struct {
		iso2     string
		expected string
		found    bool
	}{
		{"CO", "Colombia", true},
		{"AR", "Argentina", true},
		{"co", "Colombia", true},
		{"US", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.iso2, func(t *testing.T) {
			result, ok := repo.GetCountryByISO2(tt.iso2)
			if ok != tt.found {
				t.Errorf("GetCountryByISO2(%q) ok = %v, want %v", tt.iso2, ok, tt.found)
			}
			if ok && result.Name != tt.expected {
				t.Errorf("GetCountryByISO2(%q).Name = %q, want %q", tt.iso2, result.Name, tt.expected)
			}
		})
	}
}

func TestGetStatesByCountry(t *testing.T) {
	repo := setupTestRepo(t)

	tests := []struct {
		iso2     string
		search   string
		expected int
		found    bool
	}{
		{"CO", "", 1, true},
		{"CO", "ant", 1, true},
		{"CO", "xyz", 0, true},
		{"AR", "", 0, true},
		{"XX", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.iso2+"-"+tt.search, func(t *testing.T) {
			result, ok := repo.GetStatesByCountry(tt.iso2, tt.search)
			if ok != tt.found {
				t.Errorf("GetStatesByCountry(%q, %q) ok = %v, want %v", tt.iso2, tt.search, ok, tt.found)
			}
			if ok && len(result) != tt.expected {
				t.Errorf("GetStatesByCountry(%q, %q) = %d states, want %d", tt.iso2, tt.search, len(result), tt.expected)
			}
		})
	}
}

func TestGetCitiesByState(t *testing.T) {
	repo := setupTestRepo(t)

	tests := []struct {
		stateID  int
		search   string
		expected int
		found    bool
	}{
		{100, "", 2, true},
		{100, "med", 1, true},
		{100, "bello", 1, true},
		{100, "xyz", 0, true},
		{999, "", 0, false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, ok := repo.GetCitiesByState(tt.stateID, tt.search)
			if ok != tt.found {
				t.Errorf("GetCitiesByState(%d, %q) ok = %v, want %v", tt.stateID, tt.search, ok, tt.found)
			}
			if ok && len(result) != tt.expected {
				t.Errorf("GetCitiesByState(%d, %q) = %d cities, want %d", tt.stateID, tt.search, len(result), tt.expected)
			}
		})
	}
}

func TestGetCityByID(t *testing.T) {
	repo := setupTestRepo(t)

	tests := []struct {
		id       int
		expected string
		found    bool
	}{
		{1000, "Medellín", true},
		{1001, "Bello", true},
		{9999, "", false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result, ok := repo.GetCityByID(tt.id)
			if ok != tt.found {
				t.Errorf("GetCityByID(%d) ok = %v, want %v", tt.id, ok, tt.found)
			}
			if ok && result.Name != tt.expected {
				t.Errorf("GetCityByID(%d).Name = %q, want %q", tt.id, result.Name, tt.expected)
			}
		})
	}
}

func TestUpdateCityMetadata(t *testing.T) {
	repo := setupTestRepo(t)

	metadata := map[string]any{"population": 2500000.0, "timezone": "America/Bogota"}
	err := repo.UpdateCityMetadata(1000, metadata)
	if err != nil {
		t.Fatalf("UpdateCityMetadata failed: %v", err)
	}

	city, ok := repo.GetCityByID(1000)
	if !ok {
		t.Fatal("City not found after update")
	}
	if city.Metadata["population"] != 2500000.0 {
		t.Errorf("Metadata population = %v, want 2500000", city.Metadata["population"])
	}
	if city.Metadata["timezone"] != "America/Bogota" {
		t.Errorf("Metadata timezone = %v, want America/Bogota", city.Metadata["timezone"])
	}
}

func TestUpdateCityMetadata_Merge(t *testing.T) {
	repo := setupTestRepo(t)

	err := repo.UpdateCityMetadata(1000, map[string]any{"population": 2500000.0})
	if err != nil {
		t.Fatalf("First update failed: %v", err)
	}

	err = repo.UpdateCityMetadata(1000, map[string]any{"timezone": "America/Bogota"})
	if err != nil {
		t.Fatalf("Second update failed: %v", err)
	}

	city, _ := repo.GetCityByID(1000)
	if city.Metadata["population"] != 2500000.0 {
		t.Errorf("Merged metadata should preserve population")
	}
	if city.Metadata["timezone"] != "America/Bogota" {
		t.Errorf("Merged metadata should have timezone")
	}
}

func TestBulkUpdateMetadata(t *testing.T) {
	repo := setupTestRepo(t)

	updates := []model.BulkMetadataUpdate{
		{ID: 1000, Metadata: map[string]any{"population": 2500000.0}},
		{ID: 1001, Metadata: map[string]any{"population": 400000.0}},
	}

	result := repo.BulkUpdateMetadata(updates)
	if result.Updated != 2 {
		t.Errorf("BulkUpdate updated %d, want 2", result.Updated)
	}
	if result.Failed != 0 {
		t.Errorf("BulkUpdate failed %d, want 0", result.Failed)
	}

	city1, _ := repo.GetCityByID(1000)
	city2, _ := repo.GetCityByID(1001)
	if city1.Metadata["population"] != 2500000.0 {
		t.Errorf("City 1000 population = %v", city1.Metadata["population"])
	}
	if city2.Metadata["population"] != 400000.0 {
		t.Errorf("City 1001 population = %v", city2.Metadata["population"])
	}
}

func TestBulkUpdateMetadata_FailedCities(t *testing.T) {
	repo := setupTestRepo(t)

	updates := []model.BulkMetadataUpdate{
		{ID: 1000, Metadata: map[string]any{"population": 2500000.0}},
		{ID: 9999, Metadata: map[string]any{"population": 100.0}},
	}

	result := repo.BulkUpdateMetadata(updates)
	if result.Updated != 1 {
		t.Errorf("BulkUpdate updated %d, want 1", result.Updated)
	}
	if result.Failed != 1 {
		t.Errorf("BulkUpdate failed %d, want 1", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}

func TestGetStats(t *testing.T) {
	repo := setupTestRepo(t)

	stats := repo.GetStats()
	if stats.Countries != 2 {
		t.Errorf("Countries = %d, want 2", stats.Countries)
	}
	if stats.States != 1 {
		t.Errorf("States = %d, want 1", stats.States)
	}
	if stats.Cities != 2 {
		t.Errorf("Cities = %d, want 2", stats.Cities)
	}
}
