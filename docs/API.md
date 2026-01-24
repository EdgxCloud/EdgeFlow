# EdgeFlow API Documentation

## Overview

EdgeFlow provides a RESTful API for managing flows, nodes, and executions. All endpoints return JSON responses.

**Base URL:** `http://localhost:8080/api`

## Authentication

EdgeFlow supports two authentication methods:

### 1. JWT Token (Recommended for web applications)

```bash
# Login to get token (implement /api/auth/login endpoint)
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}'

# Use token in subsequent requests
curl -X GET http://localhost:8080/api/flows \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 2. API Key (Recommended for integrations)

```bash
# Generate API key (implement /api/apikeys endpoint)
curl -X POST http://localhost:8080/api/apikeys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"name": "My Integration", "permissions": ["flows:read", "flows:write"]}'

# Use API key in requests
curl -X GET http://localhost:8080/api/flows \
  -H "X-API-Key: YOUR_API_KEY"
```

## Endpoints

### Health & Metrics

#### GET /health
Health check endpoint

**Response:**
```json
{
  "status": "healthy",
  "checks": [
    {
      "name": "database",
      "status": "healthy",
      "message": "Database is healthy",
      "last_check": "2026-01-20T10:00:00Z"
    }
  ],
  "timestamp": "2026-01-20T10:00:00Z"
}
```

#### GET /metrics
Prometheus metrics endpoint

**Response:** Prometheus format metrics

#### GET /metrics/json
JSON format metrics

**Response:**
```json
{
  "flows": {
    "total": 10,
    "running": 3,
    "stopped": 7
  },
  "executions": {
    "total": 1000,
    "failed": 10,
    "success_rate": 99.0
  },
  "system": {
    "uptime_seconds": 86400,
    "memory_used_mb": 128,
    "goroutines": 50
  }
}
```

### Flows

#### GET /flows
List all flows

**Response:**
```json
{
  "flows": [
    {
      "id": "flow-123",
      "name": "My Flow",
      "description": "Example flow",
      "status": "running",
      "created_at": "2026-01-20T10:00:00Z",
      "updated_at": "2026-01-20T10:00:00Z"
    }
  ]
}
```

#### POST /flows
Create a new flow

**Request:**
```json
{
  "name": "My Flow",
  "description": "Example flow"
}
```

**Response:**
```json
{
  "id": "flow-123",
  "name": "My Flow",
  "description": "Example flow",
  "status": "stopped",
  "nodes": [],
  "connections": []
}
```

#### GET /flows/:id
Get flow by ID

#### PUT /flows/:id
Update flow

**Request:**
```json
{
  "name": "Updated Flow",
  "description": "Updated description"
}
```

#### DELETE /flows/:id
Delete flow

#### POST /flows/:id/start
Start a flow

**Response:**
```json
{
  "status": "running",
  "message": "Flow started successfully"
}
```

#### POST /flows/:id/stop
Stop a flow

#### POST /flows/:id/deploy
Deploy flow configuration

**Request:**
```json
{
  "nodes": [
    {
      "id": "node-1",
      "type": "inject",
      "name": "Inject",
      "config": {
        "payload": "Hello World",
        "interval": 1000
      }
    },
    {
      "id": "node-2",
      "type": "debug",
      "name": "Debug",
      "config": {}
    }
  ],
  "connections": [
    {
      "source": "node-1",
      "target": "node-2"
    }
  ]
}
```

### Nodes

#### GET /flows/:flowId/nodes
List nodes in a flow

#### POST /flows/:flowId/nodes
Add node to flow

**Request:**
```json
{
  "type": "inject",
  "name": "Inject",
  "x": 100,
  "y": 100,
  "config": {
    "payload": "Hello",
    "interval": 1000
  }
}
```

#### PUT /flows/:flowId/nodes/:nodeId
Update node

#### DELETE /flows/:flowId/nodes/:nodeId
Remove node from flow

### Connections

#### GET /flows/:flowId/connections
List connections in a flow

#### POST /flows/:flowId/connections
Create connection

**Request:**
```json
{
  "source": "node-1",
  "sourceOutput": 0,
  "target": "node-2",
  "targetInput": 0
}
```

#### DELETE /flows/:flowId/connections/:connectionId
Delete connection

### Node Types

#### GET /node-types
List all available node types

**Response:**
```json
{
  "nodeTypes": [
    {
      "type": "inject",
      "name": "Inject",
      "category": "core",
      "description": "ارسال پیام دوره‌ای",
      "icon": "play",
      "inputs": 0,
      "outputs": 1,
      "config": {
        "payload": "",
        "interval": 1000,
        "repeat": true
      }
    }
  ]
}
```

#### GET /node-types/:type
Get node type details

### Executions

#### GET /executions
List execution history

**Query Parameters:**
- `flow_id` - Filter by flow ID
- `status` - Filter by status (success, failed)
- `limit` - Limit results (default: 100)
- `offset` - Offset for pagination

**Response:**
```json
{
  "executions": [
    {
      "id": "exec-123",
      "flow_id": "flow-123",
      "status": "success",
      "started_at": "2026-01-20T10:00:00Z",
      "completed_at": "2026-01-20T10:00:05Z",
      "duration_ms": 5000
    }
  ],
  "total": 1000
}
```

## Error Responses

All error responses follow this format:

```json
{
  "error": "Error message description"
}
```

### HTTP Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## WebSocket API

EdgeFlow provides real-time updates via WebSocket.

**Endpoint:** `ws://localhost:8080/ws`

