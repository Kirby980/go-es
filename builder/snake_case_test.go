package builder

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple",
			input:    "UserName",
			expected: "user_name",
		},
		{
			name:     "Acronym",
			input:    "HTTPServer",
			expected: "http_server",
		},
		{
			name:     "MultiAcronym",
			input:    "XMLHTTPRequest",
			expected: "xmlhttp_request",
		},
		{
			name:     "AllCaps",
			input:    "ID",
			expected: "id",
		},
		{
			name:     "Lowercase",
			input:    "name",
			expected: "name",
		},
		{
			name:     "SingleChar",
			input:    "A",
			expected: "a",
		},
		{
			name:     "TrailingAcronym",
			input:    "GetHTTP",
			expected: "get_http",
		},
		{
			name:     "MixedCase",
			input:    "CreatedAt",
			expected: "created_at",
		},
		{
			name:     "Empty",
			input:    "",
			expected: "",
		},
		{
			name:     "LeadingAcronym",
			input:    "APIVersion",
			expected: "api_version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
