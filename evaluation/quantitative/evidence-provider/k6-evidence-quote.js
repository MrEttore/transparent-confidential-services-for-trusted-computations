import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = 'https://middleware-40815901860.europe-west4.run.app';
const payload = open('./test-payloads/challenge.json');

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
  const url = `${BASE_URL}/evidence/tdx-quote`;
  for (let i = 0; i < 5; i++) {
    http.post(url, payload, {
      headers: { 'Content-Type': 'application/json' },
      timeout: '60s',
      tags: { endpoint: 'tdx-quote', phase: 'warmup' },
    });
    sleep(0.5);
  }
}

export default function () {
  const url = `${BASE_URL}/evidence/tdx-quote`;

  const res = http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s',
    tags: { endpoint: 'tdx-quote', phase: 'steady' },
  });

  const ok = check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!ok) {
    console.error(
      `FAIL endpoint=tdx-quote status=${res.status} body=${String(
        res.body,
      ).slice(0, 300)}`,
    );
  }
}
