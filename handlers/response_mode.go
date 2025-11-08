package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Response modes supported by the OAuth server
const (
	ResponseModeQuery    = "query"       // Default for authorization code flow
	ResponseModeFragment = "fragment"    // For implicit flow
	ResponseModeFormPost = "form_post"   // POST to redirect_uri
	ResponseModeJSON     = "json"        // Return JSON response (non-standard but useful)
)

// ResponseMode determines how to return the authorization response
type ResponseMode string

// GetResponseMode determines the response mode from request
// Priority: explicit response_mode parameter > Accept header > default (query)
func GetResponseMode(r *http.Request) ResponseMode {
	// Check explicit response_mode parameter
	if mode := r.URL.Query().Get("response_mode"); mode != "" {
		switch mode {
		case ResponseModeQuery, ResponseModeFragment, ResponseModeFormPost, ResponseModeJSON:
			return ResponseMode(mode)
		}
	}

	// Check Accept header for JSON preference
	accept := r.Header.Get("Accept")
	if accept == "application/json" || accept == "*/*" {
		// If it's an API call (not from browser), prefer JSON
		if r.Header.Get("X-Requested-With") == "XMLHttpRequest" ||
			r.Header.Get("Content-Type") == "application/json" {
			return ResponseModeJSON
		}
	}

	// Default to query for traditional OAuth flow
	return ResponseModeQuery
}

// SendAuthorizationResponse sends the authorization response based on response_mode
func SendAuthorizationResponse(w http.ResponseWriter, r *http.Request, redirectURI string, params map[string]string, responseMode ResponseMode) {
	switch responseMode {
	case ResponseModeJSON:
		sendJSONResponse(w, redirectURI, params)
	case ResponseModeFragment:
		sendFragmentResponse(w, redirectURI, params)
	case ResponseModeFormPost:
		sendFormPostResponse(w, redirectURI, params)
	default: // ResponseModeQuery
		sendQueryResponse(w, redirectURI, params)
	}
}

// sendJSONResponse returns authorization response as JSON
func sendJSONResponse(w http.ResponseWriter, redirectURI string, params map[string]string) {
	response := map[string]interface{}{
		"redirect_uri": redirectURI,
	}
	
	// Add all parameters to response
	for key, value := range params {
		response[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendQueryResponse redirects with parameters in query string (default OAuth behavior)
func sendQueryResponse(w http.ResponseWriter, redirectURI string, params map[string]string) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Invalid redirect URI")
		return
	}

	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()

	http.Redirect(w, &http.Request{}, u.String(), http.StatusFound)
}

// sendFragmentResponse redirects with parameters in URL fragment
func sendFragmentResponse(w http.ResponseWriter, redirectURI string, params map[string]string) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "server_error", "Invalid redirect URI")
		return
	}

	// Build fragment
	fragment := url.Values{}
	for key, value := range params {
		fragment.Set(key, value)
	}
	u.Fragment = fragment.Encode()

	http.Redirect(w, &http.Request{}, u.String(), http.StatusFound)
}

// sendFormPostResponse sends an HTML form that auto-submits to redirect_uri
func sendFormPostResponse(w http.ResponseWriter, redirectURI string, params map[string]string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Generate HTML form
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Authorization Response</title>
</head>
<body onload="document.forms[0].submit()">
    <form method="post" action="` + redirectURI + `">
`
	for key, value := range params {
		html += `        <input type="hidden" name="` + key + `" value="` + value + `"/>
`
	}
	html += `        <noscript>
            <button type="submit">Continue</button>
        </noscript>
    </form>
</body>
</html>`

	w.Write([]byte(html))
}

// SendErrorResponse sends an error response based on response_mode
func SendErrorResponse(w http.ResponseWriter, r *http.Request, redirectURI string, errorCode, errorDescription, state string, responseMode ResponseMode) {
	params := map[string]string{
		"error":             errorCode,
		"error_description": errorDescription,
	}
	if state != "" {
		params["state"] = state
	}

	if redirectURI == "" || responseMode == ResponseModeJSON {
		// If no redirect URI or JSON mode, return JSON error
		respondError(w, http.StatusBadRequest, errorCode, errorDescription)
		return
	}

	SendAuthorizationResponse(w, r, redirectURI, params, responseMode)
}
