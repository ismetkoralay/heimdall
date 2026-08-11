package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"go.yaml.in/yaml/v4"
)

var (
	ErrMissingModelConfigMapEnvVar = errors.New("error env missing the model config map path")
	ErrReadingModelConfigFile      = errors.New("error reading model config file")
	ErrUnmarshalingModelConfigFile = errors.New("error unmarshaling model config file")
	ErrInvalidProviderURL          = errors.New("error invalid provider url")
	ErrUnknownProvider             = errors.New("error unknown provider in a model config")
	ErrProviderDefinedManyTimes    = errors.New("error provider defined more than once in the yaml")
	ErrModelDefinedManyTimes       = errors.New("error model defined more than once in the yaml")
)

type ModelConfig struct {
	Providers []ProviderSection `yaml:"providers"`
	Models    []ModelSection    `yaml:"models"`
}

type ProviderSection struct {
	Name    string `yaml:"name"`
	BaseURL string `yaml:"base_url"`
}

type ModelSection struct {
	Name          string `yaml:"name"`
	ProviderName  string `yaml:"provider"`
	UpstreamModel string `yaml:"upstream_model"`
	Fallback      string `yaml:"fallback,omitempty"`
}

type ModelProviderConfig struct {
	ProviderName    string
	ProviderBaseURL string
	UpstreamModel   string
	FallbackModel   string
}

func LoadModelProviderConfig() (map[string]ModelProviderConfig, error) {
	path := os.Getenv("MODEL_CONFIG_PATH")
	if path == "" {
		return nil, ErrMissingModelConfigMapEnvVar
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrReadingModelConfigFile, err)
	}

	var config ModelConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnmarshalingModelConfigFile, err)
	}

	providerMap := make(map[string]ProviderSection, len(config.Providers))
	for _, p := range config.Providers {
		if _, ok := providerMap[p.Name]; ok {
			return nil, fmt.Errorf("%w, provider name: %s", ErrProviderDefinedManyTimes, p.Name)
		}
		if u, err := url.Parse(p.BaseURL); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidProviderURL, err)
		} else if u.Scheme == "" || u.Host == "" {
			return nil, fmt.Errorf("%w: %q: either scheme or host is empty", ErrInvalidProviderURL, p.BaseURL)
		}

		providerMap[p.Name] = ProviderSection{
			Name:    p.Name,
			BaseURL: p.BaseURL,
		}
	}

	modelMap := make(map[string]ModelProviderConfig, len(config.Models))
	for _, m := range config.Models {
		if _, ok := modelMap[m.Name]; ok {
			return nil, fmt.Errorf("%w, model name: %s", ErrModelDefinedManyTimes, m.Name)
		}
		ps, ok := providerMap[m.ProviderName]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownProvider, m.ProviderName)
		}

		modelMap[m.Name] = ModelProviderConfig{
			ProviderName:    ps.Name,
			ProviderBaseURL: ps.BaseURL,
			UpstreamModel:   m.UpstreamModel,
			FallbackModel:   m.Fallback,
		}

	}
	return modelMap, nil
}
