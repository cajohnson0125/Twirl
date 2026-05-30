---
status: done
order: 2
complexity: quick
skills: []
depends_on: [phase-1]
---

## Objective

Score each candidate from phase 1 against weighted criteria derived from
Twirl's requirements, produce a recommendation, and update tech-stack.org.

## Context

Twirl is an AI-assisted development tool with a dynamic orchestration engine at
its core. Phase 1 produced a research report at
`docs/research/orchestration-candidates.md`. This phase evaluates those
candidates and produces a recommendation for user approval.

## Technical Decisions

- Scoring criteria derived directly from requirements.org
- Example flow from flow-brainstorm.md used as a validation test case

## Scope

- docs/research/orchestration-recommendation.md (new) — scored evaluation and recommendation

## Acceptance Criteria

- [ ] Each candidate scored against weighted criteria (dynamic dispatch,
      parallelism, loops, streaming, interrupts, state persistence, Go fit)
- [ ] Clear recommendation with rationale
- [ ] Deal-breakers called out explicitly
- [ ] Recommendation accounts for the general-purpose nature of the engine —
      not just the example flow
- [ ] Recommendation presented for user approval before any project docs are
      modified

## Constraints

- No implementation — documentation only
- Must produce a single clear recommendation, not a tie
- Do NOT modify any project documentation (tech-stack.org, requirements.org,
  etc.) — the user approves the recommendation first

## References

- docs/research/orchestration-candidates.md (phase 1 output)
- docs/project/tech-stack.org (read only — current orchestration section)
