package actions

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"testing"
)

type l1l2Data struct {
	t   Testing
	log log.Logger

	l2Seq    *L2Sequencer
	l2Eng    *L2Engine
	l2Cl     *ethclient.Client
	l2Signer types.Signer

	l1Miner  *L1Miner
	l1Cl     *ethclient.Client
	l1Signer types.Signer
}

type MalformedDataTestCase struct {
	name     string
	sequence func(d *l1l2Data)
	submit   func(d *l1l2Data) (expectedL2SafeHead common.Hash)
}

func TestMalformedData(t *testing.T) {

	sequenceSome := func(d *l1l2Data) {
		// TODO
	}

	batchSubmitAll := func(d *l1l2Data) (expectedL2SafeHead common.Hash) {
		// TODO: create batcher, submit all
	}

	testCases := []MalformedDataTestCase{
		{name: "success", sequence: sequenceSome, submit: batchSubmitAll},
	}
}

func (tc *MalformedDataTestCase) Run(t *testing.T) {
	// TODO actors

	d := &l1l2Data{
		t:        nil,
		log:      nil,
		l2Seq:    nil,
		l2Eng:    nil,
		l2Cl:     nil,
		l2Signer: nil,
		l1Miner:  nil,
		l1Cl:     nil,
		l1Signer: nil,
	}

	tc.sequence(d)

	// TODO

	// TODO build batch txs and include in L1, using batcher channelout or otherwise

	build := func() {

	}

	// TODO L1 head signal, sync verifier

	// TODO assert verifier synced towards expectedL2SafeHead
}
