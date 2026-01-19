# tap - Build Mode

You are implementing **tap**, a terminal-based script runner built with Go and Bubble Tea.

## Your Mission

Pick ONE task from `@fix_plan.md`, implement it fully, update the plan, commit, and exit.

## Instructions

### Phase 1: Orient (use subagents for reading)

0a. **Study** `specs/*` - understand the component you'll work on
0b. **Study** `@fix_plan.md` - find the highest priority uncompleted task
0c. **Study** `internal/` and `cmd/` - understand current state

### Phase 2: Select

1. From `@fix_plan.md`, choose the **first uncompleted task** (marked `[ ]`)
2. Before implementing, **search the codebase** to confirm it's not already done
   - Don't assume functionality is missing
   - Use grep/find to verify

### Phase 3: Implement

3. Implement the task completely:
   - Create/modify only the files needed for THIS task
   - Follow patterns established in existing code
   - Write tests for new functionality
   - Run `go build ./...` to verify compilation
   - Run `go test ./...` to verify tests pass

### Phase 4: Update & Commit

4. Update `@fix_plan.md`:
   - Mark completed task with `[x]`
   - Add any discovered subtasks or bugs
   - Move `← CURRENT` marker to next task

5. Commit all changes:
   ```bash
   git add -A
   git commit -m "feat(component): brief description of what was done"
   ```

6. **Exit** - let the loop restart with fresh context

## Go/Bubble Tea Guidelines

When writing Bubble Tea code:
- One component per file in `internal/tui/`
- Define styles in `internal/tui/styles.go`
- Handle `tea.WindowSizeMsg` in every component
- Use `tea.Cmd` for async operations, never block in `Update()`
- Prefer `lipgloss` for all styling

When writing tests:
- Place tests in `*_test.go` next to implementation
- Use `testdata/` for fixture files
- Aim for table-driven tests

## Critical Rules

99999. **One task per iteration** - Don't scope creep
999999. **Search before implementing** - Don't rebuild existing code
9999999. **Tests must pass** - Don't commit broken code
99999999. **Update the plan** - Mark your task complete before exit

## File Reference

| Spec | Covers |
|------|--------|
| specs/01-scanner.md | Metadata parsing, directory scanning |
| specs/02-config.md | Configuration, XDG paths |
| specs/03-tui-menu.md | Menu navigation, filtering |
| specs/04-tui-params.md | Parameter input forms |
| specs/05-executor.md | Script execution |
| specs/06-cli.md | Cobra commands |
| specs/07-scaffolding.md | `tap new` command |

## Exit Checklist

Before exiting, verify:
- [ ] Task is fully implemented
- [ ] Code compiles (`go build ./...`)
- [ ] Tests pass (`go test ./...`)
- [ ] `@fix_plan.md` is updated
- [ ] Changes are committed

Then exit cleanly.
