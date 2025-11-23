import express from 'express';
import * as Sentry from '@sentry/node';
import buggyRoutes from './routes/buggy.js';
import healthyRoutes from './routes/healthy.js';

const app = express();
const PORT = process.env.PORT || 3000;

// Initialize Sentry for error tracking
Sentry.init({
  dsn: process.env.SENTRY_DSN || '',
  environment: process.env.NODE_ENV || 'development',
  tracesSampleRate: 1.0,
  // Enable Sentry only if DSN is provided
  enabled: !!process.env.SENTRY_DSN,
});

// Sentry request handler must be the first middleware
app.use(Sentry.Handlers.requestHandler());

// Body parsing middleware
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Logging middleware
app.use((req, res, next) => {
  const timestamp = new Date().toISOString();
  console.log(`[${timestamp}] ${req.method} ${req.path}`);
  next();
});

// Serve static files from public directory
app.use(express.static('public'));

// Mount routes
app.use('/api', healthyRoutes);
app.use('/api/buggy', buggyRoutes);

// Root endpoint
app.get('/', (req, res) => {
  res.json({
    service: 'AI SRE Demo Application',
    version: '1.0.0',
    description: 'Demo application with intentionally buggy endpoints for testing AI SRE Platform',
    endpoints: {
      health: '/api/health',
      users: '/api/users',
      products: '/api/products',
      orders: '/api/orders',
      buggy: {
        averagePrice: '/api/buggy/average-price',
        user: '/api/buggy/user/:id',
        processOrders: '/api/buggy/process-orders',
        calculateDiscount: '/api/buggy/calculate-discount',
        searchUsers: '/api/buggy/search-users'
      }
    }
  });
});

// Sentry error handler must be before other error handlers
app.use(Sentry.Handlers.errorHandler());

// Global error handler
app.use((err, req, res, next) => {
  console.error('Error:', err);
  
  res.status(err.status || 500).json({
    error: {
      message: err.message || 'Internal Server Error',
      status: err.status || 500,
      timestamp: new Date().toISOString(),
      path: req.path
    }
  });
});

// 404 handler
app.use((req, res) => {
  res.status(404).json({
    error: {
      message: 'Not Found',
      status: 404,
      timestamp: new Date().toISOString(),
      path: req.path
    }
  });
});

// Start server only if not in test mode
if (process.env.NODE_ENV !== 'test') {
  app.listen(PORT, () => {
    console.log(`🚀 AI SRE Demo App running on port ${PORT}`);
    console.log(`📊 Health check: http://localhost:${PORT}/api/health`);
    console.log(`🐛 Buggy endpoints available at: http://localhost:${PORT}/api/buggy/*`);
    console.log(`📝 Sentry integration: ${process.env.SENTRY_DSN ? 'enabled' : 'disabled (set SENTRY_DSN to enable)'}`);
  });
}

export default app;
