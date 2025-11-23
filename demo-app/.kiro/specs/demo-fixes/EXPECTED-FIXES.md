# Expected Fixes for Demo Application Bugs

This document describes the expected fixes that Kiro CLI should generate for each intentional bug in the demo application. These serve as a reference for validating that the AI SRE Platform is working correctly.

## Overview

The demo application has 5 intentionally buggy endpoints. For each bug, this document provides:

1. **Original Buggy Code** - The code with the bug
2. **Expected Fix** - The corrected code
3. **Key Changes** - What was changed and why
4. **Test Cases** - How to verify the fix works

---

## 1. Division by Zero - Average Price Calculation

### Original Buggy Code

```javascript
router.get('/average-price', (req, res) => {
  const products = db.prepare('SELECT price FROM products').all();
  
  const total = products.reduce((sum, p) => sum + p.price, 0);
  
  // BUG: Division by zero when products array is empty
  const average = total / products.length;
  
  res.json({
    average,
    count: products.length,
    total
  });
});
```

### Expected Fix

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

### Key Changes

1. **Added guard clause**: Check if `products.length === 0` before division
2. **Early return**: Return appropriate response when no products exist
3. **Meaningful default**: Return `0` for average with explanatory message
4. **Preserved structure**: Response format remains consistent

### Test Cases

```bash
# Test with empty database (should return average: 0)
curl http://localhost:3000/api/buggy/average-price

# Test with products (should calculate correctly)
# After seeding database with products
curl http://localhost:3000/api/buggy/average-price
```

**Expected Results**:
- Empty: `{"average": 0, "count": 0, "total": 0, "message": "No products available"}`
- With data: `{"average": 25.5, "count": 3, "total": 76.5}`

---

## 2. Null Pointer - User Lookup

### Original Buggy Code

```javascript
router.get('/user/:id', (req, res) => {
  const userId = parseInt(req.params.id);
  
  const user = db.prepare('SELECT * FROM users WHERE id = ?').get(userId);
  
  // BUG: Accessing properties on potentially null/undefined user
  const response = {
    id: user.id,
    name: user.name,
    email: user.email,
    age: user.age,
    emailDomain: user.email.split('@')[1]
  };
  
  res.json(response);
});
```

### Expected Fix

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

### Key Changes

1. **Added null check**: Verify user exists before accessing properties
2. **Proper HTTP status**: Return 404 for non-existent resources
3. **Email validation**: Check email exists before splitting
4. **Error response**: Provide meaningful error message with context

### Test Cases

```bash
# Test with non-existent user (should return 404)
curl -i http://localhost:3000/api/buggy/user/999

# Test with valid user (should return user data)
curl http://localhost:3000/api/buggy/user/1
```

**Expected Results**:
- Non-existent: `404 {"error": "User not found", "userId": 999}`
- Valid: `200 {"id": 1, "name": "Alice", "email": "alice@example.com", "age": 30, "emailDomain": "example.com"}`

---

## 3. Array Processing - Order Processing

### Original Buggy Code

```javascript
router.post('/process-orders', (req, res) => {
  const { orders } = req.body;
  
  // BUG: Doesn't check if orders is an array or if it exists
  const results = [];
  
  // BUG: Off-by-one error - should be i < orders.length
  for (let i = 0; i <= orders.length; i++) {
    const order = orders[i];
    
    // This will throw when i === orders.length
    const product = db.prepare('SELECT * FROM products WHERE id = ?').get(order.productId);
    
    if (product && product.stock >= order.quantity) {
      const total = product.price * order.quantity;
      
      db.prepare('UPDATE products SET stock = stock - ? WHERE id = ?')
        .run(order.quantity, order.productId);
      
      db.prepare('INSERT INTO orders (user_id, product_id, quantity, total) VALUES (?, ?, ?, ?)')
        .run(order.userId, order.productId, order.quantity, total);
      
      results.push({
        orderId: db.prepare('SELECT last_insert_rowid() as id').get().id,
        status: 'success',
        total
      });
    } else {
      results.push({
        status: 'failed',
        reason: 'insufficient_stock'
      });
    }
  }
  
  res.json({ results });
});
```

### Expected Fix

```javascript
router.post('/process-orders', (req, res) => {
  const { orders } = req.body;
  
  // FIX: Validate orders is an array
  if (!orders || !Array.isArray(orders)) {
    return res.status(400).json({
      error: 'Invalid input: orders must be an array'
    });
  }
  
  if (orders.length === 0) {
    return res.json({
      results: [],
      message: 'No orders to process'
    });
  }
  
  const results = [];
  
  // FIX: Correct loop condition from <= to <
  for (let i = 0; i < orders.length; i++) {
    const order = orders[i];
    
    // FIX: Validate order has required properties
    if (!order || !order.productId || !order.quantity || !order.userId) {
      results.push({
        status: 'failed',
        reason: 'missing_required_fields'
      });
      continue;
    }
    
    const product = db.prepare('SELECT * FROM products WHERE id = ?').get(order.productId);
    
    if (product && product.stock >= order.quantity) {
      const total = product.price * order.quantity;
      
      db.prepare('UPDATE products SET stock = stock - ? WHERE id = ?')
        .run(order.quantity, order.productId);
      
      db.prepare('INSERT INTO orders (user_id, product_id, quantity, total) VALUES (?, ?, ?, ?)')
        .run(order.userId, order.productId, order.quantity, total);
      
      results.push({
        orderId: db.prepare('SELECT last_insert_rowid() as id').get().id,
        status: 'success',
        total
      });
    } else {
      results.push({
        status: 'failed',
        reason: 'insufficient_stock'
      });
    }
  }
  
  res.json({ results });
});
```

