---
description: Critically review one of my PRs against this repo's conventions and the issue's Definition of Done
argument-hint: <pr-number> <issue-number>
---

Review PR #$1 for me (linked issue: #$2). I wrote this myself and want a critical review, not a rewrite.

First gather context with `gh`:
- `gh pr view $1` and `gh pr diff $1` for the change.
- `gh issue view $2` so you can check the work against its Definition of Done.
  (If $2 is empty, find the linked issue from the PR description instead.)
- Re-read CLAUDE.md and TECH_DESIGN.md so you review against THIS repo's conventions, not generic style.

Then use the @reviewer subagent (read-only). Do NOT edit code or push commits — give me findings and let me decide what to change.

Review priorities, in order:
1. Correctness — bugs, edge cases, wrong error handling.
2. Concurrency (this is Go and I'm doing it to learn) — goroutine leaks, unclosed channels, missing context cancellation, races, anything that blocks other requests. Call these out even if minor.
3. Security — secrets in logs, unvalidated input, unsafe defaults.
4. Does it actually satisfy the issue's Definition of Done? List any unchecked item.
5. Tests — is the tricky logic covered, and does it run without a live LLM?
6. Idiomatic Go — only flag things that matter, not nitpicks.

Output format: a list of findings as `severity | file:line | issue | why it matters | suggested fix`. If something is genuinely fine, say so briefly — don't invent problems to look useful. End with a one-line verdict: ship / fix-first / needs-discussion, and the single most important thing to address.

Be direct. I'd rather hear it's wrong now than in an interview.