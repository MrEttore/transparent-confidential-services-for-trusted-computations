import http from 'k6/http';
import { check, sleep } from 'k6';
import exec from 'k6/execution';

const BASE_URL = 'http://127.0.0.1:8081';
const payload = open('./test-payloads/workloads.json');

const ENDPOINT = '/verify/workloads';
const ENDPOINT_TAG = 'verify-workloads';

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

function extractMessage(res) {
  try {
    const j = res.json();
    return String(j?.Message || j?.message || '');
  } catch (_) {
    return String(res.body || '');
  }
}

function isManifest429(res, msg) {
  if (res.status === 429) return true;
  if (res.status === 422) {
    const m = (msg || '').toLowerCase();
    return m.includes('429') && m.includes('manifest');
  }
  return false;
}

function postWithRetry(url, body, tags) {
  const headers = { 'Content-Type': 'application/json' };

  const maxAttempts = 4;
  const backoffs = [5, 10, 20, 30];

  let res;
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    res = http.post(url, body, {
      headers,
      timeout: '90s',
      tags,
    });

    if (res.status >= 200 && res.status < 300) return res;

    const msg = extractMessage(res);

    if (!isManifest429(res, msg) || attempt === maxAttempts) {
      return res;
    }

    sleep(backoffs[Math.min(attempt - 1, backoffs.length - 1)]);
  }

  return res;
}

export function setup() {
  const url = `${BASE_URL}${ENDPOINT}`;
  http.post(url, payload, {
    headers: { 'Content-Type': 'application/json' },
    timeout: '90s',
    tags: { endpoint: ENDPOINT_TAG, phase: 'warmup' },
  });
  sleep(2);
}

export default function () {
  const phase = exec.scenario.name;
  const pace = PHASE_PACE_SEC[phase] ?? 60;

  const url = `${BASE_URL}${ENDPOINT}`;
  const res = postWithRetry(url, payload, { endpoint: ENDPOINT_TAG, phase });

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
  const sleepFor = Math.max(0, pace - durSec);
  sleep(sleepFor);
}
