package mlog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"oauth2-server/logger"
	"strings"
)

func L(r *http.Request) *logger.Logger {
	if r == nil || r.Context() == nil {
		return logger.NewLogger("", "")
	}
	l, ok := r.Context().Value("logger").(*logger.Logger)
	if !ok || l == nil {
		return logger.NewLogger("", "")
	}

	return l
}

func InitLog(r *http.Request, xTid, xSid string, masking ...logger.MaskingRule) *logger.Logger {
	l := L(r)
	l.StartTransaction(xTid, xSid)

	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = strings.Join(values, ", ")
		} else {
			headers[key] = ""
		}
	}

	body := new(map[string]any)
	if r.Method != http.MethodGet {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		body = nil
	}

	// Restore the request body so it can be read again later
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	json.Unmarshal(bodyBytes, &body)
}

	l.Info(logger.ActionInfo{
		Action:            "inbound",
		ActionDescription: "get -> " + r.URL.RawPath,
	}, map[string]any{
		"method":  r.Method,
		"url":     r.URL.String(),
		"headers": headers,
		"query":   r.URL.Query(),
		"body":    body,
	}, masking...)
	return l
}

func ResponseJson(w http.ResponseWriter, data any, masking ...logger.MaskingRule) {}
