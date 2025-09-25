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
		// Prepare data structure for template
		type TaxRate struct {
			ID          int
			TaxCategory int
			StartDate   string
			EndDate     string
			RatePercent float64
		}
		type Filters struct {
			ID          string
			TaxCategory string
			StartDate   string
			EndDate     string
			RatePercent string
		}

		type PageData struct {
			Version         string
			TaxRates        []TaxRate
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
		// whitelist sort columns
		allowedSort := map[string]bool{"id": true, "tax_category_id": true, "start_date": true, "end_date": true, "rate_percent": true}
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
			ID:          q.Get("id"),
			TaxCategory: q.Get("tax_category_id"),
			StartDate:   q.Get("start_date"),
			EndDate:     q.Get("end_date"),
			RatePercent: q.Get("rate_percent"),
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

		// build WHERE clauses based on filters
		where := ""
		var whereArgs []interface{}
		clauses := make([]string, 0)
		if f.ID != "" {
			clauses = append(clauses, "id = ?")
			whereArgs = append(whereArgs, f.ID)
		}
		if f.TaxCategory != "" {
			clauses = append(clauses, "tax_category_id = ?")
			whereArgs = append(whereArgs, f.TaxCategory)
		}
		if f.StartDate != "" {
			clauses = append(clauses, "start_date LIKE ?")
			whereArgs = append(whereArgs, "%"+f.StartDate+"%")
		}
		if f.EndDate != "" {
			clauses = append(clauses, "end_date LIKE ?")
			whereArgs = append(whereArgs, "%"+f.EndDate+"%")
		}
		if f.RatePercent != "" {
			clauses = append(clauses, "rate_percent = ?")
			whereArgs = append(whereArgs, f.RatePercent)
		}
		if len(clauses) > 0 {
			where = " WHERE " + strings.Join(clauses, " AND ")
		}

		// Count total rows for pagination
		var total int
		countQuery := "SELECT COUNT(*) FROM sys_tax_rate" + where
		if err := db.QueryRow(countQuery, whereArgs...).Scan(&total); err != nil {
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

		// Query paginated tax rates with filters and sort
		selectQuery := fmt.Sprintf("SELECT id,tax_category_id,start_date,end_date,rate_percent FROM sys_tax_rate%s ORDER BY %s %s LIMIT ? OFFSET ?", where, sortParam, dirParam)
		args := append(whereArgs, pageSize, offset)
		rows, err := db.Query(selectQuery, args...)
		if err != nil {
			http.Error(w, "Database query error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tr TaxRate
			if err := rows.Scan(&tr.ID, &tr.TaxCategory, &tr.StartDate, &tr.EndDate, &tr.RatePercent); err != nil {
				http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			data.TaxRates = append(data.TaxRates, tr)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Rows error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Render template with data
		renderTemplate(w, "home.html", data)
	}
}
