import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'https://middleware-40815901860.europe-west4.run.app';
const payload = open('./test-payloads/challenge.json');

export const options = {
  scenarios: {
    steady: {
      executor: 'constant-arrival-rate',
      rate: 12,
      timeUnit: '1m',
      duration: '10m',
      preAllocatedVUs: 1,
      maxVUs: 5,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.001'],
  },
};

export function setup() {
  const url = `${BASE_URL}/evidence/infrastructure`;
  for (let i = 0; i < 3; i++) {
    http.post(url, payload, {
      headers: { 'Content-Type': 'application/json' },
      timeout: '60s',
      tags: { endpoint: 'infrastructure', phase: 'warmup' },
    });
    sleep(0.5);
  }
}

export default function () {
  const url = `${BASE_URL}/evidence/infrastructure`;

  const res = http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s',
    tags: { endpoint: 'infrastructure', phase: 'steady' },
  });

  const ok = check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!ok) {
    console.error(
      `FAIL endpoint=infrastructure status=${res.status} body=${String(
        res.body,
      ).slice(0, 300)}`,
    );
  }
}
