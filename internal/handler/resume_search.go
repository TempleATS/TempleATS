package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	mw "github.com/temple-ats/TempleATS/internal/middleware"
)

type resumeSearchRequest struct {
	Q string `json:"q"`
}

type resumeSearchResult struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Company        string `json:"company"`
	ResumeFilename string `json:"resume_filename"`
	Snippet        string `json:"snippet"`
}

// SearchResumes handles boolean search queries against resume_text.
// Supports: AND, OR, NOT, parentheses, and quoted phrases.
// Examples: "python AND machine learning", "react OR vue", "senior NOT junior"
func (s *Server) SearchResumes(w http.ResponseWriter, r *http.Request) {
	var req resumeSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	q := strings.TrimSpace(req.Q)
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "search query is required"})
		return
	}

	orgID := mw.GetOrgID(r.Context())

	// Parse boolean expression into SQL WHERE clause
	where, args, err := parseBooleanQuery(q)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sql := fmt.Sprintf(`SELECT id, name, email, COALESCE(company, '') AS company,
		COALESCE(resume_filename, '') AS resume_filename,
		COALESCE(LEFT(resume_text, 300), '') AS snippet
		FROM candidates
		WHERE organization_id = $1 AND resume_text IS NOT NULL AND (%s)
		ORDER BY name ASC
		LIMIT 100`, where)

	// Prepend orgID as $1
	allArgs := append([]interface{}{orgID}, args...)

	rows, err := s.Pool.Query(r.Context(), sql, allArgs...)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []resumeSearchResult
	for rows.Next() {
		var r resumeSearchResult
		if err := rows.Scan(&r.ID, &r.Name, &r.Email, &r.Company, &r.ResumeFilename, &r.Snippet); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		results = append(results, r)
	}
	if results == nil {
		results = []resumeSearchResult{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

// parseBooleanQuery converts a boolean search string into a SQL WHERE clause.
// Tokens are searched against resume_text using ILIKE.
// Supports AND, OR, NOT, parentheses, and "quoted phrases".
func parseBooleanQuery(q string) (string, []interface{}, error) {
	tokens := tokenize(q)
	if len(tokens) == 0 {
		return "", nil, fmt.Errorf("empty query")
	}

	var args []interface{}
	argIdx := 2 // $1 is orgID

	var parts []string
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		upper := strings.ToUpper(tok)

		switch upper {
		case "AND":
			parts = append(parts, "AND")
			i++
		case "OR":
			parts = append(parts, "OR")
			i++
		case "NOT":
			i++
			if i >= len(tokens) {
				return "", nil, fmt.Errorf("NOT without a search term")
			}
			next := tokens[i]
			if strings.ToUpper(next) == "(" {
				// NOT (subexpr)
				sub, subArgs, end, err := parseGroup(tokens, i, argIdx)
				if err != nil {
					return "", nil, err
				}
				parts = append(parts, fmt.Sprintf("NOT (%s)", sub))
				args = append(args, subArgs...)
				argIdx += len(subArgs)
				i = end
			} else {
				parts = append(parts, fmt.Sprintf("resume_text NOT ILIKE $%d", argIdx))
				args = append(args, "%"+next+"%")
				argIdx++
				i++
			}
		case "(":
			sub, subArgs, end, err := parseGroup(tokens, i, argIdx)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, fmt.Sprintf("(%s)", sub))
			args = append(args, subArgs...)
			argIdx += len(subArgs)
			i = end
		default:
			// Regular search term
			parts = append(parts, fmt.Sprintf("resume_text ILIKE $%d", argIdx))
			args = append(args, "%"+tok+"%")
			argIdx++
			i++
		}

		// If next token is not an operator, insert implicit AND
		if i < len(tokens) {
			nextUpper := strings.ToUpper(tokens[i])
			if nextUpper != "AND" && nextUpper != "OR" && len(parts) > 0 {
				last := parts[len(parts)-1]
				if last != "AND" && last != "OR" {
					parts = append(parts, "AND")
				}
			}
		}
	}

	// Clean up: ensure we don't start/end with AND/OR
	result := strings.Join(parts, " ")
	return result, args, nil
}

func parseGroup(tokens []string, start int, argIdx int) (string, []interface{}, int, error) {
	// start points to "("
	depth := 1
	i := start + 1
	var inner []string
	for i < len(tokens) && depth > 0 {
		if tokens[i] == "(" {
			depth++
		} else if tokens[i] == ")" {
			depth--
			if depth == 0 {
				i++
				break
			}
		}
		if depth > 0 {
			inner = append(inner, tokens[i])
		}
		i++
	}
	if depth != 0 {
		return "", nil, 0, fmt.Errorf("mismatched parentheses")
	}

	subQuery := strings.Join(inner, " ")
	where, args, err := parseBooleanQueryAt(subQuery, argIdx)
	if err != nil {
		return "", nil, 0, err
	}
	return where, args, i, nil
}

func parseBooleanQueryAt(q string, startArgIdx int) (string, []interface{}, error) {
	tokens := tokenize(q)
	if len(tokens) == 0 {
		return "", nil, fmt.Errorf("empty group")
	}

	var args []interface{}
	argIdx := startArgIdx

	var parts []string
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		upper := strings.ToUpper(tok)

		switch upper {
		case "AND":
			parts = append(parts, "AND")
			i++
		case "OR":
			parts = append(parts, "OR")
			i++
		case "NOT":
			i++
			if i >= len(tokens) {
				return "", nil, fmt.Errorf("NOT without a search term")
			}
			next := tokens[i]
			parts = append(parts, fmt.Sprintf("resume_text NOT ILIKE $%d", argIdx))
			args = append(args, "%"+next+"%")
			argIdx++
			i++
		case "(":
			sub, subArgs, end, err := parseGroup(tokens, i, argIdx)
			if err != nil {
				return "", nil, err
			}
			parts = append(parts, fmt.Sprintf("(%s)", sub))
			args = append(args, subArgs...)
			argIdx += len(subArgs)
			i = end
		default:
			parts = append(parts, fmt.Sprintf("resume_text ILIKE $%d", argIdx))
			args = append(args, "%"+tok+"%")
			argIdx++
			i++
		}

		if i < len(tokens) {
			nextUpper := strings.ToUpper(tokens[i])
			if nextUpper != "AND" && nextUpper != "OR" && len(parts) > 0 {
				last := parts[len(parts)-1]
				if last != "AND" && last != "OR" {
					parts = append(parts, "AND")
				}
			}
		}
	}

	return strings.Join(parts, " "), args, nil
}

// tokenize splits a query into tokens, respecting quoted phrases.
func tokenize(q string) []string {
	var tokens []string
	runes := []rune(q)
	i := 0
	for i < len(runes) {
		ch := runes[i]

		// Skip whitespace
		if unicode.IsSpace(ch) {
			i++
			continue
		}

		// Parentheses
		if ch == '(' || ch == ')' {
			tokens = append(tokens, string(ch))
			i++
			continue
		}

		// Quoted phrase
		if ch == '"' || ch == '\'' {
			quote := ch
			i++
			start := i
			for i < len(runes) && runes[i] != quote {
				i++
			}
			tokens = append(tokens, string(runes[start:i]))
			if i < len(runes) {
				i++ // skip closing quote
			}
			continue
		}

		// Regular word
		start := i
		for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != '(' && runes[i] != ')' {
			i++
		}
		tokens = append(tokens, string(runes[start:i]))
	}
	return tokens
}
