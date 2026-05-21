package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPreparedLogsResultsDRWAMRVFieldsZeroValueIsIterable(t *testing.T) {
	t.Parallel()

	var results PreparedLogsResults

	require.NotPanics(t, func() {
		for range results.DrwaDenials {
		}
		for range results.DrwaIdentities {
		}
		for range results.DrwaHolderCompliances {
		}
		for range results.DrwaAttestations {
		}
		for range results.DrwaTokenPolicies {
		}
		for range results.DrwaControlEvents {
		}
		for range results.MrvAnchoredProofs {
		}
	})

	require.Nil(t, results.DrwaDenials)
	require.Nil(t, results.DrwaIdentities)
	require.Nil(t, results.DrwaHolderCompliances)
	require.Nil(t, results.DrwaAttestations)
	require.Nil(t, results.DrwaTokenPolicies)
	require.Nil(t, results.DrwaControlEvents)
	require.Nil(t, results.MrvAnchoredProofs)
}
