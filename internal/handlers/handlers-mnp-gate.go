package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
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

// NewMnpGatePageHandler returns an http.HandlerFunc that renders the mnp_gate page.
// MnpGatePageHandler is an http.Handler that renders the mnp_gate page.
type MnpGatePageHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewMnpGatePageHandler(db *sql.DB, tmpl *template.Template) http.Handler {
	return &MnpGatePageHandler{DB: db, Tmpl: tmpl}
}

func (h *MnpGatePageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query("SELECT id, engine, max_workers, cache_days, config FROM mnp_gate_config")
	if err != nil {
		// If table missing, render empty page
		if rows == nil {
			h.Tmpl.ExecuteTemplate(w, "mnp_gate", map[string]interface{}{"Configs": []interface{}{}})
			return
		}
		http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type cfg struct {
		ID         int
		Engine     string
		MaxWorkers int
		CacheDays  int
		Config     string
	}

	var configs []cfg
	for rows.Next() {
		var id int
		var engine sql.NullString
		var maxWorkers sql.NullInt64
		var cacheDays sql.NullInt64
		var config sql.NullString
		if err := rows.Scan(&id, &engine, &maxWorkers, &cacheDays, &config); err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		configs = append(configs, cfg{
			ID:         id,
			Engine:     engine.String,
			MaxWorkers: int(maxWorkers.Int64),
			CacheDays:  int(cacheDays.Int64),
			Config:     config.String,
		})
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Render template
	data := map[string]interface{}{"Configs": configs}
	if err := h.Tmpl.ExecuteTemplate(w, "mnp_gate", data); err != nil {
		http.Error(w, "Template render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewMnpGatesListHandler renders the list of rows from `mnp_gate` table (full schema view).
// MnpGatesListHandler is an http.Handler that renders the list of rows from `mnp_gate` table.
type MnpGatesListHandler struct {
	DB   *sql.DB
	Tmpl *template.Template
}

func NewMnpGatesListHandler(db *sql.DB, tmpl *template.Template) http.Handler {
	return &MnpGatesListHandler{DB: db, Tmpl: tmpl}
}

func (h *MnpGatesListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, instance_id, group_id, supplier_id, name, code_name, engine_id, throughput_queries, connection, billing_account_id, price_list_id, type, linked_mnp_account_id, status, insert_dt, status_dt, setup_date FROM mnp_gate`
	rows, err := h.DB.Query(query)
	if err != nil {
		// If table missing, render empty list
		h.Tmpl.ExecuteTemplate(w, "mnp_gates", map[string]interface{}{"Gates": []interface{}{}})
		return
	}
	defer rows.Close()

	type gate struct {
		ID                 int
		InstanceID         int
		GroupID            int
		SupplierID         int
		Name               string
		CodeName           string
		EngineID           int
		ThroughputQueries  int
		Connection         string
		BillingAccountID   int
		PriceListID        int
		Type               string
		LinkedMnpAccountID int
		Status             string
		InsertDT           string
		StatusDT           string
		SetupDate          string
	}

	var gates []gate
	for rows.Next() {
		var (
			id          int
			instanceID  sql.NullInt64
			groupID     sql.NullInt64
			supplierID  sql.NullInt64
			name        sql.NullString
			codeName    sql.NullString
			engineID    sql.NullInt64
			throughput  sql.NullInt64
			connection  sql.NullString
			billingID   sql.NullInt64
			priceListID sql.NullInt64
			typ         sql.NullString
			linkedID    sql.NullInt64
			status      sql.NullString
			insertDT    sql.NullString
			statusDT    sql.NullString
			setupDate   sql.NullString
		)
		if err := rows.Scan(&id, &instanceID, &groupID, &supplierID, &name, &codeName, &engineID, &throughput, &connection, &billingID, &priceListID, &typ, &linkedID, &status, &insertDT, &statusDT, &setupDate); err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		g := gate{
			ID:                 id,
			InstanceID:         int(instanceID.Int64),
			GroupID:            int(groupID.Int64),
			SupplierID:         int(supplierID.Int64),
			Name:               name.String,
			CodeName:           codeName.String,
			EngineID:           int(engineID.Int64),
			ThroughputQueries:  int(throughput.Int64),
			Connection:         connection.String,
			BillingAccountID:   int(billingID.Int64),
			PriceListID:        int(priceListID.Int64),
			Type:               typ.String,
			LinkedMnpAccountID: int(linkedID.Int64),
			Status:             status.String,
			InsertDT:           insertDT.String,
			StatusDT:           statusDT.String,
			SetupDate:          setupDate.String,
		}
		gates = append(gates, g)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{"Gates": gates}
	if err := h.Tmpl.ExecuteTemplate(w, "mnp_gates", data); err != nil {
		http.Error(w, "Template render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
