# Demo UI Guide

This guide explains the interactive web UI for the AI SRE Demo Application.

## Overview

The demo UI provides a comprehensive interface for testing and monitoring the AI SRE Platform's autonomous remediation capabilities. It features real-time incident tracking, one-click bug triggering, and integration with the full incident management dashboard.

## Accessing the UI

Start the demo application and open your browser to:
```
http://localhost:3000
```

## Interface Layout

### Three Main Views

The UI uses a tab-based navigation system with three distinct views:

#### 1. 🐛 Trigger Bugs View (Default)

**Left Panel: Bug Trigger Cards**
- 5 intentional bug scenarios
- Each card includes:
  - Bug type badge (color-coded)
  - Descriptive title
  - Explanation of the bug
  - Trigger button
  - Response display area

**Right Panel: Recent Incidents**
- Shows 5 most recent incidents
- Auto-refreshes every 5 seconds
- Displays:
  - Incident ID
  - Status badge (color-coded)
  - Service name
  - Error message (truncated)
  - Links to PRs and observability platforms
  - Time ago timestamp

**Bottom: System Status**
- Demo service health
- Incident service connectivity
- Sentry integration status
- API endpoint information

#### 2. 📊 Incident Status View

- Complete list of all incidents
- Same rich formatting as Recent Incidents panel
- Scrollable list with all incident history
- Useful for reviewing past incidents

#### 3. 📈 Full Dashboard View

- Embedded iframe of the complete AI SRE Dashboard
- Full incident management interface
- Advanced filtering and sorting
- Detailed incident views

## Features

### Real-Time Updates

The UI automatically polls the Incident Service API every 5 seconds to fetch the latest incident data. A visual indicator (●) in the Recent Incidents panel shows when data is being refreshed.

**What Updates Automatically:**
- Incident status changes
- New incidents
- Pull request links (when created)
- Incident metadata

### Status Badges

Incidents display color-coded status badges:

| Status | Color | Meaning |
|--------|-------|---------|
| `pending` | Yellow | Incident received, waiting for processing |
| `workflow_triggered` | Blue | GitHub Actions workflow has been triggered |
| `in_progress` | Cyan | Remediation is actively running |
| `pr_created` | Green | Pull request has been created |
| `resolved` | Green | Incident has been resolved |
| `failed` | Red | Remediation failed |
| `no_fix_needed` | Gray | No automated fix was needed |

### Dynamic Links

The UI automatically renders links based on incident data:

**Pull Request Links** (🔗)
- Appears when `pull_request_url` is available
- Opens GitHub PR in new tab
- Allows immediate code review

**Sentry Issue Links** (🐛)
- Appears for incidents from Sentry provider
- Links to the original Sentry issue
- Provides full error context

**Datadog Links** (📊)
- Appears for incidents from Datadog provider
- Links to Datadog snapshot/alert
- Shows monitoring context

**PagerDuty Links** (📟)
- Appears for incidents from PagerDuty provider
- Links to PagerDuty incident
- Shows on-call context

## Bug Scenarios

### 1. Division by Zero
**Endpoint:** `GET /api/buggy/average-price`

Calculates average product price but doesn't handle empty product lists, resulting in division by zero.

**Expected Error:** `Infinity` or `NaN` in calculations

### 2. Null Pointer
**Endpoint:** `GET /api/buggy/user/999`

Attempts to access properties on a non-existent user object.

**Expected Error:** `TypeError: Cannot read property 'X' of undefined`

### 3. Array Processing
**Endpoint:** `POST /api/buggy/process-orders`

Off-by-one error in loop condition (`i <= array.length` instead of `i < array.length`).

**Expected Error:** `TypeError: Cannot read property 'productId' of undefined`

### 4. Type Coercion
**Endpoint:** `POST /api/buggy/calculate-discount`

No type validation on numeric inputs, causing string concatenation instead of math.

**Expected Error:** Incorrect calculation results

### 5. SQL Injection
**Endpoint:** `GET /api/buggy/search-users?name=Alice'`

String concatenation in SQL query instead of parameterized queries.

**Expected Error:** SQL syntax error

## Workflow

### Testing the Full Remediation Flow

