package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/anyfld/vistra-operation-control-room/pkg/config"
)

func TestLoadLLMConfig_EnvVars(t *testing.T) {
	t.Setenv("LLM_PROJECT_ID", "test-proj")
	t.Setenv("LLM_LOCATION", "europe-west1")
	t.Setenv("LLM_MODEL_NAME", "custom-model")

	cfg, err := config.LoadLLMConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-proj", cfg.ProjectID)
	assert.Equal(t, "europe-west1", cfg.Location)
	assert.Equal(t, "custom-model", cfg.ModelName)
}

func TestLoadLLMConfig_Defaults(t *testing.T) {
	t.Parallel()
	require.NoError(t, os.Unsetenv("LLM_PROJECT_ID"))
	require.NoError(t, os.Unsetenv("LLM_LOCATION"))
	require.NoError(t, os.Unsetenv("LLM_MODEL_NAME"))

	cfg, err := config.LoadLLMConfig()
	require.NoError(t, err)
	assert.Empty(t, cfg.ProjectID)
	assert.Equal(t, "us-central1", cfg.Location)
	assert.Equal(t, "gemini-2.0-flash", cfg.ModelName)
}
