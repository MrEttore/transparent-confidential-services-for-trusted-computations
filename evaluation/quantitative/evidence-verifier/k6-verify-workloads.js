import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'http://127.0.0.1:8081';
const payload = open('./test-payloads/workloads.json');

export const options = {
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: 1,
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 1,
      maxVUs: 10,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
  },
};

export function setup() {
  const url = `${BASE_URL}/verify/workloads`;
  for (let i = 0; i < 5; i++) {
    http.post(url, payload, {
      headers: { 'Content-Type': 'application/json' },
      timeout: '60s',
      tags: { endpoint: 'verify-workloads', phase: 'warmup' },
    });
    sleep(0.2);
  }
}

export default function () {
  const url = `${BASE_URL}/verify/workloads`;

  const res = http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s',
    tags: { endpoint: 'verify-workloads', phase: 'steady' },
  });

  const ok = check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!ok) {
    console.error(
      `FAIL endpoint=verify-workloads status=${res.status} body=${String(
        res.body,
      ).slice(0, 300)}`,
    );
  }
}
