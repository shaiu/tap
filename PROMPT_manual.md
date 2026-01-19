# tap - Build Iteration

## Task
1. Read `@fix_plan.md` → find first `[ ]` task
2. Read the relevant spec from `specs/`
3. Search codebase to confirm it's not already done
4. Implement the task fully with tests
5. Run `go build ./...` and `go test ./...`
6. Update `@fix_plan.md` → mark task `[x]`, move `← CURRENT` to next
7. Commit: `git add -A && git commit -m "feat: <what you did>"`

## Rules
- ONE task only, don't scope creep
- Search before implementing (don't rebuild existing code)
- Tests must pass before committing
- Follow patterns in existing code

## Specs Reference
| File | Topic |
|------|-------|
| specs/01-scanner.md | Metadata parser, directory scanner |
| specs/02-config.md | Config, XDG paths |
| specs/03-tui-menu.md | TUI menu, filtering |
| specs/04-tui-params.md | Parameter forms |
| specs/05-executor.md | Script execution |
| specs/06-cli.md | Cobra CLI |
| specs/07-scaffolding.md | `tap new` |
