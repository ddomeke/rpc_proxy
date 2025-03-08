package eth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ddomeke/rpc_proxy/internal/config"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Clients holds ethereum clients for L1 and L2
type Clients struct {
	L1Client   *ethclient.Client
	L2Client   *ethclient.Client
	HTTPClient *http.Client
	PortalABI  abi.ABI
}

// InitClients initializes Ethereum L1 and Optimism L2 clients
func InitClients(cfg *config.Config) (*Clients, error) {
	// L1 Client
	l1Client, err := ethclient.Dial(cfg.L1RPCURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to L1 client: %v", err)
	}

	// L2 Client
	l2Client, err := ethclient.Dial(cfg.L2RPCURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to L2 client: %v", err)
	}

	// Prepare OptimismPortal ABI
	portalABI, err := abi.JSON(strings.NewReader(OptimismPortalABI))
	if err != nil {
		return nil, fmt.Errorf("could not parse OptimismPortal ABI: %v", err)
	}

	// HTTP client
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	return &Clients{
		L1Client:   l1Client,
		L2Client:   l2Client,
		HTTPClient: httpClient,
		PortalABI:  portalABI,
	}, nil
}
