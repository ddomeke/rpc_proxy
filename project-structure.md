# Project Directory Structure

```
rpc_proxy/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── ethereum/
│   │   ├── client.go
│   │   ├── events.go
│   │   └── frozen.go
│   ├── metrics/
│   │   └── metrics.go
│   ├── monitor/
│   │   ├── l1_monitor.go
│   │   └── l2_monitor.go
│   └── proxy/
│       ├── handlers.go
│       └── server.go
├── pkg/
│   └── utils/
│       └── utils.go
├── .env
├── go.mod
└── README.md
```
