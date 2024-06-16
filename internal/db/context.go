package db

import (
	"context"

	"github.com/morphy76/g-fe-server/internal/options"
)

type ContextDbKey string
type ContextDbOptionsKey string

const (
	ctx_DB_KEY         ContextDbKey        = "db"
	ctx_DB_OPTIONS_KEY ContextDbOptionsKey = "dbOptions"
)

func ExtractDbOptions(ctx context.Context) *options.DbOptions {
	return ctx.Value(ctx_DB_OPTIONS_KEY).(*options.DbOptions)
}

func InjectDbOptions(ctx context.Context, dbOptions *options.DbOptions) context.Context {
	return context.WithValue(ctx, ctx_DB_OPTIONS_KEY, dbOptions)
}

func ExtractDb(ctx context.Context) DbClient {
	return ctx.Value(ctx_DB_KEY)
}

func InjectDb(ctx context.Context, db DbClient) context.Context {
	return context.WithValue(ctx, ctx_DB_KEY, db)
}
