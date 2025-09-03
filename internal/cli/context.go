package cli

import (
	"context"
	
	"github.com/tyktech/tyk-cli/pkg/types"
)

// Context keys for storing values in command context
type contextKey string

const (
	configKey       contextKey = "config"
	outputFormatKey contextKey = "outputFormat"
)

// withConfig adds configuration to the context
func withConfig(ctx context.Context, config *types.Config) context.Context {
	return context.WithValue(ctx, configKey, config)
}

// GetConfigFromContext retrieves configuration from context
func GetConfigFromContext(ctx context.Context) *types.Config {
	if config, ok := ctx.Value(configKey).(*types.Config); ok {
		return config
	}
	return nil
}

// withOutputFormat adds output format to the context
func withOutputFormat(ctx context.Context, format types.OutputFormat) context.Context {
	return context.WithValue(ctx, outputFormatKey, format)
}

// GetOutputFormatFromContext retrieves output format from context
func GetOutputFormatFromContext(ctx context.Context) types.OutputFormat {
	if format, ok := ctx.Value(outputFormatKey).(types.OutputFormat); ok {
		return format
	}
	return types.OutputHuman
}