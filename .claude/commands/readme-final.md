---
description: Rewrite README.md from the real code, once the project is feature-complete
allowed-tools: Read, Grep, Glob, Bash(ls:*), Bash(make -n:*), Bash(git log:*), Edit, Write
---

Only run this when Heimdall is feature-complete. A README written early describes code
that doesn't exist and quietly lies — and it's the most-read file in the repo.

First: check. List what actually exists under `cmd/`, `internal/`, `k8s/`, and the
Makefile targets. If the milestones in `PRD.md` §7 aren't done, say so and stop.

Then rewrite `README.md` based ONLY on what exists in this repository. Read the code, the
Makefile, the compose file, and the k8s manifests before writing a word.

Section order:

1. One-line description.
2. The problem it solves — 2–3 sentences, no marketing language.
3. Architecture diagram (mermaid) reflecting the real components, not the planned ones.
4. Quick start. Verify every command by running it before including it. Include the
   drop-in demo: an off-the-shelf OpenAI client pointed at the gateway with a created key.
5. Key design decisions & trade-offs — leave `<!-- TODO: mons -->` markers under each
   heading. Do not write this section's prose; it has to be my reasoning, not yours.
   The decisions worth defending: exact-match vs semantic cache, the Redis token-bucket
   rate limiter, Postgres for async jobs, and the shape of the `Provider` interface.
6. Local development.
7. Deploy to kind — include the `kind load docker-image` step. kind does not share the
   host Docker daemon, and skipping it silently redeploys stale code.

Rules: no emoji, no exhaustive feature list, no filler. Every command must work as
written. Where the skeleton README no longer matches the code, fix it to match reality.

Finally, print a short checklist of what only I can do: verify the quick-start by hand
once more, write the design-decisions prose, and record a demo GIF (a streamed response,
then a 429 when the quota trips).

$ARGUMENTS
