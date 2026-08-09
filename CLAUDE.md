# CLAUDE.md — Heimdall

## What this is

A self-hostable **LLM gateway** in Go. Unified, OpenAI-shaped API in front of pluggable
providers (local Ollama by default), with auth, rate limiting, caching, streaming,
retries/fallback, async jobs, and usage metering.

Scope and milestones: `PRD.md` (milestones are **§7**). Architecture: `TECH_DESIGN.md`
(request pipeline is §3, provider interface §4.1). Read those before proposing design
work — don't re-derive decisions that are already made there.

## Repo facts

- Module `github.com/ismetkoralay/heimdall`. Go 1.26.5, pinned in `.tool-versions`.
- Entrypoint is `cmd/heimdall/service.go` (not `main.go`). `make build` outputs `bin/service`.
- Tests use `stretchr/testify` (`assert`). Keep it — don't mix in a second assertion library.
- **What actually exists right now:** `cmd/heimdall` (empty `main`) and `internal/config`.
  Everything else in `TECH_DESIGN.md` is the plan, not the code. Never describe an
  unwritten package as if it exists.

### Makefile — the real targets

`run`, `build`, `test` (`-race -coverprofile=coverage.out`), `lint` (golangci-lint), `tidy`.

There is **no** `make worker`, no `docker-compose.yml`, no `k8s/`, and no migrations yet.
When a milestone needs one, create it and add it here in the same change.

## Conventions

Go style, error wrapping, testing, and docs rules come from `~/.claude/CLAUDE.md`. Only
the repo-specific deltas live here:

- **Commits:** `[M<milestone>] Imperative subject` — e.g. `[M0-1] Create basic make commands`.
  This repo does **not** use conventional-commit prefixes; the milestone tag wins.
- **Branches:** `M<milestone>/<short-slug>`, e.g. `M0-1/initialize-repo`. Work goes to
  `main` through a PR, not a direct push.
- **Config:** every setting is an env var read in `internal/config` with a sane default
  and validation at load time. Adding one means: field, default, validation, a
  `config_test.go` case, and a line in `.env.example`.
- **Structured logs only.** Never log a prompt body, a response body, or an API key —
  not even hashed, not even at debug, unless an explicit debug flag gates it.

## How to work here

1. **One milestone at a time**, per `PRD.md` §7. Finish and test M<n> before touching M<n+1>.
2. **Build the pipeline as composable middleware** in the `TECH_DESIGN.md` §3 order:
   auth → rate-limit → cache → router → provider → metering. Each independently testable.
3. **`Provider` interface before anything downstream.** Get the abstraction right first;
   `OllamaProvider` is just its first implementation.
4. **`FakeProvider` in every unit and CI test.** CI must never need a running model.
   Ollama is for manual e2e only.
5. **Three things get a short design note + failing test before code:** rate-limit
   atomicity (Redis Lua), cache-key normalization, and streaming flush behavior. These
   are the ones that are painful to retrofit.
6. **Plan before non-trivial work.** Say which files you'll touch and wait for a yes on
   anything that changes the pipeline order, the `Provider` interface, or the DB schema.

## Definition of done

- [ ] `make lint` clean, `make test` green with no live LLM running.
- [ ] New behavior has table-driven tests covering the error branches, not just the happy path.
- [ ] Migrations included when the schema changes; `.env.example` updated when config changes.
- [ ] Any doc claim this change invalidates is fixed in the same commit.

## Avoid

- Hosting or fine-tuning models. Heimdall only fronts them.
- Blocking on streams or sharing buffers across goroutines on the hot path.
- Caching when `temperature > 0` or `stream` is set, unless `force_cache` is explicit.
- Adding a third async delivery mode. Poll first, SSE second, webhook maybe never.
- The Next.js admin dashboard before core + async + metering are done (M4 stretch).
- Writing README.md as if the project were finished. It stays a skeleton until the end —
  run `/readme-final` when the code is actually complete.
