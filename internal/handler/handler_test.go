package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("Health response status = %q, want %q", resp["status"], "ok")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSON(w, http.StatusOK, map[string]string{"key": "value"})

	if w.Code != http.StatusOK {
		t.Errorf("writeJSON status = %d, want %d", w.Code, http.StatusOK)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want %q", w.Header().Get("Content-Type"), "application/json")
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusNotFound, "not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("writeError status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "not found" {
		t.Errorf("writeError body error = %q, want %q", resp["error"], "not found")
	}
}

func TestPaginate(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}

	tests := []struct {
		page      string
		limit     string
		wantTotal int
		wantLen   int
	}{
		{"1", "2", 5, 2},
		{"2", "2", 5, 2},
		{"3", "2", 5, 1},
		{"1", "10", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.page+"-"+tt.limit, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?page="+tt.page+"&limit="+tt.limit, nil)
			resp := paginate(req, items)

			if resp.Total != tt.wantTotal {
				t.Errorf("paginate Total = %d, want %d", resp.Total, tt.wantTotal)
			}
			if len(resp.Data.([]string)) != tt.wantLen {
				t.Errorf("paginate len(data) = %d, want %d", len(resp.Data.([]string)), tt.wantLen)
			}
		})
	}
}

func TestPaginate_EmptyResult(t *testing.T) {
	items := []string{}

	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=10", nil)
	resp := paginate(req, items)

	if resp.Total != 0 {
		t.Errorf("paginate Total = %d, want 0", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("paginate Page = %d, want 1", resp.Page)
	}
}

func TestQueryInt(t *testing.T) {
	tests := []struct {
		key      string
		value    string
		def      int
		expected int
	}{
		{"page", "5", 1, 5},
		{"page", "", 1, 1},
		{"page", "abc", 1, 1},
		{"limit", "100", 50, 100},
		{"limit", "", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.key+"-"+tt.value, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/?"+tt.key+"="+tt.value, nil)
			result := queryInt(req, tt.key, tt.def)
			if result != tt.expected {
				t.Errorf("queryInt(%q, %q, %d) = %d, want %d", tt.key, tt.value, tt.def, result, tt.expected)
			}
		})
	}
}

func TestEscapeCSV(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with,comma", "\"with,comma\""},
		{"with\"quote", "\"with\"\"quote\""},
		{"with\nnewline", "\"with\nnewline\""},
		{"", ""},
		{"multiple,commas,here", "\"multiple,commas,here\""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeCSV(tt.input)
			if result != tt.expected {
				t.Errorf("escapeCSV(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
