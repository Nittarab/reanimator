# Quick Test Guide - Demo Remediation Workflow

This is a condensed guide for quickly testing the demo remediation workflow. For detailed setup instructions, see [DEMO-WORKFLOW-SETUP.md](DEMO-WORKFLOW-SETUP.md).

## Prerequisites Checklist

- [ ] Sentry account with a project created
- [ ] Sentry auth token with `project:read` and `event:read` scopes
- [ ] GitHub repository with Actions enabled
- [ ] Demo app configured with Sentry DSN

## Quick Setup (5 minutes)

### 1. Add GitHub Secrets

Go to: **Settings → Secrets and variables → Actions → New repository secret**

Add these three secrets:

| Secret Name | Value | Where to Find |
|-------------|-------|---------------|
| `SENTRY_AUTH_TOKEN` | Your Sentry API token | Sentry → Settings → Account → API → Auth Tokens |
| `SENTRY_ORG` | Your org slug (e.g., `my-company`) | Sentry URL: `sentry.io/organizations/{org-slug}/` |
| `SENTRY_PROJECT` | Your project slug (e.g., `demo-app`) | Sentry URL: `.../projects/{project-slug}/` |

### 2. Verify Files Exist

Check these files are present:
- ✅ `.github/workflows/demo-remediate.yml`
- ✅ `demo-app/.kiro/settings/mcp.json`
- ✅ `demo-app/.kiro/specs/demo-fixes/division-by-zero.md`
- ✅ `demo-app/.kiro/specs/demo-fixes/null-pointer.md`
- ✅ `demo-app/.kiro/specs/demo-fixes/array-processing.md`

## Quick Test (2 minutes)

### Option A: Manual Trigger

1. Go to: **Actions → Demo Remediation Workflow → Run workflow**
2. Fill in:
   ```
   Incident ID: test-null-pointer-001
   Error message: TypeError: Cannot read property 'name' of undefined
   Stack trace: at getUserName (demo-app/src/routes/buggy.js:42:15)
   Service name: demo-app
   Timestamp: 2024-01-15T10:30:00Z
   ```
3. Click **"Run workflow"**
4. Wait ~2-3 minutes for completion
5. Check for new PR in **Pull requests** tab

### Option B: Trigger from Demo App

1. Start demo app:
   ```bash
   cd demo-app
   npm install
   npm start
   ```
2. Open `http://localhost:3000`
3. Click **"Trigger Null Pointer Bug"**
4. Go to Sentry and copy error details
5. Manually trigger workflow with Sentry data (see Option A)

## Expected Results

### Successful Workflow

You should see:
- ✅ All workflow steps complete successfully
- ✅ New branch created: `fix/incident-test-null-pointer-001`
- ✅ New PR created with:
  - Title: `Fix: Incident test-null-pointer-001 - demo-app`
  - Description: Detailed post-mortem
  - Changes: Code fix for the null pointer bug
- ✅ Workflow output:
  ```
  Status: success
  PR URL: https://github.com/your-org/your-repo/pull/123
  Diagnosis: [Root cause analysis]
  ```

### Workflow Logs Should Show

```
Installing Kiro CLI
✓ Kiro CLI installed

Configuring MCP servers
✓ Configured 1 MCP server(s): sentry

Creating incident context file
✓ Incident context file created

Running Kiro CLI for remediation
✓ Kiro CLI completed successfully

Checking for changes
✓ Changes detected

Creating branch
✓ Branch created: fix/incident-test-null-pointer-001

Committing changes
✓ Changes committed

Pushing branch
✓ Branch pushed

Generating post-mortem
✓ Post-mortem generated

Creating pull request
✓ Pull request created: https://github.com/.../pull/123

Sending notifications
✓ Notifications sent (or skipped if not configured)
```

## Common Issues & Quick Fixes

| Issue | Quick Fix |
|-------|-----------|
| "MCP configuration failed" | Verify all 3 Sentry secrets are set correctly |
| "No changes detected" | Check that the error is in the demo-app code |
| "PR creation failed" | Verify workflow has `contents: write` and `pull-requests: write` permissions |
| "Kiro CLI installation failed" | Check GitHub Actions runner has internet access |

## Test Different Bugs

Try these test cases:

### 1. Division by Zero
```
Incident ID: test-division-001
Error message: Invalid average calculation: Infinity
Stack trace: at calculateAverage (demo-app/src/routes/buggy.js:15:20)
Service name: demo-app
```

### 2. Null Pointer
```
Incident ID: test-null-pointer-001
Error message: TypeError: Cannot read property 'name' of undefined
Stack trace: at getUserName (demo-app/src/routes/buggy.js:42:15)
Service name: demo-app
```

### 3. Array Processing
```
Incident ID: test-array-001
Error message: TypeError: Cannot read property 'productId' of undefined
Stack trace: at processOrders (demo-app/src/routes/buggy.js:78:25)
Service name: demo-app
```

## Verify MCP Integration

In the workflow logs, look for:
```
Configuring MCP servers
Configured 1 MCP server(s): sentry
MCP configuration written to: /home/runner/work/.../demo-app/.kiro/settings/mcp.json
```

This confirms Kiro CLI can query Sentry for additional context during remediation.

## Next Steps

After successful test:
1. ✅ Review the generated PR and post-mortem
2. ✅ Merge the PR to apply the fix
3. ✅ Test the fixed endpoint to verify it works
4. ✅ Try triggering other bugs
5. ✅ Set up full integration with Incident Service (optional)

## Need Help?

- **Detailed Setup**: See [DEMO-WORKFLOW-SETUP.md](DEMO-WORKFLOW-SETUP.md)
- **Troubleshooting**: Check the troubleshooting section in the detailed guide
- **Demo App**: See [demo-app/README.md](../README.md)
- **Remediation Action**: See [remediation-action/README.md](../../remediation-action/README.md)
