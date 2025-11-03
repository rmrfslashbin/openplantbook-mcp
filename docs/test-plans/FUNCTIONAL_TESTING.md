# OpenPlantbook MCP Server - Functional Testing Guide

This document provides comprehensive test prompts for validating the OpenPlantbook MCP server functionality in a fresh Claude Code session.

## Prerequisites

Before running these tests, ensure:
1. The MCP server is built: `make build`
2. You have valid OpenPlantbook API credentials
3. The `.mcp.json` file is configured with your API key
4. The MCP server is connected to your Claude Code session

## Test Suite Overview

This test suite validates:
- All 4 MCP tools (search_plants, get_plant_care, get_care_summary, compare_conditions)
- Error handling and edge cases
- Data accuracy and formatting
- Integration with the OpenPlantbook API

---

## Test 1: Basic Search - Common Plant

**Objective:** Verify search_plants tool works with a common plant name.

**Prompt:**
```
Search for "tomato" plants using the OpenPlantbook MCP server. Show me the results including the plant IDs and display names.
```

**Expected Results:**
- Returns a list of tomato plant varieties
- Each result includes: `pid`, `display_pid`, `alias`, `category`
- Should find at least 1-3 tomato varieties
- Results are formatted clearly

**Pass Criteria:**
- ✅ Search completes successfully
- ✅ Returns valid JSON data
- ✅ Plant IDs are in lowercase with spaces (e.g., "solanum lycopersicum")
- ✅ Display names are human-readable

---

## Test 2: Search - Specific Scientific Name

**Objective:** Verify search works with scientific names.

**Prompt:**
```
Search for "Monstera deliciosa" in the OpenPlantbook database. How many varieties are found?
```

**Expected Results:**
- Finds Monstera deliciosa and potentially related varieties
- Returns results with scientific and common names
- Includes category information

**Pass Criteria:**
- ✅ Search finds at least 1 result
- ✅ Results include "monstera deliciosa" or similar PID
- ✅ Category is included in results

---

## Test 3: Search - Limit Parameter

**Objective:** Verify the limit parameter works correctly.

**Prompt:**
```
Search for "basil" plants but limit the results to only 3 plants. How many results did you get?
```

**Expected Results:**
- Returns exactly 3 results or fewer (if fewer than 3 exist)
- Results are properly formatted

**Pass Criteria:**
- ✅ Number of results respects the limit
- ✅ Search completes successfully
- ✅ Results are not truncated or corrupted

---

## Test 4: Search - No Results

**Objective:** Verify graceful handling when no plants match.

**Prompt:**
```
Search for "xyzabc123notaplant" in the OpenPlantbook database. What happens?
```

**Expected Results:**
- Returns empty results array
- No error is thrown
- Clear indication that no plants were found

**Pass Criteria:**
- ✅ Returns valid JSON with empty results
- ✅ No server error or crash
- ✅ Clear message about no results

---

## Test 5: Get Plant Care Details - Complete Information

**Objective:** Verify get_plant_care retrieves full plant details.

**Prompt:**
```
Get the complete care details for "monstera deliciosa" from OpenPlantbook. Show me the light, temperature, humidity, and moisture requirements.
```

**Expected Results:**
- Returns complete PlantDetails object
- Includes all care parameters:
  - Light: min/max lux values
  - Temperature: min/max in Celsius
  - Humidity: min/max percentage
  - Soil moisture: min/max percentage
- Includes image URL if available

**Pass Criteria:**
- ✅ All care parameters are present and numeric
- ✅ Light values are in reasonable lux range (e.g., 100-50000)
- ✅ Temperature values are in Celsius
- ✅ Humidity and moisture are 0-100 range
- ✅ Data is properly formatted JSON

---

## Test 6: Get Plant Care - Invalid Plant ID

**Objective:** Verify error handling for non-existent plant IDs.

**Prompt:**
```
Try to get care details for a plant with ID "invalid-plant-id-12345". What error do you get?
```

**Expected Results:**
- Returns an error indicating plant not found
- Error is clear and actionable
- Server doesn't crash

**Pass Criteria:**
- ✅ Returns appropriate error message
- ✅ Error indicates plant not found or invalid ID
- ✅ MCP server remains operational

---

## Test 7: Get Care Summary - Metric Units

**Objective:** Verify care summary with human-readable formatting (metric).

**Prompt:**
```
Get a human-readable care summary for "basil-sweet" using metric units. Show me the formatted output.
```

**Expected Results:**
- Returns Markdown-formatted care summary
- Includes interpreted light levels (e.g., "Medium indirect light")
- Includes interpreted moisture levels (e.g., "Keep evenly moist")
- Temperature in Celsius
- Humidity in percentage
- Clear, readable formatting

