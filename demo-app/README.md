# AI SRE Demo Application

A demonstration Node.js/Express application with intentionally buggy endpoints designed to showcase the AI SRE Platform's autonomous remediation capabilities.

## Overview

This demo application simulates a typical microservice with common programming errors that occur in production. When these bugs are triggered, they're reported to Sentry (if configured), which then triggers the AI SRE Platform's remediation workflow.

## Features

- **In-memory SQLite database** with sample data (users, products, orders)
- **Sentry integration** for error tracking and reporting
- **5 intentionally buggy endpoints** demonstrating common error patterns
- **Health check endpoint** for monitoring
- **Web UI** for easy bug triggering and testing

## Intentional Bugs

### 1. Division by Zero (`GET /api/buggy/average-price`)
- **Bug**: Divides by zero when product list is empty
- **Error Type**: `Infinity` or `NaN` in calculations
- **Trigger**: Call endpoint when no products exist

### 2. Null Pointer / Undefined Access (`GET /api/buggy/user/:id`)
- **Bug**: Accesses properties on undefined user object
- **Error Type**: `TypeError: Cannot read property 'X' of undefined`
- **Trigger**: Request non-existent user ID (e.g., 999)

### 3. Array Processing Error (`POST /api/buggy/process-orders`)
- **Bug**: Off-by-one error in loop (`i <= array.length` instead of `i < array.length`)
- **Error Type**: `TypeError: Cannot read property 'productId' of undefined`
- **Trigger**: Submit any order array

### 4. Type Coercion Error (`POST /api/buggy/calculate-discount`)
- **Bug**: No type validation on numeric inputs
- **Error Type**: String concatenation instead of mathematical operations
- **Trigger**: Send string values for price/discount

### 5. SQL Injection Vulnerability (`GET /api/buggy/search-users`)
- **Bug**: String concatenation in SQL query
- **Error Type**: SQL syntax errors with special characters
- **Trigger**: Search with special characters like `'` or `"`

## Installation

```bash
cd demo-app
npm install
```

## Configuration

### Environment Variables

Create a `.env` file in the demo-app directory:

```bash
# Server configuration
PORT=3000
NODE_ENV=development

# Sentry configuration (optional)
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
```

## Running the Application

### Development Mode (with auto-reload)
```bash
npm run dev
```

### Production Mode
```bash
npm start
```

The application will start on `http://localhost:3000`

## API Endpoints

### Healthy Endpoints

- `GET /api/health` - Health check endpoint
- `GET /api/users` - List all users
- `GET /api/products` - List all products
- `GET /api/orders` - List all orders

### Buggy Endpoints

- `GET /api/buggy/average-price` - Division by zero bug
- `GET /api/buggy/user/:id` - Null pointer bug
- `POST /api/buggy/process-orders` - Array processing bug
- `POST /api/buggy/calculate-discount` - Type coercion bug
- `GET /api/buggy/search-users?name=X` - SQL injection bug

## Web UI

Open `http://localhost:3000` in your browser to access the interactive UI where you can:
- **Trigger Bugs**: Click buttons to trigger each intentional bug
- **View Error Responses**: See real-time error responses from the demo service
- **Monitor Incidents**: View real-time incident status from the Incident Service
- **Track Remediation**: See links to pull requests and Sentry issues as they're created
- **Full Dashboard**: Embedded iframe view of the complete AI SRE Dashboard
- **System Status**: Monitor health of demo service and incident service

### UI Features

The demo UI has three main views:

1. **Trigger Bugs View** (Default)
   - 5 bug trigger cards with descriptions
   - Real-time incident status panel showing recent incidents
   - Links to PRs and observability platform issues
   - Auto-refreshes every 5 seconds

2. **Incident Status View**
   - Complete list of all incidents
   - Status badges (pending, workflow_triggered, pr_created, etc.)
   - Links to pull requests and Sentry issues
   - Incident metadata and timestamps

3. **Full Dashboard View**
   - Embedded iframe of the complete AI SRE Dashboard
   - Full incident management interface
   - Detailed incident views and filtering

📖 **For detailed UI documentation, see [UI-GUIDE.md](UI-GUIDE.md)**

## Testing with AI SRE Platform

### Quick Start

1. **Configure Sentry**: Set `SENTRY_DSN` environment variable
2. **Start the demo app**: `npm start`
3. **Trigger a bug**: Use the web UI or call buggy endpoints directly
4. **Watch Sentry**: Error should appear in your Sentry dashboard
5. **AI SRE Platform**: Should receive webhook and trigger remediation workflow
6. **Review PR**: Check GitHub for automatically generated fix pull request

### Testing the Remediation Workflow

The demo application includes a GitHub Actions workflow that demonstrates automated incident remediation:

**Quick Test** (2 minutes):
```bash
# Test workflow configuration locally
cd demo-app
./scripts/test-workflow-locally.sh

# Then trigger manually in GitHub:
# Actions → Demo Remediation Workflow → Run workflow
```

**Documentation**:
- 📖 [Quick Test Guide](.github/QUICK-TEST-GUIDE.md) - Fast 5-minute setup and test
- 📚 [Full Setup Guide](.github/DEMO-WORKFLOW-SETUP.md) - Detailed configuration and troubleshooting

**Required GitHub Secrets**:
- `SENTRY_AUTH_TOKEN` - Sentry API token
- `SENTRY_ORG` - Your Sentry organization slug
- `SENTRY_PROJECT` - Your Sentry project slug

See the guides above for step-by-step instructions.

## Database Schema

### Users Table
```sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  email TEXT NOT NULL UNIQUE,
  age INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Products Table
```sql
CREATE TABLE products (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  price REAL NOT NULL,
  stock INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Orders Table
```sql
CREATE TABLE orders (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  product_id INTEGER NOT NULL,
  quantity INTEGER NOT NULL,
  total REAL NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (product_id) REFERENCES products(id)
);
```

## Docker Support

Build and run with Docker:

```bash
docker build -t ai-sre-demo-app .
docker run -p 3000:3000 -e SENTRY_DSN=your-dsn ai-sre-demo-app
```

## Expected Fixes

The AI SRE Platform should generate fixes for:

1. **Division by Zero**: Add check for empty array before division
2. **Null Pointer**: Add null check before accessing user properties
3. **Array Processing**: Fix loop condition from `<=` to `<`
4. **Type Coercion**: Add input validation and type conversion
5. **SQL Injection**: Use parameterized queries instead of string concatenation

## Development

### Project Structure
```
demo-app/
├── src/
│   ├── db/
│   │   └── database.js       # SQLite database setup and seeding
│   ├── routes/
│   │   ├── buggy.js          # Buggy endpoints
│   │   └── healthy.js        # Working endpoints
│   └── index.js              # Express app setup
├── public/
│   └── index.html            # Web UI
├── package.json
└── README.md
```

### Adding New Bugs

1. Add endpoint to `src/routes/buggy.js`
2. Document the bug type and trigger condition
3. Update web UI in `public/index.html`
4. Add to this README

## License

MIT
