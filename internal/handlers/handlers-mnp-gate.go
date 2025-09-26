package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// MnpGateConfig represents a single row from mnp_gate_config table.
type MnpGateConfig struct {
	ID         int                    `json:"id"`
	Engine     sql.NullString         `json:"engine"`
	MaxWorkers sql.NullInt64          `json:"max_workers"`
	CacheDays  sql.NullInt64          `json:"cache_days"`
	ConfigRaw  sql.NullString         `json:"-"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

func MnpGateHandler(db *sql.DB, w http.ResponseWriter, r *http.Request) {
	selectQuery := "SELECT id, engine, max_workers, cache_days, config FROM mnp_gate_config"
	rows, err := db.Query(selectQuery)
	if err != nil {
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []MnpGateConfig

	for rows.Next() {
		var rec MnpGateConfig
		if err := rows.Scan(&rec.ID, &rec.Engine, &rec.MaxWorkers, &rec.CacheDays, &rec.ConfigRaw); err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Try to parse JSON config if present
		if rec.ConfigRaw.Valid && rec.ConfigRaw.String != "" {
			var cfg map[string]interface{}
			if err := json.Unmarshal([]byte(rec.ConfigRaw.String), &cfg); err == nil {
				rec.Config = cfg
			} else {
				// If JSON parsing fails, include raw string under a key
				rec.Config = map[string]interface{}{"_raw": rec.ConfigRaw.String}
			}
		}

		results = append(results, rec)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		http.Error(w, "JSON encode error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
