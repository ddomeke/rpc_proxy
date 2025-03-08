# OpStack RPC Proxy

This is a monitoring and proxy service for the Optimism stack that:

1. Monitors L1 deposits to the OptimismPortal contract
2. Checks incoming addresses against a FrozenAccounts contract
3. Blocks deposits from frozen accounts
4. Monitors L2 deposit confirmations
5. Collects metrics for Prometheus
6. Acts as a JSON-RPC proxy for other services

## Project Structure

```
rpc_proxy/
├── cmd/
│   └── server/
│       └── main.go            # Entry point
├── internal/
│   ├── config/
│   │   └── config.go          # Configuration loading
│   ├── ethereum/
│   │   ├── client.go          # Ethereum client initialization
│   │   ├── events.go          # Event definitions and processing
│   │   └── frozen.go          # Checking frozen accounts
│   ├── metrics/
│   │   └── metrics.go         # Prometheus metrics
│   ├── monitor/
│   │   ├── l1_monitor.go      # L1 deposit monitoring
│   │   └── l2_monitor.go      # L2 confirmation monitoring
│   └── proxy/
│       ├── handlers.go        # RPC request/response handlers
│       └── server.go          # Proxy server
├── pkg/
│   └── utils/
│       └── utils.go           # Helper functions
├── .env                       # Environment variables (create from .env.example)
├── go.mod                     # Go module file
└── README.md                  # Project documentation
```

## Setup and Configuration

1. Clone the repository
2. Copy `.env.example` to `.env` and update with your values
3. Run the following commands:

```bash
# Install dependencies
go mod download

# Build the application
go build -o rpc_proxy ./cmd/server

# Run the application
./rpc_proxy
```

## Environment Variables

The following environment variables need to be set in the `.env` file:

| Name | Description |
|------|-------------|
| L1_RPC_URL | Ethereum L1 RPC URL |
| L1_RPC_URL_WS | Ethereum L1 WebSocket URL for event subscription |
| L2_RPC_URL | Optimism L2 RPC URL |
| FROZEN_CONTRACT_ADDRESS | Address of the FrozenAccounts contract |
| OPTIMISM_PORTAL_ADDRESS | Address of the OptimismPortal contract |
| PROXY_PORT | Port for the RPC proxy server (default: 8545) |
| METRICS_PORT | Port for Prometheus metrics (default: 9100) |

## Prometheus Metrics

The service exposes the following Prometheus metrics on http://localhost:{METRICS_PORT}/metrics:

| Metric | Description |
|--------|-------------|
| opstack_total_deposits | Total number of deposits through OptimismPortal |
| opstack_blocked_deposits | Total number of blocked deposits from frozen accounts |
| opstack_deposit_latency | Time taken for a deposit to reach L2 |
| opstack_deposit_value_total | Total ETH value of all deposits in wei |
| opstack_deposits_by_account | Number of deposits grouped by sender account |
| opstack_deposit_gas_limit | Distribution of gas limits for deposits |
| opstack_deposit_value | Distribution of deposit values in ETH |
| opstack_l2_deposit_confirmation_time | Time taken for deposits to be confirmed on L2 |
| opstack_deposits_pending | Number of deposits waiting for L2 confirmation |

## Usage

After starting the service, you can:

1. Use it as a drop-in replacement for your Ethereum RPC endpoint
2. Monitor the logs for deposit information and frozen account checks
3. Configure Prometheus to scrape the metrics endpoint
4. Build Grafana dashboards using the exposed metrics
