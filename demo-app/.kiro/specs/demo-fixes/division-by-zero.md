# Division by Zero Remediation Strategy

## Bug Pattern

Division by zero errors occur when attempting to divide a number by zero or by an empty collection's length, resulting in `Infinity` or `NaN` values.

## Common Scenarios

- Calculating averages from empty arrays
- Computing ratios with zero denominators
- Statistical calculations on empty datasets

## Detection

Look for:
- Division operations: `total / count`, `sum / array.length`
- Array reduce operations followed by division
- Mathematical calculations without denominator validation

## Remediation Strategy

### 1. Identify the Division Operation

Locate the line where division occurs and identify:
- The numerator (dividend)
- The denominator (divisor)
- The context (what's being calculated)

### 2. Add Guard Clause

Before the division, add a check for zero or empty collections:

```javascript
// BAD: No validation
const average = total / products.length;

// GOOD: Guard clause with early return
if (products.length === 0) {
  return res.json({
    average: 0,
    count: 0,
    total: 0,
    message: 'No products available'
  });
}
const average = total / products.length;

// ALTERNATIVE: Ternary operator for inline handling
const average = products.length > 0 ? total / products.length : 0;
```

### 3. Choose Appropriate Default Value

Consider what makes sense for your domain:
- `0` - For counts, sums, or when zero is meaningful
- `null` - When absence of data should be explicit
- `undefined` - When the calculation shouldn't exist
- Custom message - When the client needs context

### 4. Update Response Structure

Ensure the response clearly indicates when no data is available:

```javascript
res.json({
  average: products.length > 0 ? total / products.length : 0,
  count: products.length,
  total,
  hasData: products.length > 0
});
```

## Testing

After applying the fix, verify:

1. **Empty case**: Call endpoint when collection is empty
2. **Single item**: Verify calculation with one item
3. **Multiple items**: Verify normal calculation works
4. **Response structure**: Ensure all fields are present

## Expected Fix for Demo App

For the `/api/buggy/average-price` endpoint:

```javascript
router.get('/average-price', (req, res) => {
  const products = db.prepare('SELECT price FROM products').all();
  
  const total = products.reduce((sum, p) => sum + p.price, 0);
  
  // FIX: Check for empty array before division
  if (products.length === 0) {
    return res.json({
      average: 0,
      count: 0,
      total: 0,
      message: 'No products available'
    });
  }
  
  const average = total / products.length;
  
  res.json({
    average,
    count: products.length,
    total
  });
});
```

## Prevention

To prevent this bug in the future:
- Always validate collection size before division
- Use linting rules to catch potential division by zero
- Add unit tests for empty collection scenarios
- Consider using helper functions for common calculations
