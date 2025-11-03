package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rmrfslashbin/openplantbook-go"
	"github.com/rs/xid"
)

// Server implements the MCP server for OpenPlantbook
type Server struct {
	client  *openplantbook.Client
	logger  *slog.Logger
	config  *Config
	version string
}

// New creates a new MCP server instance
func New(config *Config, version string) (*Server, error) {
	// Initialize trace ID for this server instance
	traceID := xid.New().String()

	// Set up structured logging to STDERR
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: config.LogLevel,
	})).With(
		"trace_id", traceID,
		"service", "openplantbook-mcp",
		"version", version,
		"pid", os.Getpid(),
	)

	// Determine authentication method
	var opts []openplantbook.Option
	if config.APIKey != "" {
		logger.Info("using API key authentication")
		opts = append(opts, openplantbook.WithAPIKey(config.APIKey))
	} else {
		logger.Info("using OAuth2 authentication")
		opts = append(opts, openplantbook.WithOAuth2(config.ClientID, config.ClientSecret))
	}

	// Create OpenPlantbook SDK client
	client, err := openplantbook.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("create openplantbook client: %w", err)
	}

	logger.Info("openplantbook client created successfully")

	return &Server{
		client:  client,
		logger:  logger,
		config:  config,
		version: version,
	}, nil
}

// Run starts the MCP server using stdio transport
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("starting openplantbook-mcp server")

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"openplantbook-mcp",
		s.version,
		server.WithToolCapabilities(true),
	)

	// Register all tools
	if err := s.registerTools(mcpServer); err != nil {
		return fmt.Errorf("register tools: %w", err)
	}

	// Start stdio transport
	s.logger.Info("starting stdio server")
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("serve stdio: %w", err)
	}

	return nil
}

// registerTools registers all MCP tools
func (s *Server) registerTools(mcpServer *server.MCPServer) error {
	// Tool 1: search_plants
	searchPlantsSchema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Plant name to search for (common or scientific name)",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Maximum number of results (optional, default: 10)",
			},
		},
		Required: []string{"query"},
	}

	mcpServer.AddTool(mcp.Tool{
		Name:        "search_plants",
		Description: "Search for plants by common name or scientific name in the OpenPlantbook database",
		InputSchema: searchPlantsSchema,
	}, s.handleSearchPlants)

	// Tool 2: get_plant_care
	getPlantCareSchema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"pid": map[string]interface{}{
				"type":        "string",
				"description": "Plant ID (pid) from search results",
			},
			"language": map[string]interface{}{
				"type":        "string",
				"description": "Preferred language code (e.g., 'en', 'de', 'es'), optional",
			},
		},
		Required: []string{"pid"},
	}

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_plant_care",
		Description: "Get detailed care requirements for a specific plant including moisture, temperature, light, and humidity ranges",
		InputSchema: getPlantCareSchema,
	}, s.handleGetPlantCare)

	// Tool 3: get_care_summary
	getCareSummarySchema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"pid": map[string]interface{}{
				"type":        "string",
				"description": "Plant ID (pid) from search results",
			},
			"metric": map[string]interface{}{
				"type":        "boolean",
				"description": "Use metric units (default: true)",
			},
		},
		Required: []string{"pid"},
	}

	mcpServer.AddTool(mcp.Tool{
		Name:        "get_care_summary",
		Description: "Get a human-readable summary of plant care requirements with interpreted ranges",
		InputSchema: getCareSummarySchema,
	}, s.handleGetCareSummary)

	// Tool 4: compare_conditions
	compareConditionsSchema := mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"pid": map[string]interface{}{
				"type":        "string",
				"description": "Plant ID (pid) from search results",
			},
			"current_conditions": map[string]interface{}{
				"type":        "object",
				"description": "Current sensor readings",
				"properties": map[string]interface{}{
					"moisture": map[string]interface{}{
						"type":        "number",
						"description": "Current soil moisture percentage (0-100)",
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "Current temperature in Celsius",
					},
					"light_lux": map[string]interface{}{
						"type":        "number",
						"description": "Current light level in lux",
					},
					"humidity": map[string]interface{}{
						"type":        "number",
						"description": "Current humidity percentage (0-100)",
					},
				},
			},
		},
		Required: []string{"pid", "current_conditions"},
	}

	mcpServer.AddTool(mcp.Tool{
		Name:        "compare_conditions",
		Description: "Compare actual sensor readings against ideal plant care ranges and identify issues",
		InputSchema: compareConditionsSchema,
	}, s.handleCompareConditions)

	s.logger.Info("registered tools", "count", 4)
	return nil
}

