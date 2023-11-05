package utils

import (
	"log"
	"log/slog"

	adapter "github.com/axiomhq/axiom-go/adapters/slog"
)

func SetupLogger() *adapter.Handler {
	handler, err := adapter.New()

	if err != nil {
		log.Fatalf("Failed to initialize Axiom logger: %s", err)
	}

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return handler
}
