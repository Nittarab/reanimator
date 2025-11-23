import express from 'express';
import db from '../db/database.js';

const router = express.Router();

/**
 * BUGGY ENDPOINT 1: Division by Zero
 * This endpoint calculates the average price of products
 * Bug: Doesn't handle the case when there are no products
 */
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

/**
 * BUGGY ENDPOINT 2: Null Pointer / Undefined Access
 * This endpoint gets user details by ID
 * Bug: Doesn't check if user exists before accessing properties
 */
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

/**
 * BUGGY ENDPOINT 3: Array Processing Error
 * This endpoint processes a batch of orders
 * Bug: Doesn't validate array input and has off-by-one error
 */
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

/**
 * BUGGY ENDPOINT 4: Type Coercion Error
 * This endpoint calculates discount
 * Bug: Doesn't validate numeric inputs properly
 */
router.post('/calculate-discount', (req, res) => {
  const { price, discountPercent } = req.body;
  
  // BUG: No type validation - string concatenation instead of math
  const discount = price * discountPercent / 100;
  const finalPrice = price - discount;
  
  res.json({
    originalPrice: price,
    discount,
    finalPrice,
    discountPercent
  });
});

/**
 * BUGGY ENDPOINT 5: SQL Injection Vulnerability (simulated)
 * This endpoint searches users by name
 * Bug: Uses string concatenation instead of parameterized queries
 */
router.get('/search-users', (req, res) => {
  const { name } = req.query;
  
  // BUG: Potential SQL injection if name contains special characters
  // Also doesn't handle undefined name
  const query = `SELECT * FROM users WHERE name LIKE '%${name}%'`;
  
  try {
    const users = db.prepare(query).all();
    res.json({ users });
  } catch (error) {
    // This will catch SQL errors but the endpoint is still vulnerable
    res.status(500).json({ error: error.message });
  }
});

export default router;