// handleSearchPlants handles the search_plants tool
func (s *Server) handleSearchPlants(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	traceID := xid.New().String()
	logger := s.logger.With("trace_id", traceID, "tool", "search_plants")

	// Extract parameters using helper methods
	query, err := request.RequireString("query")
	if err != nil {
		logger.Warn("invalid query parameter", "error", err)
		return mcp.NewToolResultError("query parameter is required and must be a string"), nil
	}

	// Build search options
	opts := &openplantbook.SearchOptions{
		Limit: request.GetInt("limit", 10),
	}

	logger.Info("searching plants", "query", query, "limit", opts.Limit)

	// Call SDK
	results, err := s.client.SearchPlants(ctx, query, opts)
	if err != nil {
		logger.Error("search failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	logger.Info("search completed", "results", len(results))

	// Format response
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logger.Error("marshal results failed", "error", err)
		return mcp.NewToolResultError("failed to format results"), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetPlantCare handles the get_plant_care tool
func (s *Server) handleGetPlantCare(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	traceID := xid.New().String()
	logger := s.logger.With("trace_id", traceID, "tool", "get_plant_care")

	// Extract parameters
	pid, err := request.RequireString("pid")
	if err != nil {
		logger.Warn("invalid pid parameter", "error", err)
		return mcp.NewToolResultError("pid parameter is required and must be a string"), nil
	}

	// Build detail options
	opts := &openplantbook.DetailOptions{
		Language: request.GetString("language", s.config.DefaultLang),
	}

	logger.Info("getting plant care", "pid", pid, "language", opts.Language)

	// Call SDK
	details, err := s.client.GetPlantDetails(ctx, pid, opts)
	if err != nil {
		logger.Error("get details failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get plant details: %v", err)), nil
	}

	logger.Info("plant care retrieved", "pid", details.PID, "alias", details.Alias)

	// Format response
	data, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		logger.Error("marshal details failed", "error", err)
		return mcp.NewToolResultError("failed to format details"), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetCareSummary handles the get_care_summary tool
func (s *Server) handleGetCareSummary(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	traceID := xid.New().String()
	logger := s.logger.With("trace_id", traceID, "tool", "get_care_summary")

	// Extract parameters
	pid, err := request.RequireString("pid")
	if err != nil {
		logger.Warn("invalid pid parameter", "error", err)
		return mcp.NewToolResultError("pid parameter is required and must be a string"), nil
	}

	metric := request.GetBool("metric", true)

	logger.Info("generating care summary", "pid", pid, "metric", metric)

	// Get plant details
	details, err := s.client.GetPlantDetails(ctx, pid, &openplantbook.DetailOptions{
		Language: s.config.DefaultLang,
	})
	if err != nil {
		logger.Error("get details failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get plant details: %v", err)), nil
	}

	// Generate human-readable summary
	summary := formatCareSummary(details, metric)

	logger.Info("care summary generated", "pid", details.PID)

	return mcp.NewToolResultText(summary), nil
}

// handleCompareConditions handles the compare_conditions tool
func (s *Server) handleCompareConditions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	traceID := xid.New().String()
	logger := s.logger.With("trace_id", traceID, "tool", "compare_conditions")

	// Extract parameters
	pid, err := request.RequireString("pid")
	if err != nil {
		logger.Warn("invalid pid parameter", "error", err)
		return mcp.NewToolResultError("pid parameter is required and must be a string"), nil
	}

	// Get the raw arguments to access nested object
	conditions, ok := request.GetArguments()["current_conditions"].(map[string]interface{})
	if !ok {
		logger.Warn("invalid current_conditions parameter")
		return mcp.NewToolResultError("current_conditions parameter is required and must be an object"), nil
	}

	logger.Info("comparing conditions", "pid", pid)

	// Get plant details
	details, err := s.client.GetPlantDetails(ctx, pid, &openplantbook.DetailOptions{
		Language: s.config.DefaultLang,
	})
	if err != nil {
		logger.Error("get details failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to get plant details: %v", err)), nil
	}

	// Compare conditions
	analysis := compareConditions(details, conditions)

	logger.Info("condition comparison completed", "pid", details.PID)

	return mcp.NewToolResultText(analysis), nil
}