**Pass Criteria:**
- ✅ Output is Markdown formatted
- ✅ Includes light level interpretation
- ✅ Includes moisture level interpretation
- ✅ Temperature values with °C
- ✅ Easy to read and understand

---

## Test 8: Get Care Summary - Imperial Units

**Objective:** Verify care summary supports imperial units.

**Prompt:**
```
Get a care summary for "tomato-common" using imperial units (Fahrenheit). How are the temperatures displayed?
```

**Expected Results:**
- Returns care summary with imperial units
- Temperature in Fahrenheit (if supported)
- Other parameters appropriately formatted

**Pass Criteria:**
- ✅ Summary is generated successfully
- ✅ Temperature units are appropriate
- ✅ Formatting is clear

---

## Test 9: Compare Conditions - Ideal Conditions

**Objective:** Verify condition comparison with ideal sensor readings.

**Prompt:**
```
I have a Monstera deliciosa with these sensor readings:
- Soil moisture: 45%
- Temperature: 22°C
- Light: 2000 lux
- Humidity: 70%

Compare these conditions against the ideal requirements for this plant.
```

**Expected Results:**
- Returns Markdown-formatted comparison
- Each parameter shows status (OK or needs attention)
- Uses visual indicators (✅ or ❌)
- Provides specific guidance if anything is out of range
- Shows how far off any out-of-range values are

**Pass Criteria:**
- ✅ All 4 conditions are analyzed
- ✅ Clear OK/Problem indicators
- ✅ Specific feedback for each parameter
- ✅ Numerical differences shown for out-of-range values

---

## Test 10: Compare Conditions - Low Moisture Alert

**Objective:** Verify comparison detects low moisture condition.

**Prompt:**
```
Compare these conditions for "monstera deliciosa":
- Soil moisture: 15%
- Temperature: 22°C
- Light: 2000 lux
- Humidity: 65%

What does the comparison tell me about the moisture level?
```

**Expected Results:**
- Detects moisture is too low
- Shows how many percentage points below minimum
- Clear indication that watering is needed
- Other parameters shown as OK

**Pass Criteria:**
- ✅ Moisture flagged as too low
- ✅ Specific percentage difference shown
- ✅ Actionable recommendation (e.g., "increase watering")
- ✅ Other parameters correctly assessed

---

## Test 11: Compare Conditions - Multiple Issues

**Objective:** Verify comparison handles multiple out-of-range parameters.

**Prompt:**
```
Compare these conditions for "basil-sweet":
- Soil moisture: 10%
- Temperature: 10°C
- Light: 500 lux
- Humidity: 30%

What problems does the analysis identify?
```

**Expected Results:**
- Identifies multiple issues (likely all 4 parameters)
- Each issue clearly explained
- Specific recommendations for each
- Summary of total issues

**Pass Criteria:**
- ✅ All out-of-range parameters identified
- ✅ Clear description of each issue
- ✅ Specific numerical differences shown
- ✅ Summary statement (e.g., "3 conditions need attention")

---

## Test 12: Compare Conditions - Missing Parameters

**Objective:** Verify error handling when sensor data is incomplete.

**Prompt:**
```
Try to compare conditions for "monstera deliciosa" with only these readings:
- Temperature: 22°C
- Light: 2000 lux

What happens when moisture and humidity are missing?
```

**Expected Results:**
- Handles missing parameters gracefully
- Either returns error or analyzes only provided parameters
- Clear about what couldn't be analyzed

**Pass Criteria:**
- ✅ No server crash
- ✅ Clear handling of missing data
- ✅ Appropriate error or partial analysis

---

## Test 13: Search and Detail Workflow

**Objective:** Verify complete workflow from search to details.

**Prompt:**
```
I want to grow peace lilies. First search for peace lily plants, then get the complete care requirements for the first result you find.
```

**Expected Results:**
- Performs search successfully
- Identifies peace lily plant ID
- Retrieves detailed care information
- Provides comprehensive care guidance

**Pass Criteria:**
- ✅ Search finds peace lily varieties
- ✅ Successfully retrieves details using PID from search
- ✅ Complete care information provided
- ✅ Workflow completes without errors

---

## Test 14: Multiple Sequential Requests

**Objective:** Verify server handles multiple requests in sequence.

**Prompt:**
```
Perform these tasks in order:
1. Search for "snake plant"
2. Get care details for the first result
3. Get a care summary for the same plant
4. Compare my current conditions (moisture: 30%, temp: 20°C, light: 1000 lux, humidity: 40%) against ideal

Do all 4 tasks for the snake plant.
```

