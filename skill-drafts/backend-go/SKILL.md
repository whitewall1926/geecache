---
name: backend-go
description: Use when the user wants to practice backend development with Go, improve backend engineering habits, or build/refactor Go services, libraries, APIs, concurrency-heavy modules, or distributed-system exercises with stronger attention to interfaces, data flow, error handling, testing, observability, and production-minded tradeoffs.
---

# Backend Go

Use this skill when the goal is not just to "make the Go code work", but to train backend engineering judgment while building or changing Go code.

## Outcome

Optimize for these outcomes:

- Clear boundaries between transport, business logic, storage, and infrastructure concerns
- Explicit handling of data flow, failure modes, and concurrency
- Small, testable units with stable interfaces
- Changes that are verified, not just implemented
- Explanations that teach backend reasoning, not only syntax

## Default Workflow

Follow this order unless the user asks for something narrower:

1. Identify what kind of backend work this is.
2. Map the request to interfaces, data flow, state, and failure paths.
3. Check how the current codebase organizes responsibilities.
4. Make the smallest change that improves both behavior and structure.
5. Verify with tests, builds, or focused runtime checks.
6. Explain the backend tradeoffs behind the final shape.

## Classify The Work

Start by classifying the task. This changes what "good" looks like.

- Library or core module:
  Optimize for API clarity, invariants, low coupling, and testability.
- Service or HTTP API:
  Optimize for handler boundaries, validation, timeouts, idempotency, logging, and response contracts.
- Concurrency-heavy component:
  Optimize for ownership, locking, contention, cancellation, and race safety.
- Storage or cache layer:
  Optimize for consistency, eviction, serialization, stale data handling, and performance under load.
- Distributed-system exercise:
  Optimize for node selection, partial failure behavior, retries, deduplication, and observability.

## Backend Checklist

Before changing code, reason through these questions:

- What are the inputs and outputs?
- Where does state live?
- What invariants must always hold?
- What can fail, and how does that failure surface?
- Which parts are synchronous, asynchronous, or shared across goroutines?
- What is the public interface, and what should stay internal?
- How will this be tested?

If any of these are unclear, resolve them before making structural changes.

## Design Rules

Prefer these patterns:

- Keep interfaces near the consumer, not in a central "interfaces" package by default.
- Keep transport concerns out of business logic.
- Return typed or wrapped errors when callers need to branch on failure.
- Pass `context.Context` through request-scoped or cancellation-aware paths.
- Make concurrent ownership explicit instead of relying on "probably safe" assumptions.
- Hide mutable shared state behind narrow methods.
- Prefer composition over large manager types with broad responsibilities.

Avoid these patterns unless the codebase already requires them:

- Controllers or handlers that own caching, storage, and formatting logic at once
- Global mutable state without strict justification
- Concurrency added before correctness is established
- Refactors that increase indirection without improving a real backend property

## Training Bias

When multiple implementations are possible, prefer the one that teaches stronger backend habits:

- Explicit dependencies over hidden coupling
- Small seams for tests over giant end-to-end-only logic
- Clear contracts over clever shortcuts
- Measured performance reasoning over premature micro-optimization
- Honest tradeoff explanations over vague "best practice" claims

## Verification

Always try to verify with the most relevant command available:

- `go test ./...` for broad behavior checks
- Focused package tests when full-suite runtime is too high
- `go test -race ./...` when concurrency is involved
- Benchmarks when the change claims performance impact
- Targeted manual requests for HTTP or RPC paths when applicable

If verification is skipped or blocked, say so explicitly and state the residual risk.

## Review Lens

When reviewing or self-checking backend code, look for:

- Leaky abstractions between layers
- Missing timeout or cancellation handling
- Ambiguous ownership of shared state
- Silent fallbacks that hide data or network failures
- Unclear error contracts
- Tests that only cover happy paths
- Optimizations that make the system harder to reason about without proof

## Communication Style

In explanations, prioritize:

- Request path or call path
- State transitions
- Failure behavior
- Why this boundary belongs here
- Why this verification is sufficient or insufficient

Do not explain only what the code does. Explain why the backend shape is defensible.

## Repo-Specific Adaptation

If the current repository is a learning project or systems exercise, treat it as a chance to surface backend concepts explicitly.

For cache, RPC, or distributed exercises, pay extra attention to:

- request coalescing
- duplicate suppression
- consistency of key routing
- stale or missing data semantics
- observability around remote fetches
- behavior under partial node failure

For this kind of repository, do not blindly push "enterprise" layering. Keep the design lean, but still insist on clean boundaries and testable logic.

## Output Expectations

When finishing work with this skill:

- Summarize the backend decision, not just the file diff
- Mention what was verified
- Call out any unresolved correctness or production risks
- If useful for training, mention one alternative design and why it was not chosen
