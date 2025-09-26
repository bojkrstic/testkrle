package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// NewHomeHandler returns an http.HandlerFunc that uses the provided db and tmpl.
func NewHomeHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
	renderTemplate := func(w http.ResponseWriter, tmplName string, data interface{}) {
		err := tmpl.ExecuteTemplate(w, tmplName, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Prepare data structure for template (now showing mnp_gate_config)
		type GateConfig struct {
			ID         int
			Engine     string
			MaxWorkers int
			CacheDays  int
			Config     string
		}

		type Filters struct {
			ID     string
			Engine string
		}

		type PageData struct {
			Version         string
			TaxRates        []GateConfig // reusing template field name TaxRates for compatibility
			Page            int
			PageSize        int
			Total           int
			TotalPages      int
			PrevPage        int
			NextPage        int
			Sort            string
			Dir             string
			Filters         Filters
			BaseQueryPrefix string
		}

		var data PageData

		// default pagination
		page := 1
		pageSize := 10
		if p := r.URL.Query().Get("page"); p != "" {
			if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
				page = pi
			}
		}

		data.Page = page
		data.PageSize = pageSize

		// parse filters and sort params
		q := r.URL.Query()
		sortParam := q.Get("sort")
		dirParam := q.Get("dir")
		// whitelist sort columns for mnp_gate_config
		allowedSort := map[string]bool{"id": true, "engine": true, "max_workers": true, "cache_days": true}
		if !allowedSort[sortParam] {
			sortParam = "id"
		}
		if dirParam != "asc" && dirParam != "desc" {
			dirParam = "asc"
		}
		data.Sort = sortParam
		data.Dir = dirParam

		// filters
		f := Filters{
			ID:     q.Get("id"),
			Engine: q.Get("engine"),
		}
		data.Filters = f

		// build base query prefix (preserve filters & sort, but not page)
		vals := q
		vals.Del("page")
		base := vals.Encode()
		if base != "" {
			data.BaseQueryPrefix = "?" + base + "&"
		} else {
			data.BaseQueryPrefix = "?"
		}

		// Get DB version
		if err := db.QueryRow("SELECT VERSION()").Scan(&data.Version); err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// build WHERE clauses based on filters for mnp_gate_config
		where := ""
		var whereArgs []interface{}
		clauses := make([]string, 0)
		if f.ID != "" {
			clauses = append(clauses, "id = ?")
			whereArgs = append(whereArgs, f.ID)
		}
		if f.Engine != "" {
			clauses = append(clauses, "engine LIKE ?")
			whereArgs = append(whereArgs, "%"+f.Engine+"%")
		}
		if len(clauses) > 0 {
			where = " WHERE " + strings.Join(clauses, " AND ")
		}

		// Count total rows for pagination
		var total int
		countQuery := "SELECT COUNT(*) FROM mnp_gate_config" + where
		if err := db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
			// If table is missing, treat as empty result set instead of error
			if strings.Contains(err.Error(), "1146") || strings.Contains(err.Error(), "doesn't exist") {
				data.Total = 0
				data.TotalPages = 0
				data.PrevPage = 0
				data.NextPage = 0
				renderTemplate(w, "home.html", data)
				return
			}
			http.Error(w, "Count query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data.Total = total
		if total == 0 {
			data.TotalPages = 0
			data.PrevPage = 0
			data.NextPage = 0
			renderTemplate(w, "home.html", data)
			return
		}

		// compute total pages
		totalPages := (total + pageSize - 1) / pageSize
		data.TotalPages = totalPages
		if page > totalPages {
			page = totalPages
			data.Page = page
		}
		if page > 1 {
			data.PrevPage = page - 1
		}
		if page < totalPages {
			data.NextPage = page + 1
		}

		offset := (page - 1) * pageSize

		// Query paginated mnp gate configs with filters and sort
		selectQuery := fmt.Sprintf("SELECT id,engine,max_workers,cache_days,config FROM mnp_gate_config%s ORDER BY %s %s LIMIT ? OFFSET ?", where, sortParam, dirParam)
		args := append(whereArgs, pageSize, offset)
		rows, err := db.Query(selectQuery, args...)
		if err != nil {
			// If table is missing at select time, just render empty results
			if strings.Contains(err.Error(), "1146") || strings.Contains(err.Error(), "doesn't exist") {
				data.TaxRates = nil
				renderTemplate(w, "home.html", data)
				return
			}
			http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var gc GateConfig
			if err := rows.Scan(&gc.ID, &gc.Engine, &gc.MaxWorkers, &gc.CacheDays, &gc.Config); err != nil {
				http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			data.TaxRates = append(data.TaxRates, gc)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Render template with data
		renderTemplate(w, "home.html", data)
	}
}
