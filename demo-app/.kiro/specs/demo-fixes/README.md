# Demo Fixes - Remediation Strategies

This directory contains remediation strategies for the intentionally buggy endpoints in the AI SRE Demo Application. These strategies guide Kiro CLI in diagnosing and fixing common production errors.

## Overview

The demo application contains several intentional bugs that represent common programming errors found in production systems. Each bug has a corresponding remediation strategy document that provides:

1. **Bug Pattern Description** - What the bug is and why it occurs
2. **Detection Guidelines** - How to identify the bug in code
3. **Remediation Strategy** - Step-by-step fix instructions
4. **Testing Guidance** - How to verify the fix works
5. **Prevention Tips** - How to avoid the bug in the future

## Remediation Strategies

### 1. Division by Zero (`division-by-zero.md`)

**Bug**: Dividing by zero or empty array length, resulting in `Infinity` or `NaN`.

**Affected Endpoint**: `GET /api/buggy/average-price`

**Fix Approach**: Add guard clause to check for empty collections before division.

**Key Learning**: Always validate denominators before division operations.

---

### 2. Null Pointer / Undefined Access (`null-pointer.md`)

**Bug**: Accessing properties on `null` or `undefined` objects.

**Affected Endpoint**: `GET /api/buggy/user/:id`

**Fix Approach**: Add existence checks before property access, return 404 for missing resources.

**Key Learning**: Always validate database query results before accessing properties.

---

### 3. Array Processing Errors (`array-processing.md`)

**Bug**: Off-by-one errors in loops and missing array validation.

**Affected Endpoint**: `POST /api/buggy/process-orders`

**Fix Approach**: 
- Correct loop condition from `i <= array.length` to `i < array.length`
- Validate input is an array
- Validate array elements have required properties

**Key Learning**: Use array methods (`map`, `filter`, `forEach`) instead of manual loops when possible.

---

## How Kiro CLI Uses These Strategies

When the AI SRE Platform triggers a remediation workflow:

1. **Incident Context**: Kiro receives the error message, stack trace, and service metadata
2. **Strategy Selection**: Kiro identifies which remediation strategy applies based on the error pattern
3. **Code Analysis**: Kiro locates the problematic code using the stack trace
4. **Fix Generation**: Kiro applies the remediation strategy to generate a fix
5. **Validation**: Kiro runs tests to verify the fix works
6. **Pull Request**: Kiro creates a PR with the fix and a detailed post-mortem

## Using These Strategies

### For Kiro CLI

When analyzing an incident, Kiro should:

1. Read the relevant strategy document based on the error type
2. Follow the detection guidelines to locate the bug
3. Apply the remediation strategy step-by-step
4. Use the testing guidance to validate the fix
5. Include prevention tips in the post-mortem

### For Developers

These strategies can also be used by human developers:

1. **Learning Resource**: Understand common bug patterns and fixes
2. **Code Review**: Reference when reviewing similar code
3. **Prevention**: Apply prevention tips to avoid bugs
4. **Onboarding**: Help new team members understand common pitfalls

## Expected Fixes

### Division by Zero Fix

```javascript
if (products.length === 0) {
  return res.json({
    average: 0,
    count: 0,
    total: 0,
    message: 'No products available'
  });
}
const average = total / products.length;
```

### Null Pointer Fix

```javascript
if (!user) {
  return res.status(404).json({
    error: 'User not found',
    userId
  });
}
```

### Array Processing Fix

```javascript
// Validate array input
if (!orders || !Array.isArray(orders)) {
  return res.status(400).json({
    error: 'Invalid input: orders must be an array'
  });
}

// Fix loop condition
for (let i = 0; i < orders.length; i++) {  // Changed from <=
  const order = orders[i];
  // ...
}
```

## Testing the Fixes

After Kiro applies a fix, verify:

1. **Original Bug**: The error no longer occurs
2. **Normal Operation**: Valid inputs still work correctly
3. **Edge Cases**: Empty inputs, null values, etc. are handled
4. **Error Messages**: Appropriate HTTP status codes and messages

## Adding New Strategies

To add a new remediation strategy:

1. Create a new markdown file: `{bug-type}.md`
2. Follow the template structure:
   - Bug Pattern
   - Common Scenarios
   - Detection
   - Remediation Strategy
   - Testing
   - Expected Fix
   - Prevention
3. Update this README with the new strategy
4. Add corresponding buggy endpoint to demo app

## Related Files

- **Demo App Bugs**: `demo-app/src/routes/buggy.js`
- **Demo App Tests**: `demo-app/src/routes/buggy.test.js`
- **MCP Configuration**: `demo-app/.kiro/settings/mcp.json`
- **Remediation Workflow**: `.github/workflows/demo-remediate.yml`

## Resources

- [Demo App README](../../README.md)
- [AI SRE Platform Documentation](../../../../README.md)
- [Kiro CLI Documentation](https://docs.kiro.ai)
