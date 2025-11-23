import express from 'express';
import db from '../db/database.js';

const router = express.Router();

/**
 * Health check endpoint
 */
router.get('/health', (req, res) => {
  try {
    // Check database connectivity
    db.prepare('SELECT 1').get();
    
    res.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'ai-sre-demo-app',
      database: 'connected'
    });
  } catch (error) {
    res.status(503).json({
      status: 'unhealthy',
      timestamp: new Date().toISOString(),
      service: 'ai-sre-demo-app',
      database: 'disconnected',
      error: error.message
    });
  }
});

/**
 * Get all users (working endpoint)
 */
router.get('/users', (req, res) => {
  try {
    const users = db.prepare('SELECT id, name, email, age, created_at FROM users').all();
    res.json({ users, count: users.length });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

/**
 * Get all products (working endpoint)
 */
router.get('/products', (req, res) => {
  try {
    const products = db.prepare('SELECT * FROM products').all();
    res.json({ products, count: products.length });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

/**
 * Get all orders (working endpoint)
 */
router.get('/orders', (req, res) => {
  try {
    const orders = db.prepare(`
      SELECT 
        o.id,
        o.quantity,
        o.total,
        o.created_at,
        u.name as user_name,
        p.name as product_name
      FROM orders o
      JOIN users u ON o.user_id = u.id
      JOIN products p ON o.product_id = p.id
      ORDER BY o.created_at DESC
    `).all();
    
    res.json({ orders, count: orders.length });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

export default router;
