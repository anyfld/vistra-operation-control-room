package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
)

type LLMConfig struct {
	ProjectID string `split_words:"true"`
	Location  string `default:"us-central1"`
	ModelName string `default:"gemini-2.0-flash" split_words:"true"`
}

type ServerConfig struct {
	Port string
}

type GlobalConfig struct {
	LLM    LLMConfig
	Server ServerConfig
}

func LoadLLMConfig() (LLMConfig, error) {
	var cfg LLMConfig
	err := envconfig.Process("llm", &cfg)

	return cfg, err
}

func GetGlobalConfig() (GlobalConfig, error) {
	var llmCfg LLMConfig
	if err := envconfig.Process("llm", &llmCfg); err != nil {
		return GlobalConfig{}, err
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverCfg := ServerConfig{
		Port: port,
	}

	return GlobalConfig{
		LLM:    llmCfg,
		Server: serverCfg,
	}, nil
}
