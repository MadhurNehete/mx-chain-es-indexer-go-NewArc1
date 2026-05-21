package factory

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateElasticProcessor(t *testing.T) {

	args := ArgElasticProcessorFactory{
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		EnabledIndexes:           []string{"blocks"},
		Denomination:             1,
	}

	ep, err := CreateElasticProcessor(args)
	require.Nil(t, err)
	require.NotNil(t, ep)
}

func TestParseConfiguredEmittersAcceptsHexAndSkipsEmptyEntries(t *testing.T) {
	t.Parallel()

	emitters, err := parseConfiguredEmitters(
		mock.NewPubkeyConverterMock(32),
		[]string{"", "  ", "0x1111111111111111111111111111111111111111111111111111111111111111"},
		"DRWA",
	)

	require.NoError(t, err)
	require.Len(t, emitters, 1)
	require.Equal(t, 32, len(emitters[0]))
}

func TestParseConfiguredEmittersRejectsInvalidHex(t *testing.T) {
	t.Parallel()

	_, err := parseConfiguredEmitters(mock.NewPubkeyConverterMock(32), []string{"0xnot-hex"}, "MRV")

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid MRV authorized emitter")
}
