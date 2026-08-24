# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	rehab-followup/cmd/rehab-followup	[no test files]
ok  	rehab-followup/internal/aggregate	0.001s
ok  	rehab-followup/internal/analytics	0.001s
ok  	rehab-followup/internal/auth	0.001s
ok  	rehab-followup/internal/careplan	0.001s
ok  	rehab-followup/internal/httpapi	0.012s
ok  	rehab-followup/internal/model	0.001s
ok  	rehab-followup/internal/reminder	0.001s
ok  	rehab-followup/internal/report	0.002s
ok  	rehab-followup/internal/risk	0.002s
--- FAIL: TestWorkflow22 (0.00s)
    workflow_test.go:94: file does not exist
FAIL
FAIL	rehab-followup/internal/service	0.026s
ok  	rehab-followup/internal/settings	0.012s
ok  	rehab-followup/internal/store	0.016s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/rehab-followup): exit `0`
- Frontend build (web): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Node.js version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/rehab-followup): exit `0`
- Frontend build (web): exit `0`
