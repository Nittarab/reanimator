# AI SRE Demo App - Remediation Guide

This guide explains how Kiro CLI should use the specs and MCP configuration in this directory to remediate incidents in the demo application.

## Quick Start

When a remediation workflow is triggered for this repository:

1. **Read the incident context** - Error message, stack trace, service metadata
2. **Identify the bug pattern** - Match error to one of the remediation strategies
3. **Query Sentry via MCP** - Get additional context about the error
4. **Apply the fix** - Follow the remediation strategy
5. **Validate** - Run tests to verify the fix works
6. **Create PR** - Include post-mortem with fix explanation

## Directory Structure

```
demo-app/.kiro/
├── settings/
│   ├── mcp.json              # MCP server configuration (Sentry)
│   └── README.md             # MCP setup guide
├── specs/
│   └── demo-fixes/
│       ├── README.md         # Overview of all strategies
│       ├── division-by-zero.md
│       ├── null-pointer.md
│       ├── array-processing.md
│       └── EXPECTED-FIXES.md # Reference for validation
└── REMEDIATION-GUIDE.md      # This file
```

## Workflow

### Step 1: Analyze the Incident

**Input**: Incident data from GitHub Actions workflow
```json
{
  "incident_id": "inc_123",
  "error_message": "TypeError: Cannot read property 'id' of undefined",
  "stack_trace": "at /app/src/routes/buggy.js:35:15",
  "service_name": "demo-app",
  "timestamp": "2024-01-15T10:00:00Z"
}
```

**Action**: Parse the error message to identify the bug pattern
- `"division by zero"` or `"Infinity"` → Division by Zero
- `"Cannot read property"` or `"undefined"` → Null Pointer
- `"Cannot read property 'X' of undefined"` in loop → Array Processing

### Step 2: Query Sentry for Context

**MCP Server**: Configured in `.kiro/settings/mcp.json`

**Queries to make**:
```javascript
// Get full error details
const issue = await sentry.get_issue(issueId);

// Get recent occurrences
const events = await sentry.get_events(issueId, { limit: 10 });

// Get stack trace with context
const stacktrace = await sentry.get_event_stacktrace(eventId);

// Get breadcrumbs (user actions)
const breadcrumbs = await sentry.get_event_breadcrumbs(eventId);
```

**What to look for**:
- Full stack trace with line numbers
- Request parameters that triggered the error
- Frequency of the error
- Related errors or patterns

### Step 3: Locate the Buggy Code

**Use the stack trace** to find the exact file and line:
```
at /app/src/routes/buggy.js:35:15
```

**Read the file** and surrounding context:
- 10 lines before the error
- 10 lines after the error
- The entire function containing the error

**Identify the bug pattern**:
- Division without denominator check?
- Property access without null check?
- Loop with incorrect condition?

### Step 4: Select Remediation Strategy

**Match the bug to a strategy**:

| Error Pattern | Strategy File | Key Fix |
|--------------|---------------|---------|
| Division by zero, Infinity, NaN | `division-by-zero.md` | Add guard clause for empty collections |
| Cannot read property of undefined/null | `null-pointer.md` | Add existence check, return 404 |
| Array index out of bounds | `array-processing.md` | Fix loop condition, validate array |

**Read the strategy document** for:
- Detection guidelines
- Step-by-step fix instructions
- Testing guidance
- Expected fix example

### Step 5: Generate the Fix

**Follow the strategy** step-by-step:

1. **Add validation** - Check for null, empty, or invalid inputs
2. **Add error handling** - Return appropriate HTTP status codes
3. **Preserve structure** - Keep response format consistent
4. **Add helpful messages** - Explain what went wrong

**Example for Null Pointer**:
```javascript
// 1. Add existence check
if (!user) {
  return res.status(404).json({
    error: 'User not found',
    userId
  });
}

// 2. Validate before operations
const emailDomain = user.email ? user.email.split('@')[1] : 'unknown';
```

### Step 6: Validate the Fix

**Run the test suite**:
```bash
cd demo-app
npm test
```

**Test specific scenarios**:
- The original bug case (should now be handled)
- Normal operation (should still work)
- Edge cases (empty, null, invalid inputs)

**Check test results**:
- All tests should pass
- No new errors introduced
- Edge cases handled gracefully

### Step 7: Create Pull Request

**Branch naming**: `fix/incident-{incident_id}`
```bash
git checkout -b fix/incident-inc_123
```

**Commit message**:
```
Fix: Handle null user in GET /api/buggy/user/:id

Resolves incident inc_123

- Add null check before accessing user properties
- Return 404 for non-existent users
- Validate email before splitting

Fixes TypeError: Cannot read property 'id' of undefined
```

**Post-Mortem** (in PR description):

