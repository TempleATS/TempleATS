package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type queryRequest struct {
	SQL string `json:"sql"`
}

type queryResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
}

func (s *Server) RunQuery(w http.ResponseWriter, r *http.Request) {
	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	sql := strings.TrimSpace(req.SQL)
	if sql == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sql is required"})
		return
	}

	// Only allow SELECT queries
	upper := strings.ToUpper(sql)
	if !strings.HasPrefix(upper, "SELECT") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "only SELECT queries are allowed"})
		return
	}

	// Block dangerous keywords
	for _, kw := range []string{"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "CREATE", "GRANT", "REVOKE"} {
		// Check for keyword as a separate word (preceded by space/semicolon/start)
		if strings.Contains(upper, kw) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query contains disallowed keyword: " + kw})
			return
		}
	}

	rows, err := s.Pool.Query(r.Context(), sql)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Get column names
	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = string(f.Name)
	}

	// Collect rows (limit to 500)
	var result [][]interface{}
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		result = append(result, vals)
		if len(result) >= 500 {
			break
		}
	}

	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, queryResponse{
		Columns: columns,
		Rows:    result,
		Count:   len(result),
	})
}
