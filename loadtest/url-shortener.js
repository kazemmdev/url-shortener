// k6 load test for the URL shortener API.
//
// Env vars:
//   BASE_URL       API base url                          (default http://localhost:8080)
//   PROFILE        smoke | load | stress | spike         (default load)
//   TARGET_RPS     peak total requests/sec for profile   (default per profile)
//   WRITE_PERCENT  share of traffic that creates links   (default 10)
//   SEED_URLS      links created in setup for reads     (default 200)
//   MAX_VUS        VU cap per scenario                   (default 1000)

import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const PROFILE = (__ENV.PROFILE || 'load').toLowerCase();
const WRITE_PERCENT = Number(__ENV.WRITE_PERCENT ?? 10);
const SEED_URLS = Number(__ENV.SEED_URLS || 200);
const MAX_VUS = Number(__ENV.MAX_VUS || 1000);

// Each stage is [fraction of peak rate, duration].
const PROFILES = {
  smoke: { peak: 10, stages: [[1, '30s']] },
  load: { peak: 300, stages: [[1, '1m'], [1, '3m'], [0, '30s']] },
  stress: {
    peak: 3000,
    stages: [[0.1, '1m'], [0.25, '2m'], [0.5, '2m'], [0.75, '2m'], [1, '2m'], [0, '1m']],
  },
  spike: {
    peak: 3000,
    stages: [[0.05, '30s'], [1, '10s'], [1, '1m'], [0.05, '10s'], [0.05, '30s']],
  },
};

const profile = PROFILES[PROFILE];
if (!profile) {
  throw new Error(`Unknown PROFILE "${PROFILE}". Use one of: ${Object.keys(PROFILES).join(', ')}`);
}
const peak = Number(__ENV.TARGET_RPS || profile.peak);

function arrivalScenario(exec, share) {
  const peakRate = Math.ceil(peak * share);
  return {
    executor: 'ramping-arrival-rate',
    exec,
    startRate: 0,
    timeUnit: '1s',
    preAllocatedVUs: Math.min(MAX_VUS, Math.max(10, Math.ceil(peakRate / 10))),
    maxVUs: MAX_VUS,
    stages: profile.stages.map(([fraction, duration]) => ({
      target: Math.round(peakRate * fraction),
      duration,
    })),
  };
}

const scenarios = { redirect: arrivalScenario('redirect', 1 - WRITE_PERCENT / 100) };
if (WRITE_PERCENT > 0) {
  scenarios.create = arrivalScenario('create', WRITE_PERCENT / 100);
}

export const options = {
  scenarios,
  setupTimeout: '3m',
  // Never follow the 302 to the long url, we only measure our API.
  maxRedirects: 0,
  discardResponseBodies: false,
  thresholds: {
    http_req_failed: ['rate<0.01'],
    'http_req_duration{name:redirect}': ['p(95)<200', 'p(99)<500'],
    'http_req_duration{name:create}': ['p(95)<500', 'p(99)<1000'],
    'checks{name:redirect}': ['rate>0.99'],
    'checks{name:create}': ['rate>0.99'],
  },
};

function createUrl() {
  const longUrl = `https://example.com/${__VU}/${Date.now()}/${Math.random().toString(36).slice(2)}`;
  return http.post(`${BASE_URL}/api/url`, JSON.stringify({ longUrl }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'create' },
  });
}

function shortCodeOf(res) {
  return (res.body || '').replace(/"/g, '').trim();
}

function waitForApi() {
  for (let i = 0; i < 60; i++) {
    const res = http.get(`${BASE_URL}/api/url/healthprobe`, {
      tags: { name: 'setup' },
      responseCallback: http.expectedStatuses(404),
    });
    if (res.status === 404) return;
    sleep(2);
  }
  throw new Error(`API at ${BASE_URL} did not become ready`);
}

export function setup() {
  waitForApi();

  const codes = [];
  for (let i = 0; i < SEED_URLS; i++) {
    const res = createUrl();
    if (res.status === 200) codes.push(shortCodeOf(res));
  }
  if (codes.length === 0) {
    throw new Error('Could not seed any short urls, check that the API can reach the database');
  }
  console.log(`profile=${PROFILE} peak=${peak} rps write=${WRITE_PERCENT}% seeded=${codes.length}`);
  return { codes };
}

export function redirect(data) {
  const code = data.codes[Math.floor(Math.random() * data.codes.length)];
  const res = http.get(`${BASE_URL}/api/url/${code}`, { tags: { name: 'redirect' } });
  check(res, { 'redirect is 302': (r) => r.status === 302 }, { name: 'redirect' });
}

export function create() {
  const res = createUrl();
  check(
    res,
    {
      'create is 200': (r) => r.status === 200,
      'create returns code': (r) => shortCodeOf(r).length === 7,
    },
    { name: 'create' },
  );
}
