

# Scoping Phase

```txt

  USER
    |
    v
  Orchestrator <-- receives request, classifies, decides path
    |                 >> event: REQUEST_CLASSIFIED
    |
    |   Orchestrator runs brainstorming itself (uses clarify for
    |   back-and-forth with user — subagents can't interact with user)
    |
    |   Brainstormer <--------> USER (back and forth via Orchestrator)
    |       |                       until user picks direction
    |       v
    +--<----+  >> event: BRAINSTORM_COMPLETE
    |
    +-- dispatch Documentation
    |       |  reads tbd, writes decision-log.md
    |       v
    +--<----+
    |
    +-- dispatch Researcher on chosen direction
    |       |  reads tbd for context
    |       |  writes docs/research-findings.md
    |       v
    +--<----+  >> event: RESEARCH_COMPLETE
    |
    +-- dispatch Documentation
    |       |  reads tbd + research-findings.md
    |       v
    +--<----+
    |
    +-- dispatch Presenter
    |       |  reads research-findings.md + tbd
    |       |  returns both tech + ELI5 summaries
    |       v
    +--<----+
    |
    |   Orchestrator presents to user (clarify: tech or ELI5?)
    |
    |   USER -> "plan it" or "need more info"
    |     |
    |     +-- "need more info" -> dispatch Researcher again
    |     |
    |     +-- "plan it" -> [to Planning Phase]

    ```

# Planning Phase

```txt

  +-- dispatch Planner
  |       |  reads tbd + research-findings.md
  |       |  writes plan to state.json
  |       v
  +--<----+
  |
  +-- dispatch Plan Reviewer
  |       |  reads state.json + tbd + existing docs
  |       v
  |   Plan Reviewer
  |       |
  |   PASS / FAIL
  |       |
  +--<----+  if FAIL -> dispatch Planner again (loop, no max)
  |
  |  if PASS -> present to user for approval
  |
  v
  Orchestrator presents plan to user (clarify)
    |                       |
  APPROVE             <feedback>
    |                       |
    v                       v
    |                 dispatch Planner -> Plan Reviewer (loop)
    |  >> event: PLAN_APPROVED
    |
    +-- dispatch Documentation
    |       |  reads tbd, writes prd.md, specs, initial todos
    |       v
    +--<----+
    |
    v
  [to Execution Phase]
```

# Execution Phase

```txt
  +-- dispatch Specialist Coders (one per framework needed)
  |       |  reads state.json for task details
  |       v
  +--<----+
  |
  +-- dispatch Code Reviewer
  |       |  reviews code, outputs severity-ranked findings
  |       |  writes docs/review-findings.md
  |       v
  +--<----+  >> event: CODE_REVIEW_COMPLETE
  |
  +-- dispatch Documentation
  |       |  reads tbd + review-findings.md
  |       v
  +--<----+
  |
  +-- dispatch Tech Writer
  |       |  reads review-findings.md
  |       |  writes docs/todos.md (ranked by severity)
  |       v
  +--<----+
  |
  +-- dispatch Bad News Presenter
  |       |  reads review-findings.md + todos.md
  |       |  returns both tech + ELI5 breakdowns with fix/defer options
  |       v
  +--<----+
  |
  |   Orchestrator presents findings to user (clarify: tech or ELI5?)
  |   Orchestrator collects fix/defer decision per issue
  |   writes decisions to decisions/fix-defer.md
  |   >> event: DECISIONS_MADE
  |
  |   dispatch Documentation
  |       |  reads tbd + fix-defer.md
  |       v
  +--<----+
  |
  |  FIX NOW issues:
  |    +-- dispatch Specialist Coders (fix)
  |    |       |  reads fix-defer.md for what to fix
  |    |       v
  |    +--<----+
  |    |
  |    +-- dispatch Code Reviewer
  |    |       |  writes updated review-findings.md
  |    |       v
  |    +--<----+  >> event: FIX_APPLIED
  |    |
  |    +-- dispatch Documentation
  |    |       |  reads tbd, updates docs
  |    |       v
  |    +--<----+
  |    |
  |    |  fixed -> done
  |    |  not fixed, loop < 3 -> dispatch Specialist Coders again
  |    |  not fixed, loop >= 3 -> present to user for escalation
  |
  v
  Done
  |  >> event: MILESTONE_CLOSED
  |
  +-- dispatch Documentation
  |       |  reads tbd, finalizes all docs, archive
  |       v
  +--<----+
  |
  v
  Orchestrator notifies USER -> IDLE
```
