# Null Pointer / Undefined Access Remediation Strategy

## Bug Pattern

Null pointer errors (or undefined access errors in JavaScript) occur when attempting to access properties or methods on `null` or `undefined` values, resulting in `TypeError: Cannot read property 'X' of null/undefined`.

## Common Scenarios

- Database queries that return no results
- API calls that fail or return empty responses
- Optional parameters that aren't provided
- Array/object access with invalid indices/keys

## Detection

Look for:
- Property access without null checks: `user.name`, `data.email`
- Method calls on potentially null objects: `user.toString()`
- Chained property access: `user.profile.address.city`
- Database query results used directly without validation

## Remediation Strategy

### 1. Identify the Null/Undefined Source

Locate where the potentially null/undefined value originates:
- Database query: `db.prepare('SELECT...').get(id)`
- API response: `await fetch(...).then(r => r.json())`
- Function parameter: `function process(user) { user.name }`
- Array access: `array[index]`

### 2. Add Existence Check

Before accessing properties, verify the object exists:

```javascript
// BAD: No validation
const user = db.prepare('SELECT * FROM users WHERE id = ?').get(userId);
const response = {
  id: user.id,
  name: user.name,
  email: user.email
};

// GOOD: Check existence before access
const user = db.prepare('SELECT * FROM users WHERE id = ?').get(userId);

if (!user) {
  return res.status(404).json({
    error: 'User not found',
    userId
  });
}

const response = {
  id: user.id,
  name: user.name,
  email: user.email
};
```

### 3. Choose Appropriate Error Response

Return meaningful HTTP status codes:
- `404 Not Found` - Resource doesn't exist
- `400 Bad Request` - Invalid input caused the issue
- `500 Internal Server Error` - Unexpected null (should be investigated)

### 4. Handle Nested Property Access

For deeply nested properties, use optional chaining or multiple checks:

```javascript
// Option 1: Optional chaining (modern JavaScript)
const city = user?.profile?.address?.city;

// Option 2: Multiple checks
if (user && user.profile && user.profile.address) {
  const city = user.profile.address.city;
}

// Option 3: Default values
const city = user?.profile?.address?.city || 'Unknown';
```

### 5. Validate Before Complex Operations

For operations like string splitting or array access:

```javascript
// BAD: No validation before split
const emailDomain = user.email.split('@')[1];

// GOOD: Validate existence and format
if (!user) {
  return res.status(404).json({ error: 'User not found' });
}

if (!user.email || typeof user.email !== 'string') {
  return res.status(500).json({ error: 'Invalid user email format' });
}

const emailParts = user.email.split('@');
const emailDomain = emailParts.length > 1 ? emailParts[1] : 'unknown';
```

## Testing

After applying the fix, verify:

1. **Null case**: Request non-existent resource (e.g., user ID 999)
2. **Valid case**: Request existing resource
3. **Edge cases**: Empty strings, malformed data
4. **Error response**: Verify proper status code and message

## Expected Fix for Demo App

For the `/api/buggy/user/:id` endpoint:

```javascript
router.get('/user/:id', (req, res) => {
  const userId = parseInt(req.params.id);
  
  const user = db.prepare('SELECT * FROM users WHERE id = ?').get(userId);
  
  // FIX: Check if user exists before accessing properties
  if (!user) {
    return res.status(404).json({
      error: 'User not found',
      userId
    });
  }
  
  // FIX: Validate email exists before splitting
  const emailDomain = user.email ? user.email.split('@')[1] : 'unknown';
  
  const response = {
    id: user.id,
    name: user.name,
    email: user.email,
    age: user.age,
    emailDomain
  };
  
  res.json(response);
});
```

## Prevention

To prevent this bug in the future:
- Always validate database query results before use
- Use TypeScript for compile-time null checking
- Enable strict null checks in your linter
- Add unit tests for "not found" scenarios
- Use optional chaining (`?.`) for nested property access
- Consider using Result/Option types for operations that may fail
