import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

const BASE_URL = 'http://127.0.0.1:8081';
const payload = open('./test-payloads/quote.json');

const ENDPOINT = '/verify/tdx-quote';
const ENDPOINT_TAG = 'verify-tdx-quote';

const PHASE_PACE_SEC = {
  every_60s: 60,
  every_30s: 30,
  every_10s: 10,
};

export const options = {
  scenarios: {
    every_60s: {
      executor: 'constant-vus',
      vus: 1,
      duration: '3m',
      startTime: '0s',
    },
    every_30s: {
      executor: 'constant-vus',
      vus: 1,
      duration: '3m',
      startTime: '3m10s',
    },
    every_10s: {
      executor: 'constant-vus',
      vus: 1,
      duration: '3m',
      startTime: '6m20s',
    },
  },
};

export function setup() {
  const url = `${BASE_URL}${ENDPOINT}`;
  http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s',
    tags: { endpoint: ENDPOINT_TAG, phase: 'warmup' },
  });
  sleep(2);
}

export default function () {
  const phase = exec.scenario.name;
  const pace = PHASE_PACE_SEC[phase] ?? 60;

  const url = `${BASE_URL}${ENDPOINT}`;
  const res = http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '60s',
    tags: { endpoint: ENDPOINT_TAG, phase },
  });

  const ok = check(res, {
    'status is 2xx': (r) => r.status >= 200 && r.status < 300,
  });

  if (!ok) {
    console.error(
      `FAIL endpoint=${ENDPOINT_TAG} phase=${phase} status=${
        res.status
      } body=${String(res.body || '').slice(0, 300)}`,
    );
  }

  const durSec = (res?.timings?.duration || 0) / 1000;
  sleep(Math.max(0, pace - durSec));
}
