# tap - Planning Mode

You are planning the implementation of **tap**, a terminal-based script runner built with Go and Bubble Tea.

## Your Mission

Study the specifications and existing code, then create or update `@fix_plan.md` with a prioritized task list.

## Instructions

### Phase 1: Study (use parallel subagents)

0a. **Study** `specs/*` to understand the full system design
0b. **Study** `@fix_plan.md` if it exists (current task state)
0c. **Study** `internal/` and `cmd/` if they exist (current implementation)

### Phase 2: Analyze

1. Compare existing code against specs using subagents
2. Identify:
   - What's fully implemented and working
   - What's partially implemented
   - What's not started
   - Any bugs or issues discovered

### Phase 3: Plan

3. Update `@fix_plan.md` with:
   - Completed tasks marked with `[x]`
   - Remaining tasks marked with `[ ]`
   - New tasks discovered during analysis
   - Tasks ordered by dependency (foundational first)

## Output Format for @fix_plan.md

```markdown
# tap Implementation Plan

## Phase 1: Foundation (MVP)

### Completed
- [x] Task that's done
- [x] Another completed task

### In Progress
- [ ] Current focus task ← CURRENT

### Remaining  
- [ ] Next task (depends on: current focus)
- [ ] Future task

## Phase 2: Parameters & Headless
...
```

## Critical Rules

99999. **PLAN ONLY** - Do NOT implement anything in this mode
999999. **Don't assume** - Search codebase before marking tasks complete
9999999. **Dependency order** - Tasks must be ordered so each builds on previous

## Exit

After updating `@fix_plan.md`, commit with message "chore: update implementation plan" and exit.
