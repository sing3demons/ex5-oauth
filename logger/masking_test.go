package logger

import (
	"encoding/json"
	"testing"
)

func TestMaskPartial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Short string", "ab", "***"},
		{"Medium string", "abcd", "a***"},
		{"Long string", "password123", "p*********3"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskPartial(tt.input)
			if result != tt.expected {
				t.Errorf("maskPartial(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal email", "user@example.com", "u***@example.com"},
		{"Short username", "ab@example.com", "a***@example.com"},
		{"Single char username", "a@example.com", "*@example.com"},
		{"Long username", "verylongusername@example.com", "v*******************@example.com"},
		{"Invalid email", "notanemail", "***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskEmail(tt.input)
			if result != tt.expected {
				t.Errorf("maskEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskCard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Card with dashes", "1234-5678-9012-3456", "****-****-****-3456"},
		{"Card with spaces", "1234 5678 9012 3456", "****-****-****-3456"},
		{"Card without separator", "1234567890123456", "****-****-****-3456"},
		{"Short card", "123", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskCard(tt.input)
			if result != tt.expected {
				t.Errorf("maskCard(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskData(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		rules    []MaskingRule
		expected string
	}{
		{
			name: "Mask password field",
			data: map[string]interface{}{
				"username": "john",
				"password": "secret123",
			},
			rules: []MaskingRule{
				{Field: "password", Type: MaskingTypeFull},
			},
			expected: `{"password":"***","username":"john"}`,
		},
		{
			name: "Mask nested field",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John Doe",
					"password": "secret123",
				},
			},
			rules: []MaskingRule{
				{Field: "user.password", Type: MaskingTypeFull},
			},
			expected: `{"user":{"name":"John Doe","password":"***"}}`,
		},
		{
			name: "Mask email",
			data: map[string]interface{}{
				"email": "user@example.com",
				"name":  "John",
			},
			rules: []MaskingRule{
				{Field: "email", Type: MaskingTypeEmail},
			},
			expected: `{"email":"u***@example.com","name":"John"}`,
		},
		{
			name: "Mask all fields with wildcard",
			data: map[string]interface{}{
				"users": map[string]interface{}{
					"user1": "password1",
					"user2": "password2",
				},
			},
			rules: []MaskingRule{
				{Field: "users.*", Type: MaskingTypeFull},
			},
			expected: `{"users":{"user1":"***","user2":"***"}}`,
		},
		{
			name: "Mask array elements",
			data: map[string]interface{}{
				"passwords": []interface{}{"pass1", "pass2", "pass3"},
			},
			rules: []MaskingRule{
				{Field: "passwords", Type: MaskingTypeFull, IsArray: true},
			},
			expected: `{"passwords":["***","***","***"]}`,
		},
		{
			name: "Multiple masking rules",
			data: map[string]interface{}{
				"username": "john",
				"password": "secret123",
				"email":    "john@example.com",
				"card":     "1234-5678-9012-3456",
			},
			rules: []MaskingRule{
				{Field: "password", Type: MaskingTypeFull},
				{Field: "email", Type: MaskingTypeEmail},
				{Field: "card", Type: MaskingTypeCard},
			},
			expected: `{"card":"****-****-****-3456","email":"j***@example.com","password":"***","username":"john"}`,
		},
		{
			name: "Nested array masking",
			data: map[string]interface{}{
				"result": []interface{}{
					map[string]interface{}{"username": "user1", "password": "pass1"},
					map[string]interface{}{"username": "user2", "password": "pass2"},
				},
			},
			rules: []MaskingRule{
				{Field: "result.password", Type: MaskingTypeFull, IsArray: true},
			},
			expected: `{"result":[{"password":"***","username":"user1"},{"password":"***","username":"user2"}]}`,
		},
		{
			name: "Partial masking",
			data: map[string]interface{}{
				"token": "abc123xyz789",
			},
			rules: []MaskingRule{
				{Field: "token", Type: MaskingTypePartial},
			},
			expected: `{"token":"a**********9"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskData(tt.data, tt.rules)
			
			resultJSON, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("Failed to marshal result: %v", err)
			}

			if string(resultJSON) != tt.expected {
				t.Errorf("MaskData() = %s, want %s", string(resultJSON), tt.expected)
			}
		})
	}
}

func TestMaskDataNoRules(t *testing.T) {
	data := map[string]interface{}{
		"username": "john",
		"password": "secret",
	}

	result := MaskData(data, []MaskingRule{})
	
	if result == nil {
		t.Error("Expected result to not be nil")
	}
}

func TestMaskDataInvalidPath(t *testing.T) {
	data := map[string]interface{}{
		"username": "john",
	}

	rules := []MaskingRule{
		{Field: "nonexistent.field", Type: MaskingTypeFull},
	}

	result := MaskData(data, rules)
	resultJSON, _ := json.Marshal(result)
	expectedJSON, _ := json.Marshal(data)

	if string(resultJSON) != string(expectedJSON) {
		t.Errorf("MaskData with invalid path should not modify data")
	}
}