### Messages

#### Subscribe to flow updates
```json
{
  "type": "subscribe",
  "flow_id": "flow-123"
}
```

#### Flow status update
```json
{
  "type": "flow_status",
  "flow_id": "flow-123",
  "status": "running"
}
```

#### Node execution
```json
{
  "type": "node_execution",
  "flow_id": "flow-123",
  "node_id": "node-1",
  "message": {
    "payload": "Hello World"
  }
}
```

## Rate Limiting

API requests are rate limited to:
- **Authenticated users:** 1000 requests per hour
- **API keys:** 5000 requests per hour

Rate limit headers:
- `X-RateLimit-Limit` - Request limit
- `X-RateLimit-Remaining` - Remaining requests
- `X-RateLimit-Reset` - Reset time (Unix timestamp)

## Examples

### Complete Flow Creation Example

```bash
# 1. Create flow
FLOW_ID=$(curl -X POST http://localhost:8080/api/flows \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "name": "Temperature Monitor",
    "description": "Monitor temperature sensor"
  }' | jq -r '.id')

# 2. Add nodes
curl -X POST http://localhost:8080/api/flows/$FLOW_ID/nodes \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "type": "inject",
    "name": "Trigger",
    "config": {"interval": 5000}
  }'

# 3. Deploy flow
curl -X POST http://localhost:8080/api/flows/$FLOW_ID/deploy \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d @flow-config.json

# 4. Start flow
curl -X POST http://localhost:8080/api/flows/$FLOW_ID/start \
  -H "X-API-Key: YOUR_API_KEY"
```

## SDKs

### JavaScript/TypeScript
```javascript
import { EdgeFlowClient } from '@edgeflow/client';

const client = new EdgeFlowClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'YOUR_API_KEY'
});

// Create flow
const flow = await client.flows.create({
  name: 'My Flow',
  description: 'Example'
});

// Start flow
await client.flows.start(flow.id);
```

### Python
```python
from edgeflow import EdgeFlowClient

client = EdgeFlowClient(
    base_url='http://localhost:8080',
    api_key='YOUR_API_KEY'
)

# Create flow
flow = client.flows.create(
    name='My Flow',
    description='Example'
)

# Start flow
client.flows.start(flow.id)
```

## Support

For issues and questions:
- GitHub: https://github.com/edgeflow/edgeflow/issues
- Documentation: https://docs.edgeflow.io
