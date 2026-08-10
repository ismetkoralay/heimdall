# M0 · Issue #4 — Provider Interface + Request/Response Types Design

> Design record for [issue #4](../../../issues/4). Locked decisions only — implementation is
> hand-written separately. See `TECH_DESIGN.md` §4.1/§4.2 for the original sketch this
> refines.

Package: `internal/provider`.

## Locked decisions

1. **Streaming shape:** `ChatStream` returns `iter.Seq2[Chunk, error]` (Go 1.23+
   range-over-func), not channels. No internal goroutine needed for a provider whose
   upstream call is a synchronous decode loop (Ollama NDJSON, OpenAI-style SSE);
   backpressure is pull-based for free, and breaking out of the `for range` triggers the
   iterator's cleanup automatically.

2. **Sync vs. stream duality:** two separate interface methods, `Chat` and `ChatStream` —
   no `Stream bool` flag on the request, no stream-only-plus-collect-helper design.
   `ChatRequest` carries no `Stream` field; the caller's choice of method encodes it.
   Non-streaming callers (async worker in M3, non-stream HTTP requests) call `Chat`.
   The SSE handler (M1) calls `ChatStream`. Each `Provider` implementation decides
   internally whether `Chat` collects from its own streaming path or hits a separate
   non-stream upstream call — that's hidden behind the interface.

3. **Error model:** a typed `*ProviderError` struct, not sentinel-only errors and not a
   marker interface (yet):
   ```go
   type ProviderError struct {
       Provider   string
       StatusCode int
       Retryable  bool
       Err        error
   }
   ```
   Constructed by each `Provider` implementation at the point it detects a failure
   (5xx/timeout/connection-refused → `Retryable: true`; 4xx like bad request or unknown
   model → `Retryable: false`). M2's retry wrapper and circuit breaker check it with
   `errors.As`. Named failure conditions with no per-instance data (e.g. a bad
   model-map entry) stay as plain sentinel errors, matching the existing
   `internal/config` convention, checked with `errors.Is`.

4. **Request field types:** `Temperature` and `MaxTokens` are pointers
   (`*float64`, `*int`), so "unset" is distinguishable from an explicit zero value —
   avoids sending a spurious `0` upstream when the caller didn't set the field.

5. **Response shape:** flat `ChatResponse` — no OpenAI-style `choices[]` array. No
   `n > 1` support in v1; the PRD doesn't ask for it. Additive later (`Choices
   []ChatResponse`) without breaking the existing fields.

6. **Usage placement in streaming:** usage rides on the terminal `Chunk`, as a `*Usage`
   field that is `nil` on every chunk except the last. Mirrors OpenAI's
   `stream_options.include_usage` convention. Metering (M3) watches for a non-nil
   `Usage` while relaying chunks — same signal shape for both the sync path
   (`ChatResponse.Usage`) and the stream path.

7. **Wire format vs. internal types:** `ChatRequest`/`ChatResponse`/`Chunk`/`Usage` are
   internal, provider-facing types — **not** the HTTP wire format. A separate DTO layer
   in M1 translates JSON ↔ these types. This keeps OpenAI wire-format quirks (field
   naming, optional/legacy parameters) out of the Provider-facing shape, and keeps this
   package free of JSON tags.

8. **Interface surface:** confirmed at exactly four methods — `Chat`, `ChatStream`,
   `Embed`, `Name`. No health/capability/close method added speculatively; add one only
   when a concrete second provider implementation forces the question.

9. **Package location:** `internal/provider`.

## Deliberate divergences from the OpenAI shape

- No `choices[]` array (decision 5).
- No tool/function-calling fields — not in PRD scope; additive later.
- `ForceCache` (from `TECH_DESIGN.md` §4.4's cache-bypass rule) has no OpenAI
  equivalent; it lives only on the internal `ChatRequest`, set by the M1 DTO layer.

## Deferred to later milestones

- `Retryable() bool` marker interface generalizing beyond `*ProviderError` — only worth
  adding once a second retryable-capable error type exists (e.g. a circuit-breaker
  sentinel in M2).
- Provider capability/health/close methods — add only when a concrete provider needs
  one.
