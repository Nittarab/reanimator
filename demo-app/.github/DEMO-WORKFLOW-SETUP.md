# Demo Workflow Setup Guide

This guide walks you through setting up and testing the AI SRE Platform's automated remediation workflow for the demo application.

## Overview

The demo remediation workflow (`.github/workflows/demo-remediate.yml`) demonstrates the complete end-to-end flow:

1. **Incident Trigger**: An error occurs in the demo app and is reported to Sentry
2. **Workflow Dispatch**: The Incident Service triggers the GitHub Actions workflow
3. **Kiro CLI Analysis**: The remediation action uses Kiro CLI with MCP to analyze the error
4. **Fix Generation**: Kiro generates a code fix based on the root cause analysis
5. **Pull Request**: A PR is created with the fix and a detailed post-mortem
6. **Status Update**: The Incident Service is notified of the remediation status

## Prerequisites

Before setting up the workflow, ensure you have:

- [ ] A GitHub repository with the AI SRE Platform code
- [ ] A Sentry account and project for error tracking
- [ ] The demo application configured with Sentry DSN
- [ ] GitHub Actions enabled in your repository

## Step 1: Configure Sentry

### 1.1 Create Sentry Project

1. Log in to [Sentry](https://sentry.io)
2. Create a new project or use an existing one
3. Select "Node.js" as the platform
4. Note your **Project Slug** (e.g., `demo-app`)
5. Note your **Organization Slug** (e.g., `my-company`)

### 1.2 Generate Sentry Auth Token

1. Go to: **Settings → Account → API → Auth Tokens**
2. Click **"Create New Token"**
3. Name: `AI SRE Platform - Demo Workflow`
4. Scopes: Select the following:
   - `project:read` - Read project data
   - `event:read` - Read error events
   - `org:read` - Read organization data
5. Click **"Create Token"**
6. **Copy the token** - you won't be able to see it again!

### 1.3 Configure Demo App with Sentry DSN

1. Get your Sentry DSN from: **Settings → Projects → [Your Project] → Client Keys (DSN)**
2. Update `demo-app/.env`:
   ```bash
   SENTRY_DSN=https://your-key@sentry.io/your-project-id
   ```
3. Restart the demo app to apply changes

## Step 2: Configure GitHub Secrets

### 2.1 Add Repository Secrets

1. Go to your GitHub repository
2. Navigate to: **Settings → Secrets and variables → Actions**
3. Click **"New repository secret"**
4. Add the following secrets:

#### SENTRY_AUTH_TOKEN
- **Name**: `SENTRY_AUTH_TOKEN`
- **Value**: The auth token you created in Step 1.2
- Click **"Add secret"**

#### SENTRY_ORG
- **Name**: `SENTRY_ORG`
- **Value**: Your Sentry organization slug (e.g., `my-company`)
- Click **"Add secret"**

#### SENTRY_PROJECT
- **Name**: `SENTRY_PROJECT`
- **Value**: Your Sentry project slug (e.g., `demo-app`)
- Click **"Add secret"**

### 2.2 Add Repository Variables (Optional)

If you want the workflow to report status back to the Incident Service:

1. Navigate to: **Settings → Secrets and variables → Actions → Variables**
2. Click **"New repository variable"**
3. Add:
   - **Name**: `INCIDENT_SERVICE_URL`
   - **Value**: `http://localhost:8080` (or your Incident Service URL)
   - Click **"Add variable"**

## Step 3: Verify Workflow Configuration

### 3.1 Check Workflow File

Verify that `.github/workflows/demo-remediate.yml` exists and contains:

```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  # Sentry MCP Server credentials
  SENTRY_AUTH_TOKEN: ${{ secrets.SENTRY_AUTH_TOKEN }}
  SENTRY_ORG: ${{ secrets.SENTRY_ORG }}
  SENTRY_PROJECT: ${{ secrets.SENTRY_PROJECT }}
```

### 3.2 Check MCP Configuration

Verify that `demo-app/.kiro/settings/mcp.json` exists and contains:

```json
{
  "mcpServers": {
    "sentry": {
      "command": "npx",
      "args": ["-y", "@sentry/mcp-server"],
      "env": {
        "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}",
        "SENTRY_ORG": "${SENTRY_ORG}",
        "SENTRY_PROJECT": "${SENTRY_PROJECT}"
      }
    }
  }
}
```

### 3.3 Check Remediation Strategies

Verify that remediation strategies exist in `demo-app/.kiro/specs/demo-fixes/`:
- `division-by-zero.md`
- `null-pointer.md`
- `array-processing.md`

## Step 4: Test the Workflow

### 4.1 Manual Workflow Trigger

Test the workflow manually before setting up the full integration:

1. Go to: **Actions → Demo Remediation Workflow**
2. Click **"Run workflow"**
3. Fill in the form:
   - **Incident ID**: `test-001`
   - **Error message**: `TypeError: Cannot read property 'name' of undefined`
   - **Stack trace**: `at getUserName (demo-app/src/routes/buggy.js:42:15)`
   - **Service name**: `demo-app`
   - **Timestamp**: `2024-01-15T10:30:00Z`
4. Click **"Run workflow"**

### 4.2 Monitor Workflow Execution

1. Click on the running workflow to view logs
2. Watch each step execute:
   - ✅ Checkout repository
   - ✅ Run AI SRE Remediation
     - Install Kiro CLI
     - Configure MCP servers
     - Create incident context file
     - Run Kiro CLI for remediation
     - Create branch
     - Commit changes
     - Push branch
     - Generate post-mortem
     - Create pull request
     - Send notifications
   - ✅ Output results

### 4.3 Verify Pull Request

1. Go to: **Pull requests**
2. Find the PR created by the workflow (e.g., `fix/incident-test-001`)
3. Review the PR:
   - **Title**: Should reference the incident ID
   - **Description**: Should contain a detailed post-mortem
   - **Changes**: Should contain the code fix
   - **Commits**: Should have a descriptive commit message

### 4.4 Expected Output

The workflow should output:
```
Status: success
PR URL: https://github.com/your-org/your-repo/pull/123
Diagnosis: [Root cause analysis from Kiro CLI]
```

## Step 5: Test End-to-End Flow

### 5.1 Start the Demo App

```bash
cd demo-app
npm install
npm start
```

The app should start on `http://localhost:3000`

### 5.2 Trigger a Bug

Open the demo app UI at `http://localhost:3000` and click one of the bug trigger buttons:

- **Division by Zero**: Triggers when calculating average price with no products
- **Null Pointer**: Triggers when accessing non-existent user
- **Array Processing**: Triggers when processing orders with off-by-one error

### 5.3 Verify Sentry Receives Error

1. Go to your Sentry dashboard
2. Navigate to: **Issues**
3. You should see the error appear within seconds
4. Click on the issue to view details

### 5.4 Manually Trigger Workflow with Sentry Data

1. Copy the error details from Sentry:
   - Error message
   - Stack trace
   - Timestamp
2. Go to: **Actions → Demo Remediation Workflow → Run workflow**
3. Paste the Sentry data into the form
4. Run the workflow

### 5.5 Verify MCP Integration

In the workflow logs, look for:
```
Configuring MCP servers
Configured 1 MCP server(s): sentry
```

This confirms that Kiro CLI can query Sentry for additional context.

## Step 6: Integrate with Incident Service (Optional)

To complete the full automation loop:

### 6.1 Configure Sentry Webhook

1. In Sentry, go to: **Settings → Integrations → Webhooks**
2. Add a new webhook:
   - **URL**: `http://your-incident-service:8080/api/v1/webhooks/incidents?provider=sentry`
   - **Events**: Select "Issue" events
3. Save the webhook

### 6.2 Configure Incident Service

Update `incident-service/config.yaml`:

```yaml
service_mappings:
  - service_name: demo-app
    repository: your-org/your-repo
    branch: main
```

### 6.3 Test Full Flow

1. Trigger a bug in the demo app
2. Sentry receives the error
3. Sentry sends webhook to Incident Service
4. Incident Service triggers the GitHub Actions workflow
5. Workflow creates a PR with the fix
6. Workflow reports status back to Incident Service

## Troubleshooting

### Workflow Fails at "Install Kiro CLI"

**Problem**: Kiro CLI installation fails

**Solution**:
- Check that the `kiro_version` input is valid
- Verify network connectivity in GitHub Actions runner
- Check Kiro CLI installation logs for specific errors

### Workflow Fails at "Configure MCP servers"

**Problem**: MCP configuration fails

**Solutions**:
- Verify all three Sentry secrets are set correctly
- Check secret names match exactly (case-sensitive)
- Verify Sentry auth token has correct scopes
- Test Sentry API access manually:
  ```bash
  curl -H "Authorization: Bearer YOUR_TOKEN" \
    https://sentry.io/api/0/organizations/YOUR_ORG/
  ```

### No Changes Detected

**Problem**: Workflow completes but says "No changes detected"

**Solutions**:
- Verify the error is actually in the repository code
- Check that Kiro CLI can locate the problematic file
- Review Kiro CLI logs for analysis details
- Ensure remediation strategies are present in `.kiro/specs/demo-fixes/`

### Pull Request Not Created

**Problem**: Workflow fails at "Create pull request"

**Solutions**:
- Verify `GITHUB_TOKEN` has correct permissions
- Check workflow permissions in `.github/workflows/demo-remediate.yml`:
  ```yaml
  permissions:
    contents: write
    pull-requests: write
  ```
- Ensure branch doesn't already exist
- Check GitHub API rate limits

### MCP Server Not Working

**Problem**: Kiro CLI doesn't query Sentry

**Solutions**:
- Verify MCP configuration in workflow logs
- Check that `@sentry/mcp-server` package exists
- Test MCP server locally:
  ```bash
  export SENTRY_AUTH_TOKEN=your-token
  export SENTRY_ORG=your-org
  export SENTRY_PROJECT=your-project
  npx -y @sentry/mcp-server
  ```
- Review Kiro CLI logs for MCP query attempts

### Status Not Reported to Incident Service

**Problem**: Incident Service doesn't receive status updates

**Solutions**:
- Verify `INCIDENT_SERVICE_URL` variable is set
- Check Incident Service is running and accessible
- Review workflow logs for status reporting errors
- Verify Incident Service webhook endpoint is working:
  ```bash
  curl -X POST http://localhost:8080/api/v1/webhooks/workflow-status \
    -H "Content-Type: application/json" \
    -d '{"incident_id":"test","status":"success"}'
  ```

## Security Best Practices

### Protect Secrets

- ✅ **DO**: Store credentials in GitHub Secrets
- ✅ **DO**: Use environment variable substitution (`${VAR}`)
- ❌ **DON'T**: Commit secrets to the repository
- ❌ **DON'T**: Log secret values in workflow output

### Least Privilege

- Use Sentry tokens with minimal required scopes
- Create separate tokens for CI/CD vs. development
- Rotate tokens regularly
- Revoke tokens when no longer needed

### Monitor Usage

- Review Sentry audit logs for API usage
- Set up alerts for unusual activity
- Monitor GitHub Actions usage and costs

## Next Steps

After successfully testing the demo workflow:

1. **Add More Bugs**: Create additional buggy endpoints in the demo app
2. **Write Remediation Strategies**: Add corresponding strategy documents
3. **Test Different Scenarios**: Try various error types and edge cases
4. **Integrate with CI/CD**: Set up automatic workflow triggers
5. **Monitor Metrics**: Track success rate, MTTR, and other KPIs

## Resources

- [Demo App README](../README.md)
- [Remediation Action README](../../remediation-action/README.md)
- [Sentry MCP Server Documentation](https://docs.sentry.io/mcp)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Kiro CLI Documentation](https://docs.kiro.ai)

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review workflow logs in GitHub Actions
3. Check Sentry error logs
4. Review Incident Service logs (if applicable)
5. Open an issue in the repository with:
   - Workflow run URL
   - Error messages
   - Steps to reproduce
