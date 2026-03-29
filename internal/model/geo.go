package model

// ─── Raw JSON shapes (matches countries+states+cities.json) ───────────────────

type CityRaw struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	StateID     int            `json:"state_id"`
	StateCode   string         `json:"state_code"`
	CountryID   int            `json:"country_id"`
	CountryCode string         `json:"country_code"`
	Latitude    string         `json:"latitude"`
	Longitude   string         `json:"longitude"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type StateRaw struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CountryID   int       `json:"country_id"`
	CountryCode string    `json:"country_code"`
	StateCode   string    `json:"state_code"`
	Latitude    string    `json:"latitude"`
	Longitude   string    `json:"longitude"`
	Cities      []CityRaw `json:"cities"`
}

type CountryRaw struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	ISO2        string     `json:"iso2"`
	ISO3        string     `json:"iso3"`
	PhoneCode   string     `json:"phone_code"`
	Capital     string     `json:"capital"`
	Currency    string     `json:"currency"`
	CurrencyName string    `json:"currency_name"`
	Region      string     `json:"region"`
	Subregion   string     `json:"subregion"`
	Latitude    string     `json:"latitude"`
	Longitude   string     `json:"longitude"`
	Emoji       string     `json:"emoji"`
	States      []StateRaw `json:"states"`
}

// ─── API response shapes ───────────────────────────────────────────────────────

type CountrySummary struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ISO2      string `json:"iso2"`
	ISO3      string `json:"iso3"`
	PhoneCode string `json:"phone_code"`
	Capital   string `json:"capital"`
	Currency  string `json:"currency"`
	Region    string `json:"region"`
	Subregion string `json:"subregion"`
	Emoji     string `json:"emoji"`
}

type StateSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	StateCode   string `json:"state_code"`
	CountryID   int    `json:"country_id"`
	CountryCode string `json:"country_code"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
}

type CitySummary struct {
	ID          int            `json:"id"`
	Name        string         `json:"name"`
	StateID     int            `json:"state_id"`
	StateCode   string         `json:"state_code"`
	CountryID   int            `json:"country_id"`
	CountryCode string         `json:"country_code"`
	Latitude    string         `json:"latitude"`
	Longitude   string         `json:"longitude"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type PaginatedResponse struct {
	Data  any `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