// formatCareSummary creates a human-readable care summary
func formatCareSummary(details *openplantbook.PlantDetails, metric bool) string {
	tempUnit := "°C"
	if !metric {
		tempUnit = "°F"
	}

	summary := fmt.Sprintf("# %s (%s)\n\n", details.Alias, details.DisplayPID)
	summary += fmt.Sprintf("Category: %s\n\n", details.Category)
	summary += "## Care Requirements\n\n"

	// Light
	if details.MaxLightLux > 0 {
		summary += fmt.Sprintf("**Light**: %d - %d lux", details.MinLightLux, details.MaxLightLux)
		summary += interpretLightLevel(details.MinLightLux, details.MaxLightLux)
		summary += "\n\n"
	}

	// Temperature
	if details.MaxTemp > 0 {
		if metric {
			summary += fmt.Sprintf("**Temperature**: %.1f - %.1f%s\n\n", details.MinTemp, details.MaxTemp, tempUnit)
		} else {
			minF := details.MinTemp*9/5 + 32
			maxF := details.MaxTemp*9/5 + 32
			summary += fmt.Sprintf("**Temperature**: %.1f - %.1f%s\n\n", minF, maxF, tempUnit)
		}
	}

	// Humidity
	if details.MaxEnvHumid > 0 {
		summary += fmt.Sprintf("**Humidity**: %d - %d%%\n\n", details.MinEnvHumid, details.MaxEnvHumid)
	}

	// Soil Moisture
	if details.MaxSoilMoist > 0 {
		summary += fmt.Sprintf("**Soil Moisture**: %d - %d%%", details.MinSoilMoist, details.MaxSoilMoist)
		summary += interpretMoistureLevel(details.MinSoilMoist, details.MaxSoilMoist)
		summary += "\n\n"
	}

	// Soil EC (Conductivity/Fertilizer)
	if details.MaxSoilEC > 0 {
		summary += fmt.Sprintf("**Fertilizer (EC)**: %d - %d µS/cm\n\n", details.MinSoilEC, details.MaxSoilEC)
	}

	if details.ImageURL != "" {
		summary += fmt.Sprintf("\n[Plant Image](%s)\n", details.ImageURL)
	}

	return summary
}

// interpretLightLevel provides human interpretation of light levels
func interpretLightLevel(min, max int) string {
	avg := (min + max) / 2
	switch {
	case avg < 2000:
		return " (Low light - suitable for shade-tolerant plants)"
	case avg < 10000:
		return " (Medium indirect light - typical indoor lighting)"
	case avg < 25000:
		return " (Bright indirect light - near windows)"
	default:
		return " (Full sun or very bright light - direct sunlight)"
	}
}

// interpretMoistureLevel provides human interpretation of moisture levels
func interpretMoistureLevel(min, max int) string {
	avg := (min + max) / 2
	switch {
	case avg < 20:
		return " (Dry soil - water sparingly)"
	case avg < 40:
		return " (Slightly moist - let soil dry between waterings)"
	case avg < 60:
		return " (Evenly moist - keep soil consistently moist)"
	default:
		return " (Very moist - likes wet conditions)"
	}
}

