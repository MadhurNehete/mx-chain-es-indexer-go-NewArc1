package templatesAndPolicies

import (
	"testing"

	"github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/stretchr/testify/require"
)

func TestTemplatesAndPolicyReaderNoKibana_GetElasticTemplatesAndPolicies(t *testing.T) {
	t.Parallel()

	reader := NewTemplatesAndPolicyReader(false, "", nil, nil)

	templates, policies, err := reader.GetElasticTemplatesAndPolicies()
	require.Nil(t, err)
	require.Len(t, policies, 0)
	require.Len(t, templates, 30)

	for _, index := range []string{
		dataindexer.DrwaDenialsIndex,
		dataindexer.DrwaIdentitiesIndex,
		dataindexer.DrwaHolderComplianceIndex,
		dataindexer.DrwaAttestationsIndex,
		dataindexer.DrwaTokenPoliciesIndex,
		dataindexer.DrwaControlEventsIndex,
		dataindexer.MrvAnchoredProofsIndex,
	} {
		require.Contains(t, templates, index)
	}
}
