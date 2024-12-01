package logger

import (
	"context"
	"os"
	"time"

	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/multitenancy"

	"github.com/rs/zerolog"
)

const (
	LoggerCtxKey common.CtxKey = "Logger"
	t0CtxKey     common.CtxKey = "T0"
)

// T0 returns the time the context was created
func t0(ctx context.Context) time.Time {
	return ctx.Value(t0CtxKey).(time.Time)
}

// InjectLogger adds the logger to the context
func InjectLogger(ctx context.Context, appContext context.Context) context.Context {
	logger := appContext.Value(LoggerCtxKey).(zerolog.Context)
	t0 := appContext.Value(t0CtxKey).(time.Time)
	rv := context.WithValue(ctx, LoggerCtxKey, logger)
	rv = context.WithValue(rv, t0CtxKey, t0)
	return rv
}

// InitLogger creates a Context with a new Logger
func InitLogger(ctx context.Context, trace *bool) context.Context {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if *trace {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	useLoggerBuilder := zerolog.New(os.Stdout).With()

	rv := context.WithValue(ctx, LoggerCtxKey, useLoggerBuilder)
	rv = context.WithValue(rv, t0CtxKey, time.Now())

	return rv
}

// GetLogger creates a new contextual logger
func GetLogger(ctx context.Context, category string) zerolog.Logger {
	startTime := t0(ctx)
	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		e.Dict("timing", zerolog.Dict().
			Timestamp().
			Dur("since_start", time.Since(startTime)/1000),
		)
	})

	useLoggerBuilder := ctx.Value(LoggerCtxKey).(zerolog.Context).Logger().With().Str("category", category)
	useLoggerBuilder = addOwnerInfo(ctx, useLoggerBuilder)

	return useLoggerBuilder.Logger().Hook(hook)
}

func addOwnerInfo(ctx context.Context, builder zerolog.Context) zerolog.Context {
	tenant, found := multitenancy.ExtractTenant(ctx)
	if found {
		return builder.Dict("owner", zerolog.Dict().
			Str("tenant_id", tenant.TenantID).
			Str("subscription_id", tenant.SubscriptionID).
			Str("group_id", tenant.GroupID).
			Bool("system", false),
		)
	}
	return builder.Dict("owner", zerolog.Dict().
		Bool("system", true),
	)
}