### Key Changes

1. **Array validation**: Check if orders exists and is an array
2. **Empty array handling**: Return early for empty arrays
3. **Loop condition fix**: Changed `i <= orders.length` to `i < orders.length`
4. **Element validation**: Check each order has required properties
5. **Graceful degradation**: Continue processing valid orders even if some are invalid

### Test Cases

```bash
# Test with invalid input (should return 400)
curl -X POST http://localhost:3000/api/buggy/process-orders \
  -H "Content-Type: application/json" \
  -d '{"orders": "not an array"}'

# Test with empty array (should return empty results)
curl -X POST http://localhost:3000/api/buggy/process-orders \
  -H "Content-Type: application/json" \
  -d '{"orders": []}'

# Test with valid orders (should process successfully)
curl -X POST http://localhost:3000/api/buggy/process-orders \
  -H "Content-Type: application/json" \
  -d '{"orders": [{"userId": 1, "productId": 1, "quantity": 2}]}'

# Test with missing fields (should handle gracefully)
curl -X POST http://localhost:3000/api/buggy/process-orders \
  -H "Content-Type: application/json" \
  -d '{"orders": [{"userId": 1}]}'
```

**Expected Results**:
- Invalid input: `400 {"error": "Invalid input: orders must be an array"}`
- Empty array: `200 {"results": [], "message": "No orders to process"}`
- Valid orders: `200 {"results": [{"orderId": 1, "status": "success", "total": 50}]}`
- Missing fields: `200 {"results": [{"status": "failed", "reason": "missing_required_fields"}]}`

---

## Alternative Fix: Using Array Methods

For the array processing bug, a more idiomatic JavaScript approach using array methods:

```javascript
router.post('/process-orders', (req, res) => {
  const { orders } = req.body;
  
  if (!orders || !Array.isArray(orders)) {
    return res.status(400).json({
      error: 'Invalid input: orders must be an array'
    });
  }
  
  const results = orders.map((order) => {
    // Validate order structure
    if (!order || !order.productId || !order.quantity || !order.userId) {
      return {
        status: 'failed',
        reason: 'missing_required_fields'
      };
    }
    
    const product = db.prepare('SELECT * FROM products WHERE id = ?').get(order.productId);
    
    if (!product || product.stock < order.quantity) {
      return {
        status: 'failed',
        reason: 'insufficient_stock'
      };
    }
    
    const total = product.price * order.quantity;
    
    db.prepare('UPDATE products SET stock = stock - ? WHERE id = ?')
      .run(order.quantity, order.productId);
    
    db.prepare('INSERT INTO orders (user_id, product_id, quantity, total) VALUES (?, ?, ?, ?)')
      .run(order.userId, order.productId, order.quantity, total);
    
    return {
      orderId: db.prepare('SELECT last_insert_rowid() as id').get().id,
      status: 'success',
      total
    };
  });
  
  res.json({ results });
});
```

**Note**: Either fix is acceptable. The array method approach is more idiomatic and eliminates the off-by-one error entirely by avoiding manual index management.

---

## Validation Checklist

After Kiro generates fixes, verify:

### For All Fixes

- [ ] Original bug no longer occurs
- [ ] Normal operation still works correctly
- [ ] Edge cases are handled gracefully
- [ ] Error messages are clear and actionable
- [ ] HTTP status codes are appropriate
- [ ] Response structure is consistent
- [ ] No new bugs introduced

### For Division by Zero

- [ ] Empty collection returns sensible default
- [ ] Non-empty collection calculates correctly
- [ ] Response includes helpful message

### For Null Pointer

- [ ] Non-existent resource returns 404
- [ ] Existing resource returns data
- [ ] Nested property access is safe

### For Array Processing

- [ ] Invalid input returns 400
- [ ] Empty array handled gracefully
- [ ] Valid array processes correctly
- [ ] Invalid elements handled without crashing
- [ ] Loop doesn't access out-of-bounds indices

---

## Post-Mortem Template

Each fix should include a post-mortem with:

1. **What Broke**: Description of the bug and error
2. **Why It Broke**: Root cause analysis
3. **How the Fix Addresses It**: Explanation of the solution
4. **Test Results**: Verification that the fix works
5. **Prevention**: How to avoid this bug in the future

Example post-mortem structure:

```markdown
# Post-Mortem: Division by Zero in Average Price Calculation

## What Broke
The `/api/buggy/average-price` endpoint threw an error when calculating 
the average price of products when no products existed in the database.

## Why It Broke
The code divided the total price by `products.length` without checking 
if the array was empty, resulting in division by zero (Infinity/NaN).

## How the Fix Addresses It
Added a guard clause to check if `products.length === 0` before performing 
the division. When no products exist, the endpoint now returns a meaningful 
response with average: 0 and an explanatory message.

## Test Results
✓ Empty database returns {"average": 0, "count": 0, "total": 0}
✓ Database with products calculates average correctly
✓ Response structure remains consistent

## Prevention
- Always validate collection size before division
- Add unit tests for empty collection scenarios
- Use linting rules to catch potential division by zero
```

---

## Summary

These expected fixes demonstrate the AI SRE Platform's ability to:

1. **Detect** common bug patterns from error messages and stack traces
2. **Diagnose** root causes by analyzing code context
3. **Fix** bugs with appropriate validation and error handling
4. **Validate** fixes work correctly through testing
5. **Document** changes with clear post-mortems

The fixes prioritize:
- **Safety**: Validate inputs before processing
- **Clarity**: Provide meaningful error messages
- **Consistency**: Maintain response structure
- **Robustness**: Handle edge cases gracefully
