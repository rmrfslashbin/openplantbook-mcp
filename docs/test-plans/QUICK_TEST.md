# OpenPlantbook MCP Server - Quick Test Guide

This is a condensed test guide for quickly validating the OpenPlantbook MCP server is working correctly in a fresh Claude Code session.

## Quick Setup Verification

**Prompt:**
```
Is the OpenPlantbook MCP server connected? List the available tools.
```

**Expected:** Should show 4 tools: search_plants, get_plant_care, get_care_summary, compare_conditions

---

## 5-Minute Smoke Test

Run these 5 essential tests to verify basic functionality:

### Test 1: Basic Search (30 seconds)

**Prompt:**
```
Search for "monstera" plants in the OpenPlantbook database.
```

**Expected:** Returns 1-3 Monstera varieties with PIDs like "monstera deliciosa" (lowercase with spaces)

---

### Test 2: Get Plant Details (30 seconds)

**Prompt:**
```
Get the care requirements for "monstera deliciosa" including light, temperature, and moisture needs.
```

**Expected:** Returns complete care data with lux, temperature (°C), and moisture percentage ranges

---

### Test 3: Human-Readable Summary (45 seconds)

**Prompt:**
```
Get a human-readable care summary for "ocimum basilicum".
```

**Expected:** Returns Markdown-formatted summary with interpreted light levels and care instructions

---

### Test 4: Condition Comparison (60 seconds)

**Prompt:**
```
Compare these sensor readings against ideal conditions for "monstera deliciosa":
- Moisture: 45%
- Temperature: 22°C
- Light: 2000 lux
- Humidity: 65%
```

**Expected:** Returns comparison with ✅/❌ indicators showing each parameter is within range

---

### Test 5: Error Handling (30 seconds)

**Prompt:**
```
Try to get care details for a plant with ID "nonexistent-plant-xyz".
```

**Expected:** Returns clear error message (not found), server remains operational

---

## Pass/Fail Criteria

✅ **PASS:** All 5 tests complete successfully with expected results

❌ **FAIL:** Any test fails, returns unexpected errors, or server crashes

## If Tests Fail

1. Check `.mcp.json` configuration
2. Verify API key is valid in `.env`
3. Ensure binary is built: `make build`
4. Check server logs for errors
5. Restart Claude Code and reconnect MCP server

## Full Testing

For comprehensive testing, see [FUNCTIONAL_TESTING.md](./FUNCTIONAL_TESTING.md) which includes 20 detailed test cases.
