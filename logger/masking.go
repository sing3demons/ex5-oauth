package logger

import (
	"encoding/json"
	"reflect"
	"strings"
)

type MaskingType string

const (
	MaskingTypeFull    MaskingType = "full"    // แสดงเป็น "***"
	MaskingTypePartial MaskingType = "partial" // แสดงบางส่วน เช่น "abc***xyz"
	MaskingTypeEmail   MaskingType = "email"   // แสดงเป็น "a***@example.com"
	MaskingTypeCard    MaskingType = "card"    // แสดงเป็น "****-****-****-1234"
)

type MaskingRule struct {
	Field   string      // path ใช้ dot notation เช่น "body.password", "result.*.username"
	Type    MaskingType // ประเภทการ mask
	IsArray bool        // true เมื่อต้องการ mask array elements
}

// MaskData applies masking rules to the data
func MaskData(data interface{}, rules []MaskingRule) interface{} {
	if len(rules) == 0 {
		return data
	}

	// Convert to map for easier manipulation
	var dataMap map[string]interface{}
	
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return data
	}
	
	if err := json.Unmarshal(jsonBytes, &dataMap); err != nil {
		return data
	}

	for _, rule := range rules {
		applyMaskingRule(dataMap, rule.Field, rule.Type, rule.IsArray)
	}

	return dataMap
}

func applyMaskingRule(data map[string]interface{}, path string, maskType MaskingType, isArray bool) {
	parts := strings.Split(path, ".")
	applyMaskingRecursive(data, parts, maskType, isArray)
}

func applyMaskingRecursive(data interface{}, pathParts []string, maskType MaskingType, isArray bool) {
	if len(pathParts) == 0 {
		return
	}

	currentPart := pathParts[0]
	remainingParts := pathParts[1:]

	switch v := data.(type) {
	case map[string]interface{}:
		if currentPart == "*" {
			// Apply to all keys
			for key := range v {
				if len(remainingParts) == 0 {
					v[key] = maskValue(v[key], maskType)
				} else {
					applyMaskingRecursive(v[key], remainingParts, maskType, isArray)
				}
			}
		} else if val, exists := v[currentPart]; exists {
			if len(remainingParts) == 0 {
				v[currentPart] = maskValue(val, maskType)
			} else {
				applyMaskingRecursive(val, remainingParts, maskType, isArray)
			}
		}

	case []interface{}:
		if isArray {
			for i := range v {
				if len(remainingParts) == 0 {
					v[i] = maskValue(v[i], maskType)
				} else {
					applyMaskingRecursive(v[i], pathParts, maskType, isArray)
				}
			}
		}
	}
}

func maskValue(value interface{}, maskType MaskingType) interface{} {
	strValue, ok := value.(string)
	if !ok {
		// Try to convert to string
		strValue = toString(value)
	}

	if strValue == "" {
		return value
	}

	switch maskType {
	case MaskingTypeFull:
		return "***"
	case MaskingTypePartial:
		return maskPartial(strValue)
	case MaskingTypeEmail:
		return maskEmail(strValue)
	case MaskingTypeCard:
		return maskCard(strValue)
	default:
		return "***"
	}
}

func maskPartial(s string) string {
	length := len(s)
	if length <= 3 {
		return "***"
	}
	if length <= 6 {
		return string(s[0]) + "***"
	}
	return string(s[0]) + strings.Repeat("*", length-2) + string(s[length-1])
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	
	username := parts[0]
	domain := parts[1]
	
	if len(username) <= 1 {
		return "*@" + domain
	}
	if len(username) <= 3 {
		return string(username[0]) + "***@" + domain
	}
	
	return string(username[0]) + strings.Repeat("*", len(username)-1) + "@" + domain
}

func maskCard(card string) string {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(card, " ", ""), "-", "")
	
	if len(cleaned) < 4 {
		return "****"
	}
	
	last4 := cleaned[len(cleaned)-4:]
	return "****-****-****-" + last4
}

func toString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return string(rune(v.Int()))
	case reflect.Float32, reflect.Float64:
		return string(rune(int(v.Float())))
	default:
		jsonBytes, _ := json.Marshal(value)
		return string(jsonBytes)
	}
}
