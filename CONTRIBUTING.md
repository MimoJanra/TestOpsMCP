# Contributing to Allure MCP Server

Thank you for considering contributing to Allure MCP Server! We welcome improvements, bug fixes, and new features.

## How to Contribute

### 1. Report Bugs

Found a bug? Open an issue with:

- **Title:** Clear, concise description (e.g., "Crashes on invalid project_id")
- **Environment:** Go version, OS, Docker version (if applicable)
- **Steps to reproduce:** Clear steps to trigger the bug
- **Expected vs actual:** What should happen vs what does happen
- **Logs:** Any relevant log output (JSON logs preferred)

### 2. Suggest Features

Have an idea? Open an issue with `[FEATURE]` in the title:

- **Use case:** Why is this needed?
- **Proposed solution:** How should it work?
- **Alternatives:** Other approaches you considered
- **Impact:** What tools or workflows does this enable?

### 3. Submit Code

#### Fork & Clone

```bash
git clone https://github.com/YOUR_USERNAME/TestOpsMCP.git
cd TestOpsMCP
git remote add upstream https://github.com/MimoJanra/TestOpsMCP.git
```

#### Create a Feature Branch

```bash
git checkout -b feature/my-feature
# or
git checkout -b fix/my-bug-fix
```

#### Development Workflow

```bash
# Install Go 1.22+
# https://golang.org/dl/

# Build
make build

# Run tests
make test

# Check code quality
make lint

# Format code
make fmt

# Run everything
make check
```

#### Make Your Changes

- **Keep commits atomic** — one change per commit
- **Write descriptive commit messages** — explain the WHY, not just WHAT
- **Add tests** for new features or bug fixes
- **Update docs** if behavior changes

Example commit:

```
Add support for launch auto-submission

Implements new auto_submit parameter on run_allure_launch tool.
When true, launch is auto-submitted after test collection completes.

Fixes #123
```

#### Tests

Add tests for new functionality:

```bash
# Test file: internal/tools/my_feature_test.go
go test ./internal/tools -v -run TestMyFeature
```

Coverage:

```bash
go test -cover ./...
```

#### Documentation

Update relevant docs:

- `README.md` — for major features or setup changes
- `docs/API.md` — for new tools or endpoints
- `docs/DEPLOYMENT.md` — for new deployment patterns
- Code comments — only if non-obvious

### 4. Submit a Pull Request

#### Before You Push

```bash
# Make sure tests pass
make check

# Update documentation
# Edit README.md, docs/API.md, etc.

# Rebase on main
git fetch upstream
git rebase upstream/main

# Push to your fork
git push origin feature/my-feature
```

#### Create the PR

On GitHub, click "New Pull Request" and:

- **Title:** Clear, concise (e.g., "Add launch auto-submission feature")
- **Description:** Use the template:

```markdown
## Summary
Brief description of what this PR does.

## Changes
- Change 1
- Change 2

## Testing
How can reviewers test this?

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes
- [ ] Follows code style (make fmt)
```

#### Code Review

- Respond to feedback
- Make requested changes in new commits (don't force-push during review)
- Once approved, maintainer will squash and merge

---

## Code Style

### Go

Follow [Go conventions](https://golang.org/doc/effective_go):

```bash
# Auto-format
make fmt

# Check
make lint
```

### Commit Messages

**Format:**

```
[Type] Concise summary

Detailed explanation if needed. Explain the WHY.

Fixes #123
Related: #456
```

**Types:**
- `feat` — New feature
- `fix` — Bug fix
- `docs` — Documentation
- `refactor` — Code refactoring
- `test` — Tests
- `chore` — Build, CI, deps

**Examples:**

```
feat: Add launch auto-submission parameter

fix: Handle invalid project_id without crashing

docs: Update API.md with auto_submit examples

test: Add coverage for launch polling
```

### Variable Naming

```go
// Good
var launchID int
var apiToken string
var isRunning bool

// Avoid
var l int
var token_val string
var running string
```

### Comments

Only comment non-obvious code:

```go
// Good: explains WHY
// Retry with exponential backoff; Allure API has transient failures
for i := 0; i < maxRetries; i++ {

// Bad: obvious from code
// Increment counter
count++
```

---

## Development Setup

### Prerequisites

- Go 1.22+
- Make
- Git

### IDE Setup

#### VS Code

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go"
  }
}
```

#### GoLand / IntelliJ

- Install Go plugin
- Mark `internal/` as Sources
- Enable gofmt on save

### Docker Development

Build & test locally:

```bash
docker build -t allure-mcp-dev .
docker run -it -e ALLURE_BASE_URL=http://localhost:4040 \
           -e ALLURE_TOKEN=test \
           allure-mcp-dev
```

---

## Common Tasks

### Add a New Tool

1. Define the tool in `internal/tools/registry.go`
2. Add handler in `internal/tools/registry.go`
3. Add tests in `internal/tools/registry_test.go`
4. Document in `docs/API.md`

### Add a New Endpoint

1. Add handler in `internal/mcp/server.go` (HTTP mode)
2. Add tests in `internal/mcp/server_test.go`
3. Document in `docs/API.md`

### Update Configuration

1. Add env var to `internal/config/config.go`
2. Add to `.env.example`
3. Document in `docs/INSTALLATION.md`

---

## Maintainer Guidelines

### Release

```bash
# Tag and push
git tag -a v1.1.0 -m "Release 1.1.0"
git push upstream v1.1.0

# Build and push Docker image
docker build -t registry.example.com/allure-mcp:1.1.0 .
docker push registry.example.com/allure-mcp:1.1.0
```

### PR Review Checklist

- [ ] Tests pass locally
- [ ] Code follows style (make lint, make fmt)
- [ ] Documentation updated
- [ ] Commit messages are clear
- [ ] No unnecessary dependencies added
- [ ] No secrets committed
- [ ] Backward compatible (or documented breaking change)

---

## Questions?

- Open a GitHub Discussion
- Email: mimojanra@gmail.com
- Check existing issues/PRs for similar topics

---

## License

By contributing, you agree your code is licensed under [Apache License 2.0](LICENSE).

Thank you for making Allure MCP Server better! 🎉
