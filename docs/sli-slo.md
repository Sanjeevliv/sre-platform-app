# Service Level Indicators (SLIs) and Objectives (SLOs)

> This document defines the reliability targets for the SRE Platform.

## Overview

| Service        | SLI              | SLO Target | Measurement Window |
| -------------- | ---------------- | ---------- | ------------------ |
| api-service    | Availability     | 99.9%      | 30 days            |
| api-service    | Latency (p99)    | < 300ms    | 30 days            |
| worker-service | Job Success Rate | 99.5%      | 30 days            |

---

## SLI Definitions

### 1. API Availability

**Definition**: The percentage of successful HTTP requests (non-5xx responses).

```
SLI = (total_requests - 5xx_errors) / total_requests × 100
```

**Prometheus Query**:

```promql
sum(rate(api_http_request_duration_seconds_count{status!~"5.."}[5m]))
/
sum(rate(api_http_request_duration_seconds_count[5m]))
```

---

### 2. API Latency (p99)

**Definition**: 99th percentile response time for all API requests.

```
SLI = histogram_quantile(0.99, request_duration)
```

**Prometheus Query**:

```promql
histogram_quantile(0.99,
  sum(rate(api_http_request_duration_seconds_bucket[5m])) by (le)
)
```

---

### 3. Worker Job Success Rate

**Definition**: The percentage of jobs completed without errors.

```
SLI = successful_jobs / total_jobs × 100
```

**Prometheus Query**:

```promql
sum(rate(worker_service_jobs_processed_total[5m]))
/
(sum(rate(worker_service_jobs_processed_total[5m])) + sum(rate(worker_service_jobs_failed_total[5m])))
```

---

## SLO Targets

### API Service

| SLO          | Target  | Error Budget (30 days)      |
| ------------ | ------- | --------------------------- |
| Availability | 99.9%   | 43.2 minutes downtime       |
| Latency p99  | < 300ms | 0.1% of requests can exceed |

### Worker Service

| SLO              | Target | Error Budget (30 days) |
| ---------------- | ------ | ---------------------- |
| Job Success Rate | 99.5%  | 0.5% of jobs can fail  |

---

## Error Budget

**Error Budget** = 100% - SLO Target

For 99.9% availability over 30 days:

- Total minutes: 30 × 24 × 60 = 43,200 minutes
- Error budget: 0.1% × 43,200 = **43.2 minutes**

### Error Budget Policy

| Budget Remaining | Action                          |
| ---------------- | ------------------------------- |
| > 50%            | Normal development velocity     |
| 25-50%           | Reduce risky deployments        |
| < 25%            | Focus on reliability work only  |
| Exhausted        | Freeze all non-critical changes |

---

## Alerting Thresholds

| Alert                  | Condition                | Severity |
| ---------------------- | ------------------------ | -------- |
| SLO Burn Rate High     | > 14.4x burn rate for 1h | Critical |
| SLO Burn Rate Elevated | > 6x burn rate for 6h    | Warning  |
| Error Budget Low       | < 25% remaining          | Warning  |
| Error Budget Exhausted | 0% remaining             | Critical |

---

## Dashboard

Access the SLO dashboard at: `monitor.sanjeevsethi.in` (when deployed)

Key metrics displayed:

- Current SLI values
- SLO compliance status
- Error budget remaining
- Burn rate trends
