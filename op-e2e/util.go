package op_e2e

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
	"time"
)

// Testing interface, shared between Hive and native Go testing.
type T interface {
	require.TestingT

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
	Failed() bool
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

type SystemInterface struct {
	Keys      map[string]Key
	L2Clients map[string]L2Client

	Contracts struct {
		GasPriceOracle *bindings.GasPriceOracle
	}
}

func MakeSystemInterface() struct {
}

type Key struct {
}

type ELClient struct {
	name string
	*ethclient.Client
}

func (el *ELClient) WaitForTransaction(t T, txHash common.Hash, timeout time.Duration) *types.Receipt {
	receipt, err := waitForTransaction(txHash, el.Client, timeout)
	require.NoError(t, err, "failed waiting for tx %s from %s for time %s", txHash, el.name, timeout)
	return receipt
}

type L2Client struct {
	ELClient
}

func TestTransaction(t T, ctx context.Context, seqL2 *ethclient.Client) {

}
