# Mock BOSH Director

A mock BOSH Director server for testing the bosh-mcp-server with Claude Code.

## Features

- All 18 BOSH API endpoints needed by bosh-mcp-server
- Realistic sample data (3 deployments, VMs, instances, stemcells, releases)
- Task simulation with state progression (queued → processing → done)
- Destructive operations modify state (delete, recreate, start/stop)
- Self-signed TLS certificates
- Basic authentication

## Quick Start

```bash
# Build
go build -o mock-bosh-director ./cmd/mock-bosh-director

# Run with defaults (port 25555, admin/admin)
./mock-bosh-director

# Run with custom settings
./mock-bosh-director -port 8443 -username test -password secret -speed 10
```

## CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | 25555 | Port to listen on |
| `-username` | admin | Basic auth username |
| `-password` | admin | Basic auth password |
| `-tls` | true | Enable TLS with self-signed cert |
| `-speed` | 1.0 | Simulation speed multiplier |
| `-debug` | false | Enable debug logging |

## Using with bosh-mcp-server

1. Start the mock director:
   ```bash
   ./mock-bosh-director
   ```

2. Configure bosh-mcp-server to point to it:
   ```bash
   export BOSH_ENVIRONMENT=https://localhost:25555
   export BOSH_CLIENT=admin
   export BOSH_CLIENT_SECRET=admin
   export BOSH_CA_CERT=""
   ```

3. Or use the included `.claude/mcp.json` configuration for Claude Code.

## Sample Data

The mock server includes:

**Deployments:**
- `cf` - Cloud Foundry with 8 VMs (diego_cell, router, api, uaa, doppler)
- `redis` - Redis cluster with 2 VMs
- `mysql` - MySQL PXC cluster with 3 VMs

**Infrastructure:**
- 3 stemcells (ubuntu-jammy, ubuntu-bionic)
- 9 releases (cf-deployment, diego, redis, pxc, etc.)
- Cloud config, runtime configs, CPI config

**Tasks:**
- 8 historical tasks in various states (done, error)
- New tasks created by operations progress through states

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/info` | GET | Director info |
| `/deployments` | GET | List deployments |
| `/deployments/:name` | GET/DELETE | Get/delete deployment |
| `/deployments/:name/vms` | GET | List VMs |
| `/deployments/:name/instances` | GET | List instances |
| `/deployments/:name/variables` | GET | List variables |
| `/deployments/:name/jobs/:job` | PUT | Change job state |
| `/deployments/:name?state=recreate` | PUT | Recreate VMs |
| `/tasks` | GET | List tasks |
| `/tasks/:id` | GET | Get task |
| `/tasks/:id/output` | GET | Get task output |
| `/stemcells` | GET | List stemcells |
| `/releases` | GET | List releases |
| `/configs` | GET | Get configs (cloud/runtime/cpi) |
| `/locks` | GET | List locks |

## Testing

```bash
# Run all tests
go test ./... -v

# Test specific package
go test ./internal/mockbosh -v
```

## Project Structure

```
bosh-mock-director/
├── cmd/mock-bosh-director/
│   └── main.go           # CLI entry point
├── internal/mockbosh/
│   ├── types.go          # BOSH API types
│   ├── fixtures.go       # Sample data
│   ├── state.go          # Thread-safe state manager
│   ├── tasks.go          # Task simulation
│   ├── handlers.go       # HTTP handlers
│   ├── server.go         # HTTP server
│   └── *_test.go         # Tests
└── .claude/
    └── mcp.json          # Claude Code config
```
