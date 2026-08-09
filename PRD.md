# Heimdall — Product Requirements Document

> **One-liner:** A self-hostable **LLM gateway** in Go that sits in front of multiple model providers, exposing unified **sync and async** endpoints with auth, rate limiting, caching, streaming, retries/fallback, and full usage + cost accounting.

---

## 1. Goals & Non-Goals

### Goals
- One **unified API** for chat/completions + embeddings across providers.
- **Provider routing**: pick a backend by model name, with fallback on failure.
- **Sync** request/response **and** **async** (submit job → poll / SSE / webhook callback).
- **Streaming** responses (SSE) for chat.
- **Auth** via API keys, with **per-key rate limits and quotas**.
- **Caching** (exact-match, optionally semantic) to cut latency and load.
- **Usage metering**: tokens, request counts, estimated cost, per key + per model.
- **Observability**: metrics, structured logs, traces.

### Non-Goals (v1)
- Training, fine-tuning, or model hosting (Heimdall *fronts* models, Ollama hosts them).
- A billing system (it *meters* usage; invoicing is out of scope).
- A full web console (a minimal admin API is enough; a small dashboard is a stretch).

## 2. Target Users & Use Cases

| User | Use case |
|------|----------|
| App team | Calls `POST /v1/chat/completions` with their key; doesn't care which provider serves it. |
| Platform owner | Issues keys, sets quotas, watches a usage dashboard, swaps providers without app changes. |
| Batch/offline job | Submits hundreds of async requests, polls for results later. |

**Primary scenario:** A team sends an OpenAI-style chat request to Heimdall. Heimdall authenticates the key, checks the rate limit, finds a cached answer or routes to Ollama, streams tokens back, and records token usage + cost against the key.

## 3. Functional Requirements

### 3.1 API Surface (OpenAI-compatible where practical)
- FR-1: `POST /v1/chat/completions` — sync, supports `stream: true` (SSE).
- FR-2: `POST /v1/embeddings` — sync.
- FR-3: `POST /v1/async/chat/completions` → returns `job_id`.
- FR-4: `GET /v1/async/jobs/{id}` → status + result when ready; optional SSE subscribe; optional webhook callback on completion.
- Being OpenAI-shaped means existing client SDKs work with just a base-URL change — a great demo.

### 3.2 Auth & Governance
- FR-5: Every request requires an API key (`Authorization: Bearer hd_...`).
- FR-6: Per-key **rate limit** (requests/min) and **quota** (tokens/day) enforced; over-limit returns `429` with `Retry-After`.
- FR-7: Admin endpoints (separate admin key): create/revoke keys, set limits, read usage.

### 3.3 Routing & Resilience
- FR-8: Requests route to a provider by model name via a config-driven model map.
- FR-9: On provider error/timeout, optional **fallback** to a secondary model/provider.
- FR-10: Retries with exponential backoff + jitter; a **circuit breaker** trips a failing provider.

### 3.4 Caching
- FR-11: Exact-match cache keyed by normalized request → response (configurable TTL, skip when `stream` or `temperature>0` unless forced).
- FR-12 (stretch): **Semantic cache** — embed the prompt, return a cached answer if cosine similarity ≥ threshold.

### 3.5 Metering
- FR-13: Record per request: key, model, prompt/completion tokens, latency, cache hit, estimated cost (from a configurable price table).
- FR-14: Aggregated usage queryable per key and per model over a time range.

## 4. Non-Functional Requirements
- **Cost:** $0 locally (Ollama). Adding a paid provider is config + a price-table entry only.
- **Latency overhead:** gateway adds < ~15ms p50 on a cache miss (excluding model time).
- **Throughput:** handle concurrent streams without head-of-line blocking.
- **Reliability:** a down provider must not take the gateway down — fail fast, fall back, surface a clean error.
- **Security:** keys stored hashed; admin surface separated; no prompt content logged unless debug explicitly enabled.

## 5. Success Metrics
- Drop-in works: an off-the-shelf OpenAI client talks to Heimdall unmodified.
- Cache hit ratio on a repeated workload.
- Correct quota enforcement (load test that 429s kick in at the boundary).
- Usage numbers reconcile with provider-reported tokens within tolerance.

## 6. Milestones

| Milestone | Scope |
|-----------|-------|
| **M0 – Skeleton** | HTTP server, config loader, Ollama provider, passthrough `POST /v1/chat/completions` (non-stream). |
| **M1 – Core gateway (MVP)** | API-key auth, model→provider routing, SSE streaming, structured logging. |
| **M2 – Governance** | Redis rate limiting + quotas, exact-match cache, retries + fallback + circuit breaker. |
| **M3 – Async + metering** | Async job submit/poll/SSE + Redis-backed worker queue; usage metering in Postgres; admin API for keys + usage. |
| **M4 – Ops & stretch** | Prometheus metrics, OpenTelemetry tracing, Docker + kind manifests, README + demo; **stretch:** semantic cache, tiny Next.js admin dashboard. |

## 7. Open Questions
- Exact-match cache normalization rules (whitespace, message ordering) — define precisely to avoid false hits.
- Async result delivery: support all three (poll/SSE/webhook) or start with poll only? (Start poll-only, add SSE.)
- Cost table maintenance for real providers — ship a default JSON, allow override.
