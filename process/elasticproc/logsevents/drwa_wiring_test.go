package logsevents

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/stretchr/testify/require"
)

func TestExtractDataFromLogsDRWAAuthorizedEmitterPopulatesPreparedResults(t *testing.T) {
	t.Parallel()

	const (
		txHash     = "tx-hash"
		blockHash  = "block-hash"
		blockRound = uint64(77)
	)

	args := createMockArgs()
	args.DRWAAuthorizedEmitters = [][]byte{[]byte("authorized-drwa")}
	proc, err := NewLogsAndEventsProcessor(args)
	require.NoError(t, err)

	results := proc.ExtractDataFromLogs(
		[]*transaction.LogData{drwaIdentityLog(txHash, []byte("authorized-drwa"))},
		&data.PreparedResults{Transactions: []*data.Transaction{{Hash: txHash}}},
		2,
		3,
		1234000,
		blockHash,
		blockRound,
	)

	require.Len(t, results.DrwaIdentities, 1)
	require.Equal(t, "erd1subject", results.DrwaIdentities[0].Subject)
	require.Equal(t, drwaIdentityRegisteredEvent, results.DrwaIdentities[0].EventType)
	require.Equal(t, blockHash, results.DrwaIdentities[0].BlockHash)
	require.Equal(t, blockRound, results.DrwaIdentities[0].BlockRound)
	require.Equal(t, uint32(2), results.DrwaIdentities[0].ShardID)
	require.Equal(t, 0, results.DrwaIdentities[0].EventOrder)
	require.True(t, results.DBLogs[0].Events[0].Identifier == drwaIdentityRegisteredEvent)
}

func TestExtractDataFromLogsDRWANilEmittersFailsClosed(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	proc, err := NewLogsAndEventsProcessor(args)
	require.NoError(t, err)

	results := proc.ExtractDataFromLogs(
		[]*transaction.LogData{drwaIdentityLog("tx-hash", []byte("unauthorized"))},
		&data.PreparedResults{Transactions: []*data.Transaction{{Hash: "tx-hash"}}},
		2,
		3,
		1234000,
		"block-hash",
		77,
	)

	require.Empty(t, results.DrwaIdentities)
	require.NotEmpty(t, results.DBLogs)
}

func TestExtractDataFromLogsMRVAuthorizedEmitterPopulatesPreparedResults(t *testing.T) {
	t.Parallel()

	const txHash = "mrv-hash"

	args := createMockArgs()
	args.MRVAuthorizedEmitters = [][]byte{[]byte("authorized-mrv")}
	proc, err := NewLogsAndEventsProcessor(args)
	require.NoError(t, err)

	results := proc.ExtractDataFromLogs(
		[]*transaction.LogData{
			{
				TxHash: txHash,
				Log: &transaction.Log{
					Address: []byte("authorized-mrv"),
					Events:  []*transaction.Event{mrvAnchoredEvent()},
				},
			},
		},
		&data.PreparedResults{Transactions: []*data.Transaction{{Hash: txHash}}},
		2,
		3,
		1234000,
		"mrv-block",
		88,
	)

	require.Len(t, results.MrvAnchoredProofs, 1)
	require.Equal(t, "report-1", results.MrvAnchoredProofs[0].ReportID)
	require.Equal(t, "mrv-block", results.MrvAnchoredProofs[0].BlockHash)
	require.Equal(t, uint64(88), results.MrvAnchoredProofs[0].BlockRound)
}

func drwaIdentityLog(txHash string, emitter []byte) *transaction.LogData {
	return &transaction.LogData{
		TxHash: txHash,
		Log: &transaction.Log{
			Address: emitter,
			Events: []*transaction.Event{
				{
					Identifier: []byte(drwaIdentityRegisteredEvent),
					Topics: [][]byte{
						[]byte("erd1subject"),
						[]byte("US"),
						[]byte("company"),
					},
				},
			},
		},
	}
}
