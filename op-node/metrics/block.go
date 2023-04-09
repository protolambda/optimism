package metrics

import (
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/metrics"
)

// BlockMetrics keeps metrics of the block-contents.
// Also see block-ref metrics for time/number/hash type metrics data.
type BlockMetrics struct {
	BodySizes         prometheus.Histogram
	TransactionCounts prometheus.Histogram

	GasUsed prometheus.Histogram

	TransactionSize        *prometheus.HistogramVec
	TransactionCallData    *prometheus.HistogramVec
	TransactionNonce       prometheus.Histogram
	TransactionGasLimit    prometheus.Histogram
	TransactionType        *prometheus.CounterVec
	TransactionMaxFee      prometheus.Histogram
	TransactionPriorityFee prometheus.Histogram

	BaseFees     prometheus.Histogram
	BaseFeeGauge prometheus.Gauge
}

var goerliBaseInbox = common.HexToAddress("0x8453100000000000000000000000000000000000")

const (
	l1InfoTx           = "l1_info"
	userDepositTx      = "user_deposit"
	contractDeployment = "contract"
	otherTx            = "other"
)

var sizeBuckets = []float64{0, 100, 1000, 5000, 10_000, 25_000, 50_000, 75_000, 100_000, 150_000, 200_000}

var feeBuckets = []float64{0, 0.1, 1, 6.25, 12.5, 25, 50, 100, 200, 400, 800}

func NewBlockMetrics(factory metrics.Factory, ns string, subsystem string, displayName string) *BlockMetrics {
	return &BlockMetrics{

		BodySizes: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "block_body_sizes",
			Help:      displayName + " block body size in bytes: total sum of transaction sizes, with 4 bytes overhead per tx.",
			Buckets:   sizeBuckets,
		}),

		TransactionCounts: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_counts",
			Help:      displayName + " transaction count per block",
			Buckets:   []float64{0, 1, 5, 10, 20, 40, 80, 160},
		}),

		GasUsed: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "block_gas_used",
			Help:      displayName + " total gas used in block",
		}),

		TransactionSize: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_size",
			Help:      displayName + " transaction size in bytes, tagged if recognized",
			Buckets:   sizeBuckets,
		}, []string{"tx_tag"}),

		TransactionCallData: factory.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_call_data",
			Help:      displayName + " transaction call data length in bytes, tagged if recognized",
			Buckets:   sizeBuckets,
		}, []string{"tx_tag"}),

		TransactionNonce: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_nonce",
			Help:      displayName + " transaction nonce, to detect anomalies in user transactions, e.g. new users or power-users",
			Buckets:   []float64{0, 2, 4, 10, 25, 50, 100, 1000, 10_000, 50_000},
		}),

		TransactionType: factory.NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_type",
			Help:      displayName + " transaction type usage",
		}, []string{"tx_type"}),

		TransactionMaxFee: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_max_fee",
			Help:      displayName + " transaction max fee per gas in gwei",
			Buckets:   feeBuckets,
		}),

		TransactionPriorityFee: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "transaction_priority_fee",
			Help:      displayName + " transaction priority fee per gas in gwei",
			Buckets:   feeBuckets,
		}),

		BaseFees: factory.NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "base_fees",
			Help:      displayName + " block base-fee per gas in gwei, histogram data",
			Buckets:   feeBuckets,
		}),
		BaseFeeGauge: factory.NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Subsystem: subsystem,
			Name:      "base_fee_gauge",
			Help:      displayName + "block base-fee per gas in gwei, gauge",
		}),
	}
}

func (bm *BlockMetrics) record(gasUsed uint64, baseFee *big.Int, txs types.Transactions) {
	bodySize := uint64(0)
	for i, tx := range txs {
		if tx == nil {
			continue
		}
		size := tx.Size()
		bodySize += size
		txTag := otherTx
		if tx.Type() == types.DepositTxType {
			if i == 0 {
				txTag = l1InfoTx
			} else {
				txTag = userDepositTx
			}
		} else if tx.To() != nil {
			switch *tx.To() {
			case chaincfg.Goerli.BatchInboxAddress:
				txTag = "op-inbox"
			case goerliBaseInbox:
				txTag = "base-inbox"
				// More chains (even non op-stack addresses) may be tagged here,
				// if they have a big impact on data availability.
			}
		} else {
			txTag = contractDeployment
		}
		bm.TransactionSize.WithLabelValues(txTag).Observe(float64(size))
		bm.TransactionCallData.WithLabelValues(txTag).Observe(float64(len(tx.Data())))
		bm.TransactionNonce.Observe(float64(tx.Nonce()))
		bm.TransactionType.WithLabelValues(strconv.Itoa(int(tx.Type()))).Inc()

		bm.TransactionPriorityFee.Observe(bigWeiToFloatGwei(tx.GasTipCap()))
		bm.TransactionMaxFee.Observe(bigWeiToFloatGwei(tx.GasFeeCap()))
	}
	bm.BodySizes.Observe(float64(bodySize))
	bm.TransactionCounts.Observe(float64(len(txs)))
	bm.GasUsed.Observe(float64(gasUsed))
	baseFeeFloat := bigWeiToFloatGwei(baseFee)
	bm.BaseFees.Observe(baseFeeFloat)
	bm.BaseFeeGauge.Set(baseFeeFloat)
}

func (bm *BlockMetrics) RecordBlock(info eth.BlockInfo, txs types.Transactions) {
	bm.record(info.GasUsed(), info.BaseFee(), txs)
}

func (bm *BlockMetrics) RecordExecutionPayload(payload *eth.ExecutionPayload) {
	var txs = make([]*types.Transaction, len(payload.Transactions))
	for i, encTx := range payload.Transactions {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(encTx); err != nil {
			continue
		}
		txs[i] = &tx
	}
	bm.record(uint64(payload.GasUsed), payload.BaseFeePerGas.ToBig(), txs)
}

var bigGweiUnit = new(big.Int).SetUint64(params.GWei)

func bigWeiToFloatGwei(v *big.Int) float64 {
	var quo, rem big.Int
	quo.QuoRem(v, bigGweiUnit, &rem)
	return float64(quo.Uint64()) + (float64(rem.Uint64()) / params.GWei)
}
