package logger_test

import (
	"context"
	"testing"

	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestHeadersSuite(t *testing.T) {
	t.Log("Test log lifecycle")

	t.Run("Initializatio and retrieval", func(t *testing.T) {
		t.Log("Test logger initialization")

		trace := false
		ctx := logger.InitLogger(context.Background(), &trace)

		log := ctx.Value(logger.LoggerCtxKey).(zerolog.Context).Logger()
		assert.NotNil(t, log, "Logger should be initialized and not nil")

		log = logger.GetLogger(ctx, "test-category")
		assert.NotNil(t, log, "Logger should be initialized and not nil")
	})
}