**Expected Results:**
- All 4 requests complete successfully
- Data is consistent across requests
- No timeout or connection issues
- Proper flow from search → details → summary → comparison

**Pass Criteria:**
- ✅ All 4 tools execute successfully
- ✅ Data consistency (same plant across all requests)
- ✅ No errors or timeouts
- ✅ Logical progression through workflow

---

## Test 15: Language Parameter

**Objective:** Verify language parameter for get_plant_care.

**Prompt:**
```
Get the plant care details for "monstera deliciosa" in German (language code: "de"). Are any fields translated?
```

**Expected Results:**
- Request completes with language parameter
- May return translated content or same content
- No errors from language parameter

**Pass Criteria:**
- ✅ Request succeeds with language parameter
- ✅ Returns valid plant data
- ✅ No errors from language code

---

## Test 16: Rate Limiting Behavior

**Objective:** Verify graceful handling if rate limited.

**Prompt:**
```
If I've been making many requests to the OpenPlantbook API, what would happen if I hit the rate limit? Try searching for "cactus" and see if there's any rate limit indication.
```

**Expected Results:**
- Request either succeeds or returns clear rate limit error
- Error message is informative
- Server remains operational

**Pass Criteria:**
- ✅ Request succeeds OR returns clear rate limit error
- ✅ If rate limited, error message is clear
- ✅ Server doesn't crash

---

## Test 17: Special Characters in Search

**Objective:** Verify search handles special characters.

**Prompt:**
```
Search for plants with special characters like "Bird of Paradise" or "String of Pearls". Do these searches work correctly?
```

**Expected Results:**
- Searches with spaces work correctly
- Common names with special punctuation handled
- Results are relevant

**Pass Criteria:**
- ✅ Searches complete successfully
- ✅ Finds relevant plants
- ✅ No errors from special characters

---

## Test 18: Edge Case - Empty Search Query

**Objective:** Verify error handling for empty search.

**Prompt:**
```
What happens if you try to search for an empty string "" in the plant database?
```

**Expected Results:**
- Returns validation error
- Clear message about empty query
- Server remains operational

**Pass Criteria:**
- ✅ Returns appropriate error
- ✅ Error message is clear
- ✅ No server crash

---

## Test 19: Cache Behavior

**Objective:** Verify caching works for repeated requests.

**Prompt:**
```
Request the care details for "monstera deliciosa" twice in a row. Is the second request faster? (This tests the SDK's caching behavior)
```

**Expected Results:**
- Both requests succeed
- Second request may be faster (cache hit)
- Data is identical

**Pass Criteria:**
- ✅ Both requests succeed
- ✅ Data is consistent
- ✅ No errors

---

## Test 20: End-to-End Real-World Scenario

**Objective:** Complete real-world use case from start to finish.

**Prompt:**
```
I'm a beginner plant parent and want to start with easy houseplants. Help me by:

1. Searching for "pothos" plants
2. Getting the care summary for pothos
3. Telling me if my apartment conditions are suitable:
   - My apartment gets about 1500 lux of indirect light
   - Temperature is usually 21°C
   - Humidity around 50%
   - I can water to maintain soil moisture around 40%

Use the OpenPlantbook MCP tools to help me decide if pothos is a good choice.
```

**Expected Results:**
- Complete workflow executes smoothly
- Search finds pothos varieties
- Care summary is clear and helpful
- Condition comparison provides actionable guidance
- Claude provides helpful interpretation

**Pass Criteria:**
- ✅ All tools execute successfully
- ✅ Data is accurate and helpful
- ✅ Provides clear recommendation
- ✅ Workflow feels natural and useful

---

## Test Summary Checklist

After running all tests, verify:

- [ ] All 4 MCP tools are functional
- [ ] Search works with various query types
- [ ] Plant care details are comprehensive
- [ ] Care summaries are readable and formatted
- [ ] Condition comparisons are accurate and helpful
- [ ] Error handling is graceful and informative
- [ ] Server remains stable throughout testing
- [ ] Data accuracy matches OpenPlantbook API
- [ ] Workflows feel natural and useful
- [ ] No crashes, hangs, or connection issues

## Reporting Issues

If any tests fail, document:
1. Test number and name
2. Exact prompt used
3. Expected vs actual result
4. Error messages or logs
5. Steps to reproduce

## Success Criteria

The MCP server is considered fully functional when:
- ✅ At least 18 out of 20 tests pass completely
- ✅ All 4 tools execute successfully in at least one test each
- ✅ No critical errors or crashes occur
- ✅ Error handling is appropriate for edge cases
- ✅ Real-world scenarios (Test 20) complete successfully
