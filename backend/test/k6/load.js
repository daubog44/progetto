import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    stages: [
        { duration: '30s', target: 20 }, // Wrap up to 20 users
        { duration: '1m', target: 20 },  // Stay at 20 users
        { duration: '30s', target: 0 },  // Scale down
    ],
};

export default function () {
    // Test Gateway Health (Use env var or default to docker service name)
    const target = __ENV.TARGET || 'http://gateway-service:8888';
    let res = http.get(`${target}/ping`);
    check(res, { 'status was 200': (r) => r.status == 200 });

    // Test Metrics Middleware
    // We expect this to generate traces and metrics
    sleep(1);
}