1. **Trigger a Bug**
   - Click any bug trigger button
   - Observe the error response in the card

2. **Watch for Incident**
   - Within 2-5 seconds, incident appears in sidebar
   - Status starts as `pending`

3. **Monitor Status Changes**
   - Status updates to `workflow_triggered`
   - Then `in_progress` as Kiro CLI runs
   - Finally `pr_created` when fix is ready

4. **Review the Fix**
   - Click the "🔗 Pull Request" link
   - Review the automated fix
   - Check the post-mortem in PR description

5. **Check Observability Platform**
   - Click the provider link (Sentry, Datadog, etc.)
   - View the original error context
   - Correlate with the fix

## Configuration

### Incident Service URL

The UI automatically detects the environment:

- **Local Development:** `http://localhost:8080`
- **Docker Environment:** `http://incident-service:8080`

You can verify the configured URL in the System Status section.

### Refresh Interval

Default: 5 seconds

To modify, edit `demo-app/public/index.html`:
```javascript
const REFRESH_INTERVAL = 5000; // milliseconds
```

### Dashboard URL

Default: `http://localhost:3001`

To modify, edit the iframe src in `demo-app/public/index.html`:
```html
<iframe src="http://your-dashboard-url" ...>
```

## Troubleshooting

### Incidents Not Appearing

**Check:**
1. Incident Service is running: `curl http://localhost:8080/api/v1/health`
2. System Status shows "Connected" for Incident Service
3. Browser console for JavaScript errors (F12)
4. Network tab shows successful API calls

### Links Not Showing

**Possible Causes:**
- Incident status hasn't progressed to `pr_created` yet
- Provider data doesn't include URL fields
- Incident is still in `pending` or `workflow_triggered` state

**Solution:** Wait for workflow to complete, or check incident details in the Incident Status view.

### Dashboard Iframe Not Loading

**Check:**
1. Dashboard service is running: `curl http://localhost:3001`
2. Browser allows iframe embedding (check console for CSP errors)
3. CORS settings if dashboard is on different domain

### Auto-Refresh Not Working

**Check:**
1. Browser console for errors
2. Incident Service API is accessible
3. JavaScript is enabled in browser

## Development

### File Location

The demo UI is a single-page application located at:
```
demo-app/public/index.html
```

### Key Functions

**View Management:**
- `switchView(view)` - Changes between tabs
- `startAutoRefresh()` - Initializes polling

**Data Loading:**
- `loadRecentIncidents()` - Fetches recent incidents
- `loadAllIncidents()` - Fetches all incidents
- `renderIncidents(containerId, incidents)` - Renders incident list

**Health Checks:**
- `checkHealth()` - Checks demo service
- `checkIncidentService()` - Checks incident service

**Bug Triggers:**
- `triggerBug1()` through `triggerBug5()` - Trigger specific bugs

**Utilities:**
- `getTimeAgo(date)` - Converts timestamp to relative time
- `truncate(str, length)` - Truncates long strings
- `showResponse(elementId, data, isError)` - Displays API responses

### Adding Custom Styling

The UI uses inline CSS for simplicity. To customize:

1. Locate the `<style>` section in `index.html`
2. Modify CSS classes
3. Test in browser
4. Refresh to see changes

### Adding New Features

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed instructions on extending the demo UI.

## Best Practices

### For Demonstrations

1. **Start Clean:** Clear old incidents before demos
2. **Explain Flow:** Walk through the three views
3. **Show Real-Time:** Trigger bug and watch status update
4. **Review PR:** Show the automated fix quality
5. **Highlight Links:** Demonstrate integration with observability platforms

### For Development

1. **Test All Bugs:** Verify each scenario works
2. **Check Links:** Ensure all provider links render correctly
3. **Monitor Console:** Watch for JavaScript errors
4. **Verify Timing:** Confirm auto-refresh works
5. **Test Responsive:** Check on different screen sizes

## Support

For issues or questions:
- Check the [main README](../README.md)
- Review [CONTRIBUTING.md](../CONTRIBUTING.md)
- Open a GitHub issue

## License

MIT License - see [LICENSE](../LICENSE) for details.