// compareConditions compares current conditions against ideal ranges
func compareConditions(details *openplantbook.PlantDetails, conditions map[string]interface{}) string {
	analysis := fmt.Sprintf("# Condition Analysis for %s\n\n", details.Alias)
	issues := []string{}
	ok := []string{}

	// Check moisture
	if moisture, exists := conditions["moisture"].(float64); exists && details.MaxSoilMoist > 0 {
		min, max := float64(details.MinSoilMoist), float64(details.MaxSoilMoist)
		if moisture < min {
			diff := min - moisture
			issues = append(issues, fmt.Sprintf("❌ **Soil Moisture Too Low**: Current %.1f%%, needs %.0f-%.0f%% (%.1f%% below minimum)", moisture, min, max, diff))
		} else if moisture > max {
			diff := moisture - max
			issues = append(issues, fmt.Sprintf("❌ **Soil Moisture Too High**: Current %.1f%%, needs %.0f-%.0f%% (%.1f%% above maximum)", moisture, min, max, diff))
		} else {
			ok = append(ok, fmt.Sprintf("✅ **Soil Moisture**: %.1f%% (within %.0f-%.0f%% range)", moisture, min, max))
		}
	}

	// Check temperature
	if temp, exists := conditions["temperature"].(float64); exists && details.MaxTemp > 0 {
		min, max := details.MinTemp, details.MaxTemp
		if temp < min {
			diff := min - temp
			issues = append(issues, fmt.Sprintf("❌ **Temperature Too Low**: Current %.1f°C, needs %.1f-%.1f°C (%.1f°C below minimum)", temp, min, max, diff))
		} else if temp > max {
			diff := temp - max
			issues = append(issues, fmt.Sprintf("❌ **Temperature Too High**: Current %.1f°C, needs %.1f-%.1f°C (%.1f°C above maximum)", temp, min, max, diff))
		} else {
			ok = append(ok, fmt.Sprintf("✅ **Temperature**: %.1f°C (within %.1f-%.1f°C range)", temp, min, max))
		}
	}

	// Check light
	if light, exists := conditions["light_lux"].(float64); exists && details.MaxLightLux > 0 {
		min, max := float64(details.MinLightLux), float64(details.MaxLightLux)
		if light < min {
			diff := min - light
			issues = append(issues, fmt.Sprintf("❌ **Light Too Low**: Current %.0f lux, needs %.0f-%.0f lux (%.0f lux below minimum)", light, min, max, diff))
		} else if light > max {
			diff := light - max
			issues = append(issues, fmt.Sprintf("❌ **Light Too High**: Current %.0f lux, needs %.0f-%.0f lux (%.0f lux above maximum)", light, min, max, diff))
		} else {
			ok = append(ok, fmt.Sprintf("✅ **Light**: %.0f lux (within %.0f-%.0f lux range)", light, min, max))
		}
	}

	// Check humidity
	if humid, exists := conditions["humidity"].(float64); exists && details.MaxEnvHumid > 0 {
		min, max := float64(details.MinEnvHumid), float64(details.MaxEnvHumid)
		if humid < min {
			diff := min - humid
			issues = append(issues, fmt.Sprintf("❌ **Humidity Too Low**: Current %.1f%%, needs %.0f-%.0f%% (%.1f%% below minimum)", humid, min, max, diff))
		} else if humid > max {
			diff := humid - max
			issues = append(issues, fmt.Sprintf("❌ **Humidity Too High**: Current %.1f%%, needs %.0f-%.0f%% (%.1f%% above maximum)", humid, min, max, diff))
		} else {
			ok = append(ok, fmt.Sprintf("✅ **Humidity**: %.1f%% (within %.0f-%.0f%% range)", humid, min, max))
		}
	}

	// Build output
	if len(issues) > 0 {
		analysis += "## Issues Detected\n\n"
		for _, issue := range issues {
			analysis += issue + "\n\n"
		}
	}

	if len(ok) > 0 {
		analysis += "## Conditions Within Range\n\n"
		for _, condition := range ok {
			analysis += condition + "\n\n"
		}
	}

	if len(issues) == 0 && len(ok) == 0 {
		analysis += "No conditions were provided for comparison.\n"
	}

	if len(issues) == 0 && len(ok) > 0 {
		analysis += "\n**Summary**: All monitored conditions are within ideal ranges! 🌱\n"
	} else if len(issues) > 0 {
		analysis += fmt.Sprintf("\n**Summary**: %d condition(s) need attention.\n", len(issues))
	}

	return analysis
}
