# API Request/Response Logging

## Overview

The frontend now includes comprehensive API logging that captures all requests and responses to the Go backend. This is extremely helpful for debugging authentication issues, API errors, and understanding the data flow.

## Features

✅ **Automatic logging** - All API requests are logged automatically
✅ **Color-coded console output** - Easy to distinguish requests, responses, and errors
✅ **Detailed information** - Headers, body, timing, status codes
✅ **Security** - Sensitive headers (auth tokens, cookies) are partially redacted
✅ **Performance tracking** - Shows request duration in milliseconds
✅ **Development-friendly** - Auto-enabled in development mode

## How to Use

### Automatic Logging

API logging is **automatically enabled** in development mode. Just open your browser console and you'll see:

- **→ GET /api/auth/profile/me** (blue) - Outgoing requests
- **← 200 OK (45ms)** (green) - Successful responses
- **← 403 Forbidden (120ms)** (red) - Error responses
- **✖ API ERROR Network Error** (red) - Network failures

### Manual Control

#### Enable Logging

```javascript
// In browser console
__apiLogger.enable()
```

#### Disable Logging

```javascript
// In browser console
__apiLogger.disable()
```

#### Check Status

```javascript
// The logger instance is available globally
console.log(__apiLogger.instance)
```

## What Gets Logged

### Request Information

For each outgoing request, you'll see:

- **Method** - GET, POST, PUT, DELETE
- **Full URL** - Complete endpoint URL
- **Headers** - All request headers (auth tokens partially redacted)
- **Body** - Request payload (for POST/PUT)
- **Timestamp** - When the request was sent

### Response Information

For each response, you'll see:

- **Status Code** - 200, 401, 403, 500, etc.
- **Status Text** - OK, Forbidden, etc.
- **Duration** - How long the request took in milliseconds
- **Response Headers** - All headers returned by server
- **Response Body** - The actual data returned
- **Timestamp** - When the response was received

### Error Information

For errors, you'll see:

- **Error Message** - What went wrong
- **Status Code** - HTTP status if available
- **Duration** - How long before it failed
- **Error Details** - Full error object for debugging
- **Stack Trace** - If available

## Example Output

### Successful Request

```
→ GET http://localhost:8080/api/auth/profile/me

Request Details:
┌────────────┬─────────────────────────────────────────────┐
│ Method     │ GET                                         │
│ URL        │ http://localhost:8080/api/auth/profile/me   │
│ Timestamp  │ 2025-12-23T18:30:00.000Z                    │
└────────────┴─────────────────────────────────────────────┘

Headers:
┌───────────────┬──────────────────────────────────────┐
│ authorization │ Bearer eyJhbGc...truncated...aMA9A    │
│ accept        │ */*                                  │
│ content-type  │ application/json                     │
└───────────────┴──────────────────────────────────────┘

← 200 OK (45ms)

Response Summary:
┌───────────┬──────────────────────────────────┐
│ Status    │ 200 OK                           │
│ Duration  │ 45ms                             │
│ Timestamp │ 2025-12-23T18:30:00.045Z         │
└───────────┴──────────────────────────────────┘

Response Body:
{
  member_id: "member-test-123",
  email: "user@example.com",
  organization_id: 1,
  permissions: ["org:view", "resource:create"]
}
```

### Error Response

```
→ GET http://localhost:8080/api/auth/profile/me

← 403 Forbidden (120ms)

Response Summary:
┌───────────┬──────────────────────────────────┐
│ Status    │ 403 Forbidden                    │
│ Duration  │ 120ms                            │
│ Timestamp │ 2025-12-23T18:30:00.120Z         │
└───────────┴──────────────────────────────────┘

Response Body:
{
  error: "organization not found",
  code: "FORBIDDEN",
  details: "Organization ID could not be resolved"
}

✖ API ERROR 403 (120ms)

Error Details:
┌──────────┬────────────────────────────────────────┐
│ Message  │ organization not found                 │
│ Status   │ 403                                    │
│ Duration │ 120ms                                  │
└──────────┴────────────────────────────────────────┘
```

## Security & Privacy

### Redacted Information

The logger automatically redacts sensitive information:

- **Authorization headers** - Only shows first 20 and last 20 characters
- **Cookies** - Shown as `[REDACTED - Contains session cookies]`
- **API keys** - Shown as `[REDACTED]`

### Example

```
Headers:
┌───────────────┬────────────────────────────────────────────────┐
│ authorization │ Bearer eyJhbGciOiJSUzI1Ni...MA9A (partial)    │
│ cookie        │ [REDACTED - Contains session cookies]         │
└───────────────┴────────────────────────────────────────────────┘
```

## Debugging 403 Errors

When you see a 403 Forbidden error, the logger helps you:

1. **Check the JWT token** - Verify it's being sent correctly
2. **Check the organization_id** - See what org ID is in the token
3. **Check response details** - See why the backend rejected it
4. **Measure timing** - See if requests are timing out

### Common 403 Causes

Based on the logs, you can identify:

- **Missing authorization header** - Token not attached
- **Expired token** - Token has expired
- **Organization mismatch** - Org ID in token doesn't match database
- **Insufficient permissions** - User doesn't have required role
- **Account not found** - User account not in database

## Production Usage

The logger is **automatically disabled** in production builds to:
- Avoid performance overhead
- Prevent leaking sensitive data to console
- Reduce console noise

To enable logging in production (for debugging):

```javascript
localStorage.setItem('API_LOGGER_ENABLED', 'true')
// Then reload the page
```

## Files

- `lib/utils/api-logger.ts` - Logger implementation
- `lib/api/api/client/api-client.ts` - Integration point
- `docs/API-LOGGING.md` - This documentation

## Troubleshooting

### Logs not showing?

1. Check you're in development mode (`process.env.NODE_ENV === 'development'`)
2. Try manually enabling: `__apiLogger.enable()`
3. Refresh the page after enabling
4. Check browser console is set to show all levels (not just errors)

### Too much noise?

```javascript
// Disable logging
__apiLogger.disable()
```

### Need to see specific requests?

The logs are collapsible in the console. Click to expand only the requests you care about.

## Benefits

✅ **Faster debugging** - See exactly what's being sent/received
✅ **Better error messages** - Understand why requests fail
✅ **Performance insights** - Identify slow API calls
✅ **Security auditing** - Verify auth headers are correct
✅ **Integration testing** - Validate request/response formats

---

**Pro Tip:** Keep the console open while developing to catch API issues immediately!
