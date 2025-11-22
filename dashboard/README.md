# AI SRE Platform Dashboard

React-based web dashboard for monitoring incidents and remediation progress.

## Technology Stack

- **React 18** with TypeScript
- **Vite** for build tooling
- **TanStack Query** (React Query) for API state management
- **React Router** for routing
- **shadcn/ui** for UI components (Tailwind CSS based)
- **Vitest** for testing

## Getting Started

### Prerequisites

- Node.js 20+
- npm or yarn

### Installation

```bash
npm install
```

### Development

```bash
# Start development server
npm run dev

# The dashboard will be available at http://localhost:3000
# API requests are proxied to http://localhost:8080
```

### Building

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

### Testing

```bash
# Run tests
npm test

# Run tests in watch mode
npm run test:watch
```

## Project Structure

```
dashboard/
├── src/
│   ├── api/              # API client and types
│   ├── components/       # React components
│   │   ├── ui/          # shadcn/ui components
│   │   └── Layout.tsx   # Main layout component
│   ├── pages/           # Page components
│   ├── lib/             # Utility functions
│   ├── test/            # Test setup
│   ├── App.tsx          # Main app component
│   ├── main.tsx         # Entry point
│   └── index.css        # Global styles
├── public/              # Static assets
└── index.html           # HTML template
```

## Configuration

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Configure the API base URL:

```
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Features

- Real-time incident monitoring
- Incident filtering and search
- Incident detail view with timeline
- Manual remediation triggering
- Configuration management
- Responsive design

## API Integration

The dashboard communicates with the Incident Service API:

- `GET /api/v1/incidents` - List incidents
- `GET /api/v1/incidents/:id` - Get incident details
- `POST /api/v1/incidents/:id/trigger` - Trigger remediation
- `GET /api/v1/incidents/:id/events` - Get incident events

## Development Notes

- The dashboard uses shadcn/ui components which are based on Radix UI and Tailwind CSS
- API state is managed with TanStack Query for automatic caching and refetching
- The app uses React Router for client-side routing
- Tests use Vitest with React Testing Library and happy-dom
