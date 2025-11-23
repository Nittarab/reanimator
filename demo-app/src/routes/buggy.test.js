import { describe, it, expect, beforeAll } from 'vitest';
import request from 'supertest';
import app from '../index.js';

describe('Buggy Endpoints', () => {
  describe('GET /api/buggy/average-price', () => {
    it('should calculate average price when products exist', async () => {
      const response = await request(app)
        .get('/api/buggy/average-price')
        .expect(200);

      expect(response.body).toHaveProperty('average');
      expect(response.body).toHaveProperty('count');
      expect(response.body).toHaveProperty('total');
      expect(response.body.count).toBeGreaterThan(0);
    });
  });

  describe('GET /api/buggy/user/:id', () => {
    it('should return user when user exists', async () => {
      const response = await request(app)
        .get('/api/buggy/user/1')
        .expect(200);

      expect(response.body).toHaveProperty('id');
      expect(response.body).toHaveProperty('name');
      expect(response.body).toHaveProperty('email');
    });

    it('should throw error when user does not exist (null pointer bug)', async () => {
      const response = await request(app)
        .get('/api/buggy/user/999')
        .expect(500);

      expect(response.body.error).toHaveProperty('message');
      expect(response.body.error.message).toContain('Cannot read properties of undefined');
    });
  });

  describe('POST /api/buggy/process-orders', () => {
    it('should throw error due to off-by-one bug', async () => {
      const response = await request(app)
        .post('/api/buggy/process-orders')
        .send({
          orders: [
            { userId: 1, productId: 1, quantity: 1 }
          ]
        })
        .expect(500);

      expect(response.body.error).toHaveProperty('message');
      expect(response.body.error.message).toContain('Cannot read properties of undefined');
    });
  });

  describe('POST /api/buggy/calculate-discount', () => {
    it('should handle numeric inputs correctly', async () => {
      const response = await request(app)
        .post('/api/buggy/calculate-discount')
        .send({
          price: 100,
          discountPercent: 20
        })
        .expect(200);

      expect(response.body).toHaveProperty('finalPrice');
      expect(response.body.finalPrice).toBe(80);
    });

    it('should demonstrate type coercion bug with string inputs', async () => {
      const response = await request(app)
        .post('/api/buggy/calculate-discount')
        .send({
          price: "100",
          discountPercent: "20"
        })
        .expect(200);

      // With proper validation, this should fail or convert types
      // But the buggy code will do string math
      expect(response.body).toHaveProperty('finalPrice');
    });
  });

  describe('GET /api/buggy/search-users', () => {
    it('should search users by name', async () => {
      const response = await request(app)
        .get('/api/buggy/search-users?name=Alice')
        .expect(200);

      expect(response.body).toHaveProperty('users');
      expect(Array.isArray(response.body.users)).toBe(true);
    });

    it('should handle SQL injection attempt', async () => {
      // This should ideally fail safely, but the buggy code might throw SQL error
      const response = await request(app)
        .get('/api/buggy/search-users?name=Alice\'');

      // Either returns error or handles it
      expect([200, 500]).toContain(response.status);
    });
  });
});
