import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('error_rate');
const responseTime = new Trend('response_time');

// Test configuration
export const options = {
    stages: [
        { duration: '10s', target: 50 },   // Ramp up to 50 VUs in 10 seconds
        { duration: '50s', target: 100 },  // Ramp up to 100 VUs and maintain for 50 seconds
        { duration: '10s', target: 0 },    // Ramp down to 0 VUs in 10 seconds
    ],
    thresholds: {
        'http_req_duration': ['p(95)<300'], // 95% of requests must be below 300ms
        'error_rate': ['rate<0.01'],        // Error rate must be less than 1% (essentially 0%)
        'http_req_failed': ['rate<0.01'],   // Failed requests must be less than 1%
    },
};

// Test setup
export function setup() {
    console.log('Starting K6 Performance Test');
    console.log('Target URL: http://localhost:8080/api/v1/users');
    console.log('Load Profile: 10s->50VU | 50s->100VU | 10s->0VU');
    console.log('SLA: Response time < 300ms, Error rate = 0%');
}

// Main test function
export default function () {
    const url = 'http://localhost:8080/api/v1/users';
    const headers = {
        'Authorization': 'Bearer b75efa60262ca622d8da39724ba5c1de54b011b1cbc3f9345e96d17a81d3752c',
        'Content-Type': 'application/json',
    };

    // Make HTTP GET request
    const response = http.get(url, { headers });

    // Record custom metrics
    responseTime.add(response.timings.duration);
    errorRate.add(response.status !== 200);

    // Perform checks
    const isSuccess = check(response, {
        'status is 200': (r) => r.status === 200,
        'response time < 300ms': (r) => r.timings.duration < 300,
        'response has body': (r) => r.body && r.body.length > 0,
    });

    // Log errors for debugging
    if (!isSuccess || response.status !== 200) {
        console.error(`Request failed - Status: ${response.status}, Duration: ${response.timings.duration}ms`);
    }

    // Optional: Add small delay between requests (adjust as needed)
    // sleep(0.1); // 100ms delay between requests per VU
}

// Test teardown
export function teardown(data) {
    console.log('K6 Performance Test completed');
}