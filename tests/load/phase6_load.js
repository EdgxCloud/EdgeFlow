import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const apiErrors = new Counter('api_errors');
const apiSuccessRate = new Rate('api_success_rate');
const flowExecutionTime = new Trend('flow_execution_time');
const websocketLatency = new Trend('websocket_latency');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 20 },   // Ramp up to 20 users
    { duration: '1m', target: 50 },    // Ramp up to 50 users
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '2m', target: 100 },   // Stay at 100 users
    { duration: '1m', target: 0 },     // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests should complete under 500ms
    'api_success_rate': ['rate>0.95'],   // 95% success rate
    'api_errors': ['count<100'],         // Less than 100 errors total
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_KEY = __ENV.API_KEY || 'test-api-key-12345';

// Setup function runs once per VU at the beginning
export function setup() {
  // Create test flows for load testing
  const flows = [];

  for (let i = 0; i < 10; i++) {
    const flow = {
      name: `LoadTest-Flow-${i}`,
      description: `Load testing flow ${i}`,
      nodes: [
        {
          id: `inject-${i}`,
          type: 'inject',
          name: 'Inject',
          config: {
            interval: 1000,
            payload: `test-${i}`,
          },
        },
        {
          id: `debug-${i}`,
          type: 'debug',
          name: 'Debug',
          config: {
            complete: false,
          },
        },
      ],
      connections: [
        {
          id: `conn-${i}`,
          source: `inject-${i}`,
          target: `debug-${i}`,
        },
      ],
    };

    const response = http.post(
      `${BASE_URL}/api/flows`,
      JSON.stringify(flow),
      {
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': API_KEY,
        },
      }
    );

    if (response.status === 201 || response.status === 200) {
      flows.push(JSON.parse(response.body));
    }
  }

  return { flows };
}

// Main test scenario
export default function (data) {
  group('API Health Checks', () => {
    const healthResponse = http.get(`${BASE_URL}/api/health`);
    check(healthResponse, {
      'health check status is 200': (r) => r.status === 200,
      'health check has status field': (r) => JSON.parse(r.body).status !== undefined,
    }) || apiErrors.add(1);

    apiSuccessRate.add(healthResponse.status === 200);
  });

  group('Flow Management', () => {
    // List all flows
    const listResponse = http.get(`${BASE_URL}/api/flows`, {
      headers: { 'X-API-Key': API_KEY },
    });

    check(listResponse, {
      'list flows status is 200': (r) => r.status === 200,
      'list flows returns array': (r) => Array.isArray(JSON.parse(r.body)),
    }) || apiErrors.add(1);

    apiSuccessRate.add(listResponse.status === 200);

    // Get random flow
    if (data.flows && data.flows.length > 0) {
      const randomFlow = data.flows[Math.floor(Math.random() * data.flows.length)];

      const getResponse = http.get(`${BASE_URL}/api/flows/${randomFlow.id}`, {
        headers: { 'X-API-Key': API_KEY },
      });

      check(getResponse, {
        'get flow status is 200': (r) => r.status === 200,
        'get flow returns correct id': (r) => JSON.parse(r.body).id === randomFlow.id,
      }) || apiErrors.add(1);

      apiSuccessRate.add(getResponse.status === 200);
    }
  });

  group('Flow Execution', () => {
    if (data.flows && data.flows.length > 0) {
      const randomFlow = data.flows[Math.floor(Math.random() * data.flows.length)];

      const startTime = Date.now();

      // Start flow
      const startResponse = http.post(
        `${BASE_URL}/api/flows/${randomFlow.id}/start`,
        null,
        { headers: { 'X-API-Key': API_KEY } }
      );

      check(startResponse, {
        'start flow status is 200': (r) => r.status === 200,
      }) || apiErrors.add(1);

      apiSuccessRate.add(startResponse.status === 200);

      // Wait a bit
      sleep(0.5);

      // Stop flow
      const stopResponse = http.post(
        `${BASE_URL}/api/flows/${randomFlow.id}/stop`,
        null,
        { headers: { 'X-API-Key': API_KEY } }
      );

      check(stopResponse, {
        'stop flow status is 200': (r) => r.status === 200,
      }) || apiErrors.add(1);

      apiSuccessRate.add(stopResponse.status === 200);

      const executionTime = Date.now() - startTime;
      flowExecutionTime.add(executionTime);
    }
  });

  group('Node Operations', () => {
    // Get node types
    const typesResponse = http.get(`${BASE_URL}/api/nodes/types`, {
      headers: { 'X-API-Key': API_KEY },
    });

    check(typesResponse, {
      'get node types status is 200': (r) => r.status === 200,
      'node types is array': (r) => Array.isArray(JSON.parse(r.body)),
      'has core nodes': (r) => {
        const types = JSON.parse(r.body);
        return types.some(t => t.category === 'core');
      },
    }) || apiErrors.add(1);

    apiSuccessRate.add(typesResponse.status === 200);
  });

  group('Metrics Endpoint', () => {
    const metricsResponse = http.get(`${BASE_URL}/metrics`);

    check(metricsResponse, {
      'metrics endpoint status is 200': (r) => r.status === 200,
      'metrics contains prometheus format': (r) => r.body.includes('# TYPE'),
    }) || apiErrors.add(1);

    apiSuccessRate.add(metricsResponse.status === 200);
  });

  // Random sleep between 0.5-2 seconds
  sleep(Math.random() * 1.5 + 0.5);
}

// Teardown function runs once at the end
export function teardown(data) {
  // Clean up test flows
  if (data.flows) {
    for (const flow of data.flows) {
      http.del(`${BASE_URL}/api/flows/${flow.id}`, null, {
        headers: { 'X-API-Key': API_KEY },
      });
    }
  }
}

// Handle summary
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'load-test-results.json': JSON.stringify(data),
  };
}

function textSummary(data, options = {}) {
  const indent = options.indent || '';
  const enableColors = options.enableColors || false;

  let summary = `\n${indent}Load Test Summary\n`;
  summary += `${indent}================\n\n`;

  // Metrics
  for (const [name, metric] of Object.entries(data.metrics)) {
    summary += `${indent}${name}:\n`;
    if (metric.type === 'counter') {
      summary += `${indent}  Count: ${metric.values.count}\n`;
      summary += `${indent}  Rate: ${metric.values.rate.toFixed(2)}/s\n`;
    } else if (metric.type === 'rate') {
      summary += `${indent}  Rate: ${(metric.values.rate * 100).toFixed(2)}%\n`;
      summary += `${indent}  Passes: ${metric.values.passes}\n`;
      summary += `${indent}  Fails: ${metric.values.fails}\n`;
    } else if (metric.type === 'trend') {
      summary += `${indent}  Avg: ${metric.values.avg.toFixed(2)}\n`;
      summary += `${indent}  Min: ${metric.values.min.toFixed(2)}\n`;
      summary += `${indent}  Max: ${metric.values.max.toFixed(2)}\n`;
      summary += `${indent}  p(90): ${metric.values['p(90)'].toFixed(2)}\n`;
      summary += `${indent}  p(95): ${metric.values['p(95)'].toFixed(2)}\n`;
    }
    summary += '\n';
  }

  return summary;
}
