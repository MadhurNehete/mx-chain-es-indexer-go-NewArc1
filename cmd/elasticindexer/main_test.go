package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadClusterConfig_RequiresElasticCredentialsByDefault(t *testing.T) {
	clearElasticCredentialEnv(t)

	configPath := filepath.Join(t.TempDir(), "prefs.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[config]
    [config.elastic-cluster]
        url = "http://localhost:9200"
        username = ""
        password = ""
        allow-insecure-no-auth-dev = false
`), 0600))

	_, err := loadClusterConfig(configPath)

	require.Error(t, err)
	require.Contains(t, err.Error(), "elasticsearch username and password are required")
}

func TestLoadClusterConfig_AllowsExplicitDevNoAuthOverride(t *testing.T) {
	clearElasticCredentialEnv(t)

	configPath := filepath.Join(t.TempDir(), "prefs.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[config]
    [config.elastic-cluster]
        url = "http://localhost:9200"
        username = ""
        password = ""
        allow-insecure-no-auth-dev = true
`), 0600))

	_, err := loadClusterConfig(configPath)

	require.NoError(t, err)
}

func TestLoadClusterConfig_RejectsPartialElasticCredentials(t *testing.T) {
	clearElasticCredentialEnv(t)

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "missing password",
			username: "elastic",
			password: "",
		},
		{
			name:     "missing username",
			username: "",
			password: "secret",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "prefs.toml")
			require.NoError(t, os.WriteFile(configPath, []byte(`
[config]
    [config.elastic-cluster]
        url = "http://localhost:9200"
        username = "`+testCase.username+`"
        password = "`+testCase.password+`"
        allow-insecure-no-auth-dev = false
`), 0600))

			_, err := loadClusterConfig(configPath)

			require.Error(t, err)
			require.Contains(t, err.Error(), "elasticsearch username and password must be set together")
		})
	}
}

func TestLoadClusterConfig_AllowsElasticCredentials(t *testing.T) {
	clearElasticCredentialEnv(t)

	configPath := filepath.Join(t.TempDir(), "prefs.toml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
[config]
    [config.elastic-cluster]
        url = "http://localhost:9200"
        username = "elastic"
        password = "secret"
        allow-insecure-no-auth-dev = false
`), 0600))

	_, err := loadClusterConfig(configPath)

	require.NoError(t, err)
}

func TestLoadClusterConfig_UsesElasticCredentialsFromEnvironment(t *testing.T) {
	clearElasticCredentialEnv(t)
	t.Setenv(elasticUsernameEnv, "elastic")
	t.Setenv(elasticPasswordEnv, "secret")

	cfg, err := loadClusterConfig("./config/prefs.toml")

	require.NoError(t, err)
	require.Equal(t, "elastic", cfg.Config.ElasticCluster.UserName)
	require.Equal(t, "secret", cfg.Config.ElasticCluster.Password)
	require.False(t, cfg.Config.ElasticCluster.AllowInsecureNoAuthDev)
}

func TestLoadClusterConfig_DefaultsElasticUsernameWhenOnlyPasswordComesFromEnvironment(t *testing.T) {
	clearElasticCredentialEnv(t)
	t.Setenv(elasticPasswordEnv, "secret")

	cfg, err := loadClusterConfig("./config/prefs.toml")

	require.NoError(t, err)
	require.Equal(t, defaultElasticUsername, cfg.Config.ElasticCluster.UserName)
	require.Equal(t, "secret", cfg.Config.ElasticCluster.Password)
}

func TestLoadClusterConfig_DefaultPrefsRequireCredentials(t *testing.T) {
	clearElasticCredentialEnv(t)

	_, err := loadClusterConfig("./config/prefs.toml")

	require.Error(t, err)
	require.Contains(t, err.Error(), "elasticsearch username and password are required")
}

func clearElasticCredentialEnv(t *testing.T) {
	t.Helper()

	t.Setenv(elasticUsernameEnv, "")
	t.Setenv(elasticPasswordEnv, "")
}
