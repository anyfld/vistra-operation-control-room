package config

import (
	"github.com/kelseyhightower/envconfig"
)

type LLMConfig struct {
	ProjectID string `split_words:"true"`
	Location  string `default:"us-central1"`
	ModelName string `default:"gemini-2.0-flash" split_words:"true"`
}

type GlobalConfig struct {
	WebrtcUrlTemplate string `split_words:"true"`
}

func LoadLLMConfig() (LLMConfig, error) {
	var cfg LLMConfig
	err := envconfig.Process("llm", &cfg)

	return cfg, err
}

func GetGlobalConfig() (GlobalConfig, error) {
	var cfg GlobalConfig
	err := envconfig.Process("", &cfg)

	return cfg, err
}
