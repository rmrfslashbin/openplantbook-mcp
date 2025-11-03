# OpenPlantbook MCP Server

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)

MCP (Model Context Protocol) server that provides AI assistants like Claude with access to the [OpenPlantbook API](https://open.plantbook.io) - a crowd-sourced database of plant care information.

## Features

- **4 MCP Tools** for plant data access:
  - `search_plants` - Search for plants by name
  - `get_plant_care` - Get detailed care requirements
  - `get_care_summary` - Human-readable care summary
  - `compare_conditions` - Compare sensor readings against ideal ranges
- **Dual Authentication**: Supports both API Key and OAuth2
- **Structured Logging**: JSON logs to STDERR with trace IDs
- **Graceful Shutdown**: Proper signal handling
- **Built on Official SDK**: Uses [openplantbook-go](https://github.com/rmrfslashbin/openplantbook-go) v1.0.0

## Quick Start

### Installation

```bash
# Install from source
git clone https://github.com/rmrfslashbin/openplantbook-mcp.git
cd openplantbook-mcp
make install

# Or build locally
make build
```

### Configuration

Set your OpenPlantbook API credentials via environment variables:

```bash
# Option 1: API Key authentication (recommended for read operations)
export OPENPLANTBOOK_API_KEY="your_api_key_here"

# Option 2: OAuth2 authentication (for full API access)
export OPENPLANTBOOK_CLIENT_ID="your_client_id"
export OPENPLANTBOOK_CLIENT_SECRET="your_client_secret"
```

Get your credentials from: https://open.plantbook.io/apikey/show/

### Claude Desktop Setup

Add to your Claude Desktop config file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "openplantbook": {
      "command": "/path/to/openplantbook-mcp",
      "args": [],
      "env": {
        "OPENPLANTBOOK_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

Restart Claude Desktop to load the server.

## Usage

Once configured, you can ask Claude questions like:

- "What are the care requirements for Monstera deliciosa?"
- "Search for peace lily plants"
- "Compare my plant sensor readings against ideal conditions"
- "Give me a summary of basil care requirements"

### Example Interactions

**Searching for plants:**
```
User: "Search for monstera plants"
Claude: [Uses search_plants tool]
Found 3 Monstera varieties:
- Monstera deliciosa (Swiss Cheese Plant)
- Monstera adansonii (Monkey Mask)
- Monstera friedrichsthalii
```

**Getting care details:**
```
User: "What are the care requirements for monstera-deliciosa?"
Claude: [Uses get_plant_care tool]
Monstera deliciosa needs:
- Light: 1500-3000 lux (Medium indirect light)
- Temperature: 18-27°C
- Humidity: 60-80%
- Soil Moisture: 25-60% (Keep evenly moist)
```

**Comparing conditions:**
```
User: "My monstera has 15% moisture, 22°C temp, and 1800 lux light. How does that compare?"
Claude: [Uses compare_conditions tool]
Analysis for Monstera deliciosa:
- ❌ Soil Moisture Too Low: 15% (needs 25-60%, 10% below minimum)
- ✅ Temperature: 22°C (within 18-27°C range)
- ✅ Light: 1800 lux (within 1500-3000 lux range)

Summary: 1 condition needs attention - increase watering.
```

## Available Tools

### search_plants

Search for plants by common or scientific name.

**Parameters:**
- `query` (string, required): Plant name to search
- `limit` (number, optional): Max results (default: 10)

**Example:**
```json
{
  "query": "tomato",
  "limit": 5
}
```

### get_plant_care

Get detailed care requirements for a specific plant.

**Parameters:**
- `pid` (string, required): Plant ID from search results
- `language` (string, optional): Language code (e.g., "en", "de", "es")

**Example:**
```json
{
  "pid": "monstera-deliciosa",
  "language": "en"
}
```

### get_care_summary

Get a human-readable care summary with interpreted ranges.

**Parameters:**
- `pid` (string, required): Plant ID from search results
- `metric` (boolean, optional): Use metric units (default: true)

**Example:**
```json
{
  "pid": "basil-sweet",
  "metric": true
}
```

### compare_conditions

Compare current sensor readings against ideal plant care ranges.

**Parameters:**
- `pid` (string, required): Plant ID from search results
- `current_conditions` (object, required): Sensor readings
  - `moisture` (number): Soil moisture percentage (0-100)
  - `temperature` (number): Temperature in Celsius
  - `light_lux` (number): Light level in lux
  - `humidity` (number): Humidity percentage (0-100)

**Example:**
```json
{
  "pid": "monstera-deliciosa",
  "current_conditions": {
    "moisture": 45,
    "temperature": 22,
    "light_lux": 2000,
    "humidity": 65
  }
}
```

## Configuration Options

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENPLANTBOOK_API_KEY` | API key for authentication | - |
| `OPENPLANTBOOK_CLIENT_ID` | OAuth2 client ID | - |
| `OPENPLANTBOOK_CLIENT_SECRET` | OAuth2 client secret | - |
| `OPENPLANTBOOK_LOG_LEVEL` | Log level (debug, info, warn, error) | info |
| `OPENPLANTBOOK_CACHE_ENABLED` | Enable caching | true |
| `OPENPLANTBOOK_CACHE_TTL_HOURS` | Cache TTL in hours | 24 |
| `OPENPLANTBOOK_DEFAULT_LANGUAGE` | Default language code | en |

### Config File

Alternatively, create `~/.config/openplantbook-mcp/config.json`:

```json
{
  "api_key": "your_api_key_here",
  "log_level": "info",
  "cache_enabled": true,
  "cache_ttl_hours": 24,
  "default_language": "en"
}
```

Or specify a custom config file:

```bash
openplantbook-mcp -config /path/to/config.json
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run linters
make lint
```

### Project Structure

```
openplantbook-mcp/
├── cmd/
│   └── openplantbook-mcp/  # Main entry point
│       └── main.go
├── internal/
│   └── server/             # MCP server implementation
│       ├── server.go       # Core server with tool handlers
│       └── config.go       # Configuration management
├── examples/
│   ├── config.example.json
│   └── claude_desktop_config.json
├── Makefile
├── README.md
└── LICENSE
```

## Troubleshooting

### Server Not Showing Up in Claude

1. Check Claude Desktop config file location is correct
2. Verify the `command` path points to the built binary
3. Restart Claude Desktop after config changes
4. Check logs in Claude Desktop: Help → View Logs

### Authentication Errors

```
Configuration error: authentication required
```

**Solution:** Set `OPENPLANTBOOK_API_KEY` or both `OPENPLANTBOOK_CLIENT_ID` and `OPENPLANTBOOK_CLIENT_SECRET` in environment or config file.

### Multiple Auth Methods Error

```
Configuration error: multiple authentication methods provided
```

**Solution:** Use either API Key OR OAuth2 credentials, not both.

### Viewing Logs

The server logs to STDERR in JSON format. In Claude Desktop, logs include:
- Trace IDs for request tracking
- Tool invocations with parameters
- API call results
- Error details

## Performance

- **Caching**: API responses cached for 24 hours (plant details) or 1 hour (search results)
- **Rate Limiting**: Respects OpenPlantbook's 200 requests/day limit
- **Concurrent Safe**: All operations are thread-safe

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Related Projects

- **OpenPlantbook SDK**: https://github.com/rmrfslashbin/openplantbook-go
- **OpenPlantbook API**: https://open.plantbook.io
- **MCP Specification**: https://spec.modelcontextprotocol.io/
- **MCP Go SDK**: https://github.com/mark3labs/mcp-go

## Credits

- Built with [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
- Uses [openplantbook-go SDK](https://github.com/rmrfslashbin/openplantbook-go)
- Data from [OpenPlantbook](https://open.plantbook.io) community

## Support

- **Issues**: https://github.com/rmrfslashbin/openplantbook-mcp/issues
- **OpenPlantbook Discord**: https://discord.gg/dguPktq9Zh
- **Email**: code@sigler.io
