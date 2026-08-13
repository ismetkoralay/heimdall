# Heimdall — Technical Design

> Companion to `PRD.md`. Architecture for the LLM gateway, runnable end-to-end on local models.

---

## 1. Tech Stack

| Concern | Choice | Why |
|--------|--------|-----|
| Language | **Go** | Excellent for high-concurrency proxies; streaming, goroutines, single binary. |
| HTTP | `chi` or stdlib | Middleware chaining for auth/rate-limit/logging. |
| LLM (local) | **Ollama** | Free local inference; OpenAI-compatible-ish endpoints. |
| State | **Postgres** | API keys, usage records, async jobs. |
| Hot path store | **Redis** | Rate-limit counters, exact cache, async job queue, circuit state. |
| Metrics/trace | Prometheus + OpenTelemetry | Platform-grade observability. |
| Deploy | Docker + k8s, **kind** | Zero cloud cost; lighter than minikube (nodes run as containers). |
| Admin UI (stretch) | **Next.js (React)** | Small dashboard for keys + usage. |

## 2. Architecture

```
                         ┌─────────────────────────── Heimdall ──────────────────────────┐
 client (OpenAI SDK) ───▶│  [auth] → [rate-limit] → [cache] → [router] → [provider] ──────┼──▶ Ollama
   Bearer hd_...         │     │          │            │          │            │           │   (OpenAI / Bedrock
                         │   Postgres   Redis        Redis    model-map    retries/CB      │    adapters, pluggable)
                         │  (keys)    (counters)    (exact)               fallback          │
                         │                                                                  │
                         │  async: POST → enqueue(Redis) → worker pool → store result(PG) ──┼──▶ provider
                         │  metering: every response → usage record(PG) + metrics(Prom)     │
                         └──────────────────────────────────────────────────────────────────┘
```

## 3. Request Pipeline (middleware order)
1. **Auth** — extract bearer key, look up hashed key (cached in Redis), attach key context. Reject `401` if invalid/revoked.
2. **Rate limit / quota** — Redis token-bucket per key (rpm) + daily token quota check. `429 + Retry-After` on breach.
3. **Cache** — on eligible requests, hash the normalized request; return cached response on hit (mark `X-Heimdall-Cache: hit`).
4. **Router** — resolve `model` → provider+endpoint from config model-map.
5. **Provider call** — with retry (backoff+jitter), fallback model, circuit breaker per provider.
6. **Metering** — record usage (tokens, latency, cost, cache flag) and emit metrics.

## 4. Components

### 4.1 Provider abstraction
```go
type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
    ChatStream(ctx context.Context, req ChatRequest) (stream <-chan Chunk, errc <-chan error)
    Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
    Name() string
}
```
- `OllamaProvider` first. `OpenAIProvider`, `BedrockProvider` later — same interface.
- Model map (config):
```yaml
providers:
  - name: ollama
    base_url: http://localhost:11434
models:
  - name: gpt-4o-mini
    provider: ollama
    upstream_model: qwen2.5:7b
    fallback: llama3.1:8b
  - name: text-embed
    provider: ollama
    upstream_model: nomic-embed-text
```

### 4.2 Streaming
- SSE handler relays provider chunks as `data:` events, flushing per token; `[DONE]` sentinel at end. Each goroutine owns its stream — no shared buffers, no head-of-line blocking.

### 4.3 Rate limiting & quotas
- Redis token bucket (Lua script for atomicity) keyed `rl:{keyID}` for rpm.
- Daily token quota counter `quota:{keyID}:{yyyymmdd}` incremented post-response; pre-check rejects when exhausted.

### 4.4 Caching
- **Exact:** key = `sha256(model + normalized messages + params)`. Store response JSON in Redis with TTL. Skip when streaming or `temperature>0` (unless `force_cache`).
- **Semantic (stretch):** embed prompt → search a small vector index (reuse Qdrant or pgvector) → reuse answer if similarity ≥ threshold. Guard with a max age + opt-in flag.

### 4.5 Resilience
- Retry wrapper: N attempts, exponential backoff + jitter, only on retryable errors (timeouts, 5xx).
- Circuit breaker per provider (e.g. `sony/gobreaker`): trips on consecutive failures, half-opens after cooldown.
- Fallback: on open circuit or hard failure, try the configured fallback model once.

### 4.6 Async subsystem
- `POST /v1/async/...` validates + enqueues a job (Redis list/stream) and returns `job_id` + `202`.
- Worker pool consumes jobs, calls the same provider pipeline, writes status/result to Postgres.
- `GET /v1/async/jobs/{id}` returns `queued|running|done|error` + result. Optional SSE subscription; optional webhook POST on completion.

### 4.7 Metering & admin
- Each response writes a `usage` row. Cost = tokens × price-table rate.
- Admin API (admin key): `POST /admin/keys`, `DELETE /admin/keys/{id}`, `GET /admin/usage?key=&from=&to=`.

## 5. Data Model (Postgres)
```
api_keys(id, hashed_key, name, rpm_limit, daily_token_quota, revoked, created_at)
usage(id, key_id, model, provider, prompt_tokens, completion_tokens,
      latency_ms, cache_hit, est_cost_usd, created_at)
async_jobs(id, key_id, status, request_json, result_json, error, created_at, updated_at)
```

## 6. Local Development
1. `docker compose up` (`make up`) → Heimdall + Ollama + Postgres/Redis (later milestones). Ollama runs
   as its own compose service; an init job pulls the configured models on startup. Heimdall reaches it
   at `http://ollama:11434` over the compose network — no host-installed Ollama or `host.docker.internal`
   needed.
2. Create a key via admin API; point an OpenAI client at `http://localhost:8080/v1` with that key and chat.
3. Hit it twice to see a cache hit; spam it to trip the 429; submit an async job and poll.

## 7. Deployment (kind)
- Multi-stage Dockerfile → distroless.
- `k8s/`: Deployment (gateway), Deployment (async worker), Postgres + Redis (or Helm/manifests), Secrets, Service, optional Ingress. HPA on the gateway is a nice stretch demo.
- Flow: `kind create cluster --name dev` → `docker build` → `kind load docker-image heimdall:latest --name dev` → `kubectl apply -k k8s/`.
- **Note:** kind does not share the host Docker daemon — locally-built images must be loaded with `kind load docker-image` after each rebuild (then `kubectl rollout restart`). Manifests use `imagePullPolicy: IfNotPresent`.

## 8. Observability
- Metrics: `heimdall_requests_total{model,status}`, latency histogram, `cache_hits_total`, `rate_limited_total`, `provider_errors_total{provider}`, breaker state gauge.
- Traces (OTel): span per request across middleware → provider call.
- Logs: structured, request-ID correlated; prompt bodies redacted unless debug.

## 9. Testing Strategy
- Unit: auth, rate-limit Lua logic, cache key normalization, router resolution, retry/backoff, breaker transitions.
- `FakeProvider` for pipeline tests (deterministic, fast).
- Contract test: an OpenAI client library can actually talk to Heimdall.
- Load test (k6/vegeta): verify 429 boundary + streaming under concurrency.

## 10. Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| OpenAI-compatibility edge cases | Target the common subset; document deviations. |
| Cache false hits | Strict normalization; conservative default (no cache when temp>0). |
| Async complexity creep | Ship poll-only first; SSE/webhook after. |
| Local model slowness skews tests | Use `FakeProvider` for CI; Ollama only for manual/e2e. |
