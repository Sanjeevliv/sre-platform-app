
# SRE Platform Application (sre-platform-app)

This repository contains the Go microservices for the SRE Portfolio Project, built by Sanjeev Sethi.

This project is the "Application" layer that runs on the "Platform" defined in the [`sre-platform-infra`](https://github.com/sanjeevsethi/sre-platform-infra) repository.

## 🎯 Purpose

The goal is to build a reliable, observable, and automated application system. The services are deliberately decoupled to demonstrate SRE principles of fault isolation.

## 🛠️ Microservices

This project consists of two distinct Go services:

### 1. `api-service`
* **Description:** A simple REST API server that acts as the user-facing entry point.
* **Instrumentation:** Exposes a `/metrics` endpoint with custom Prometheus metrics (e.g., `api_service_http_requests_total`).
* **Health:** Exposes a `/healthz` endpoint for Kubernetes liveness/readiness probes.

### 2. `worker-service`
* **Description:** A background processor that pulls jobs from a Redis queue (the "jobs" list) and processes them.
* **Instrumentation:** Exposes a `/metrics` endpoint with custom Prometheus metrics (e.g., `worker_service_jobs_processed_total`).
* **Health:** Exposes a `/healthz` endpoint.

## 🚀 Tech Stack

* **Language:** Go (Golang)
* **Containerization:** Docker (using multi-stage, `scratch`-based builds for security and efficiency)
* **Queue:** Redis
* **Observability:** Prometheus (via the `prometheus/client_golang` library)