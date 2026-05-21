package elasticproc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/logsevents"
	"github.com/stretchr/testify/require"
)

func TestPrepareAndSaveTransactionsDataIndexesDRWAIdentityEvent(t *testing.T) {
	t.Parallel()

	ep := newDRWATestElasticProcessor(t, [][]byte{[]byte("authorized-drwa")})
	buffers := data.NewBufferSlice(data.DefaultMaxBulkSize)

	err := ep.prepareAndSaveTransactionsData(
		drwaTestHeaderData(),
		nil,
		drwaTestPool("tx-hash", []byte("authorized-drwa")),
		nil,
		buffers,
	)

	require.NoError(t, err)
	require.Len(t, buffers.Buffers(), 1)
	body := buffers.Buffers()[0].String()
	require.Contains(t, body, `"drwa-identities"`)
	require.Contains(t, body, `"txHash":"tx-hash"`)
	require.Contains(t, body, `"eventType":"drwaIdentityRegistered"`)
	require.Contains(t, body, `"blockHash":"626c6f636b2d68617368"`)
	require.Contains(t, body, `"blockRound":77`)
}

func TestPrepareAndSaveTransactionsDataDropsUnauthorizedDRWAEmitter(t *testing.T) {
	t.Parallel()

	ep := newDRWATestElasticProcessor(t, [][]byte{[]byte("authorized-drwa")})
	buffers := data.NewBufferSlice(data.DefaultMaxBulkSize)

	err := ep.prepareAndSaveTransactionsData(
		drwaTestHeaderData(),
		nil,
		drwaTestPool("tx-hash", []byte("other-drwa")),
		nil,
		buffers,
	)

	require.NoError(t, err)
	require.Equal(t, "", strings.Join(bufferStrings(buffers), ""))
}

func TestElasticProcessorFinalizedBlockUpdatesDRWAMRVIndexes(t *testing.T) {
	t.Parallel()

	args := createMockElasticProcessorArgs()
	args.EnabledIndexes = map[string]struct{}{
		dataindexer.DrwaDenialsIndex:          {},
		dataindexer.DrwaIdentitiesIndex:       {},
		dataindexer.DrwaHolderComplianceIndex: {},
		dataindexer.DrwaAttestationsIndex:     {},
		dataindexer.DrwaTokenPoliciesIndex:    {},
		dataindexer.DrwaControlEventsIndex:    {},
		dataindexer.MrvAnchoredProofsIndex:    {},
	}
	calls := make(map[string]string)
	args.DBClient = &mock.DatabaseWriterStub{
		UpdateByQueryCalled: func(index string, body *bytes.Buffer) error {
			calls[index] = body.String()
			return nil
		},
	}
	ep := newElasticsearchProcessor(args.DBClient, args)

	err := ep.FinalizedBlock(&outport.FinalizedBlock{
		ShardID:    4,
		HeaderHash: []byte{0xaa, 0xbb, 0xcc},
	})

	require.NoError(t, err)
	require.Len(t, calls, 7)
	for _, index := range drwaMRVIndexes() {
		require.Contains(t, calls, index)
		require.Contains(t, calls[index], `"blockHash":"aabbcc"`)
		require.Contains(t, calls[index], `"shardID":4`)
		require.Contains(t, calls[index], `ctx._source.isFinalized = true`)
	}
}

func TestRemoveDRWARecordsInCaseOfRevertUsesBlockHashAndShard(t *testing.T) {
	t.Parallel()

	args := createMockElasticProcessorArgs()
	args.EnabledIndexes = map[string]struct{}{
		dataindexer.DrwaIdentitiesIndex:    {},
		dataindexer.MrvAnchoredProofsIndex: {},
	}
	calls := make(map[string]string)
	args.DBClient = &mock.DatabaseWriterStub{
		DoQueryRemoveCalled: func(index string, body *bytes.Buffer) error {
			calls[index] = body.String()
			return nil
		},
	}
	ep := newElasticsearchProcessor(args.DBClient, args)

	err := ep.removeDRWARecordsInCaseOfRevert(2, "block-hash")

	require.NoError(t, err)
	require.Len(t, calls, 2)
	require.Contains(t, calls[dataindexer.DrwaIdentitiesIndex], `"blockHash":"block-hash"`)
	require.Contains(t, calls[dataindexer.DrwaIdentitiesIndex], `"shardID":2`)
	require.Contains(t, calls[dataindexer.MrvAnchoredProofsIndex], `"blockHash":"block-hash"`)
}

func newDRWATestElasticProcessor(t *testing.T, emitters [][]byte) *elasticProcessor {
	t.Helper()

	args := createMockElasticProcessorArgs()
	args.EnabledIndexes = map[string]struct{}{
		dataindexer.DrwaIdentitiesIndex: {},
	}
	args.TransactionsProc = &mock.DBTransactionProcessorStub{
		PrepareTransactionsForDatabaseCalled: func(_ []*block.MiniBlock, _ *data.HeaderData, _ *outport.TransactionPool) *data.PreparedResults {
			return &data.PreparedResults{Transactions: []*data.Transaction{{Hash: "tx-hash"}}}
		},
	}
	balanceConverter, err := converters.NewBalanceConverter(10)
	require.NoError(t, err)
	logsArgs := logsevents.ArgsLogsAndEventsProcessor{
		PubKeyConverter:        &mock.PubkeyConverterMock{},
		Marshalizer:            &mock.MarshalizerMock{},
		BalanceConverter:       balanceConverter,
		Hasher:                 &mock.HasherMock{},
		DRWAAuthorizedEmitters: emitters,
	}
	logsProc, err := logsevents.NewLogsAndEventsProcessor(logsArgs)
	require.NoError(t, err)
	args.LogsAndEventsProc = logsProc

	return newElasticsearchProcessor(args.DBClient, args)
}

func drwaTestHeaderData() *data.HeaderData {
	return &data.HeaderData{
		ShardID:        2,
		NumberOfShards: 3,
		TimestampMs:    1234000,
		Round:          77,
		HeaderHash:     []byte("block-hash"),
	}
}

func drwaTestPool(txHash string, emitter []byte) *outport.TransactionPool {
	return &outport.TransactionPool{
		Logs: []*transaction.LogData{
			{
				TxHash: txHash,
				Log: &transaction.Log{
					Address: emitter,
					Events: []*transaction.Event{
						{
							Identifier: []byte("drwaIdentityRegistered"),
							Topics: [][]byte{
								[]byte("erd1subject"),
								[]byte("US"),
								[]byte("company"),
							},
						},
					},
				},
			},
		},
	}
}

func bufferStrings(buffers *data.BufferSlice) []string {
	result := make([]string, 0, len(buffers.Buffers()))
	for _, buff := range buffers.Buffers() {
		result = append(result, buff.String())
	}

	return result
}