```markdown
# Post-Mortem: Null Pointer in User Lookup

## What Broke
The `/api/buggy/user/:id` endpoint threw a TypeError when requesting 
a non-existent user ID (e.g., 999).

## Why It Broke
The code accessed properties on the `user` object without checking if 
the database query returned a result. When no user exists, `user` is 
`undefined`, causing the error.

## How the Fix Addresses It
1. Added null check: `if (!user)` before accessing properties
2. Return 404 status code for non-existent users
3. Added email validation before splitting

## Test Results
✓ Non-existent user returns 404 with error message
✓ Valid user returns data correctly
✓ Email domain extracted safely

## Prevention
- Always validate database query results
- Use TypeScript for compile-time null checking
- Add unit tests for "not found" scenarios
```

## Bug Pattern Reference

### Division by Zero

**Error Messages**:
- `"Infinity"`
- `"NaN"`
- `"division by zero"`

**Stack Trace Location**: Line with division operation
```javascript
const average = total / products.length;  // ← Error here
```

**Fix**: Add guard clause
```javascript
if (products.length === 0) {
  return res.json({ average: 0, count: 0, total: 0 });
}
```

### Null Pointer

**Error Messages**:
- `"Cannot read property 'X' of undefined"`
- `"Cannot read property 'X' of null"`
- `"undefined is not an object"`

**Stack Trace Location**: Line accessing property
```javascript
const response = { id: user.id };  // ← Error here
```

**Fix**: Add existence check
```javascript
if (!user) {
  return res.status(404).json({ error: 'User not found' });
}
```

### Array Processing

**Error Messages**:
- `"Cannot read property 'X' of undefined"` (in loop)
- `"array[i] is undefined"`

**Stack Trace Location**: Inside loop body
```javascript
for (let i = 0; i <= orders.length; i++) {  // ← Bug here
  const order = orders[i];  // ← Error here when i === length
}
```

**Fix**: Correct loop condition and validate array
```javascript
if (!Array.isArray(orders)) {
  return res.status(400).json({ error: 'Invalid input' });
}
for (let i = 0; i < orders.length; i++) {  // Changed <= to <
```

## Testing Checklist

After generating a fix, verify:

- [ ] Original error no longer occurs
- [ ] Normal operation still works
- [ ] Edge cases handled (empty, null, invalid)
- [ ] HTTP status codes appropriate (404, 400, 500)
- [ ] Error messages clear and actionable
- [ ] Response structure consistent
- [ ] All tests pass
- [ ] No new bugs introduced

## Common Pitfalls

### Don't Over-Fix

❌ **Wrong**: Adding unnecessary complexity
```javascript
// Too much validation
if (!user || typeof user !== 'object' || !user.hasOwnProperty('id')) {
  // ...
}
```

✅ **Right**: Simple, clear validation
```javascript
if (!user) {
  return res.status(404).json({ error: 'User not found' });
}
```

### Don't Change Unrelated Code

❌ **Wrong**: Refactoring while fixing
```javascript
// Don't rename variables, restructure, or optimize
```

✅ **Right**: Minimal fix for the specific bug
```javascript
// Only add the necessary validation
```

### Don't Ignore Edge Cases

❌ **Wrong**: Only fixing the reported case
```javascript
if (products.length === 0) {
  return res.json({ average: 0 });
}
// What about negative length? (shouldn't happen, but...)
```

✅ **Right**: Handle all edge cases
```javascript
if (!products || products.length === 0) {
  return res.json({ average: 0, count: 0, total: 0 });
}
```

## Resources

- **Remediation Strategies**: `.kiro/specs/demo-fixes/`
- **Expected Fixes**: `.kiro/specs/demo-fixes/EXPECTED-FIXES.md`
- **MCP Configuration**: `.kiro/settings/mcp.json`
- **Demo App Code**: `src/routes/buggy.js`
- **Tests**: `src/routes/buggy.test.js`

## Support

If you encounter issues:

1. **Check the strategy docs** - They contain detailed guidance
2. **Review expected fixes** - See what the fix should look like
3. **Query Sentry** - Get more context about the error
4. **Ask for help** - Include incident ID and error details in the PR

## Example End-to-End Flow

```
1. Incident received: "Cannot read property 'id' of undefined"
   ↓
2. Query Sentry: Get full stack trace and context
   ↓
3. Identify pattern: Null pointer error
   ↓
4. Read strategy: specs/demo-fixes/null-pointer.md
   ↓
5. Locate code: src/routes/buggy.js:35
   ↓
6. Generate fix: Add null check before property access
   ↓
7. Run tests: npm test (all pass)
   ↓
8. Create PR: With post-mortem and fix explanation
   ↓
9. Success: Incident resolved, PR ready for review
```
