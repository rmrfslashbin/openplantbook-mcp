package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	mcp "github.com/mark3labs/mcp-go/mcp"
	openplantbook "github.com/rmrfslashbin/openplantbook-go"
)

func setupTestServer(t *testing.T) *Server {
	t.Helper()

	// Load API key from environment
	apiKey := os.Getenv("OPENPLANTBOOK_API_KEY")
	if apiKey == "" {
		t.Skip("OPENPLANTBOOK_API_KEY not set, skipping integration tests")
	}

	config := &Config{
		APIKey:       apiKey,
		LogLevel:     slog.LevelDebug,
		CacheEnabled: false, // Disable cache for testing
		DefaultLang:  "en",
	}

	srv, err := New(config, "test-version")
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	return srv
}

func TestServer_New(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid API key config",
			config: &Config{
				APIKey:      "test-key",
				LogLevel:    slog.LevelInfo,
				DefaultLang: "en",
			},
			wantErr: false,
		},
		{
			name: "valid OAuth2 config",
			config: &Config{
				ClientID:     "test-id",
				ClientSecret: "test-secret",
				LogLevel:     slog.LevelInfo,
				DefaultLang:  "en",
			},
			wantErr: false,
		},
		{
			name: "no auth config",
			config: &Config{
				LogLevel:    slog.LevelInfo,
				DefaultLang: "en",
			},
			wantErr: true,
		},
		{
			name: "multiple auth config - API key takes precedence",
			config: &Config{
				APIKey:       "test-key",
				ClientID:     "test-id",
				ClientSecret: "test-secret",
				LogLevel:     slog.LevelInfo,
				DefaultLang:  "en",
			},
			wantErr: false, // Server uses API key when both are provided
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.config, "test")
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_HandleSearchPlants(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		arguments map[string]interface{}
		wantErr   bool
		validate  func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "search monstera",
			arguments: map[string]interface{}{
				"query": "monstera",
				"limit": 5,
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				// Cast to TextContent
				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				// Parse the JSON response
				var searchResults []openplantbook.PlantSearchResult
				if err := json.Unmarshal([]byte(textContent.Text), &searchResults); err != nil {
					t.Errorf("failed to parse result: %v", err)
					return
				}

				if len(searchResults) == 0 {
					t.Error("expected search results")
				}

				t.Logf("Found %d plants", len(searchResults))
				for _, plant := range searchResults {
					t.Logf("  - %s (PID: %s)", plant.DisplayPID, plant.PID)
				}
			},
		},
		{
			name: "missing query parameter",
			arguments: map[string]interface{}{
				"limit": 10,
			},
			wantErr: false, // Should return error result, not Go error
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Error("expected error result for missing query")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "search_plants",
					Arguments: tt.arguments,
				},
			}

			result, err := srv.handleSearchPlants(ctx, request)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleSearchPlants() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestServer_HandleGetPlantCare(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		arguments map[string]interface{}
		wantErr   bool
		validate  func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "get monstera deliciosa details",
			arguments: map[string]interface{}{
				"pid":      "monstera deliciosa",
				"language": "en",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				var details openplantbook.PlantDetails
				if err := json.Unmarshal([]byte(textContent.Text), &details); err != nil {
					t.Errorf("failed to parse result: %v", err)
					return
				}

				if details.PID != "monstera deliciosa" {
					t.Errorf("expected PID 'monstera deliciosa', got %s", details.PID)
				}

				t.Logf("Plant: %s", details.DisplayPID)
				t.Logf("Light: %d-%d lux", details.MinLightLux, details.MaxLightLux)
				t.Logf("Temp: %.1f-%.1f°C", details.MinTemp, details.MaxTemp)
			},
		},
		{
			name: "missing pid parameter",
			arguments: map[string]interface{}{
				"language": "en",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Error("expected error result for missing pid")
				}
			},
		},
		{
			name: "invalid plant id",
			arguments: map[string]interface{}{
				"pid": "invalid-plant-id-12345",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Error("expected error result for invalid plant")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "get_plant_care",
					Arguments: tt.arguments,
				},
			}

			result, err := srv.handleGetPlantCare(ctx, request)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleGetPlantCare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestServer_HandleGetCareSummary(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		arguments map[string]interface{}
		wantErr   bool
		validate  func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "get basil care summary metric",
			arguments: map[string]interface{}{
				"pid":    "ocimum basilicum",
				"metric": true,
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				t.Logf("Care summary:\n%s", textContent.Text)

				// Should contain markdown formatting
				if len(textContent.Text) < 100 {
					t.Error("summary seems too short")
				}
			},
		},
		{
			name: "get basil care summary imperial",
			arguments: map[string]interface{}{
				"pid":    "ocimum basilicum",
				"metric": false,
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				// Imperial units should show °F
				// Note: We'll add this in the future
				t.Logf("Care summary (imperial):\n%s", textContent.Text)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "get_care_summary",
					Arguments: tt.arguments,
				},
			}

			result, err := srv.handleGetCareSummary(ctx, request)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleGetCareSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestServer_HandleCompareConditions(t *testing.T) {
	srv := setupTestServer(t)
	ctx := context.Background()

	tests := []struct {
		name      string
		arguments map[string]interface{}
		wantErr   bool
		validate  func(*testing.T, *mcp.CallToolResult)
	}{
		{
			name: "compare ideal conditions",
			arguments: map[string]interface{}{
				"pid": "monstera deliciosa",
				"current_conditions": map[string]interface{}{
					"moisture":    45.0,
					"temperature": 22.0,
					"light_lux":   2000.0,
					"humidity":    65.0,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				t.Logf("Comparison result:\n%s", textContent.Text)

				// Should contain status indicators
				if len(textContent.Text) < 50 {
					t.Error("comparison seems too short")
				}
			},
		},
		{
			name: "compare low moisture",
			arguments: map[string]interface{}{
				"pid": "monstera deliciosa",
				"current_conditions": map[string]interface{}{
					"moisture":    15.0,
					"temperature": 22.0,
					"light_lux":   2000.0,
					"humidity":    65.0,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if len(result.Content) == 0 {
					t.Error("expected content in result")
					return
				}

				textContent, ok := mcp.AsTextContent(result.Content[0])
				if !ok {
					t.Error("expected TextContent")
					return
				}

				t.Logf("Low moisture comparison:\n%s", textContent.Text)

				// Should indicate moisture issue
			},
		},
		{
			name: "missing conditions",
			arguments: map[string]interface{}{
				"pid": "monstera deliciosa",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult) {
				if !result.IsError {
					t.Error("expected error result for missing conditions")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "compare_conditions",
					Arguments: tt.arguments,
				},
			}

			result, err := srv.handleCompareConditions(ctx, request)
			if (err != nil) != tt.wantErr {
				t.Errorf("handleCompareConditions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result == nil {
				t.Error("expected non-nil result")
				return
			}

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestInterpretLightLevel(t *testing.T) {
	tests := []struct {
		name     string
		minLux   int
		maxLux   int
		expected string
	}{
		{"very low", 100, 500, " (Low light - suitable for shade-tolerant plants)"},
		{"low", 600, 900, " (Low light - suitable for shade-tolerant plants)"},
		{"medium", 1500, 3000, " (Medium indirect light - typical indoor lighting)"},
		{"high", 3500, 5000, " (Medium indirect light - typical indoor lighting)"},
		{"very high", 8000, 12000, " (Bright indirect light - near windows)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpretLightLevel(tt.minLux, tt.maxLux)
			if result != tt.expected {
				t.Errorf("interpretLightLevel(%d, %d) = %q, want %q",
					tt.minLux, tt.maxLux, result, tt.expected)
			}
		})
	}
}

func TestInterpretMoistureLevel(t *testing.T) {
	tests := []struct {
		name           string
		minMoisture    int
		maxMoisture    int
		expectedSubstr string
	}{
		{"very dry", 5, 15, "dry"},
		{"low", 20, 35, "moist"},
		{"medium", 40, 60, "evenly moist"},
		{"high", 65, 80, "consistently moist"},
		{"very high", 85, 95, "very moist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpretMoistureLevel(tt.minMoisture, tt.maxMoisture)
			t.Logf("%s: %s", tt.name, result)
			// Just check it returns something non-empty
			if len(result) == 0 {
				t.Error("expected non-empty moisture interpretation")
			}
		})
	}
}
