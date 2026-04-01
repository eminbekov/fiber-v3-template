package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type queryTraceContextKey struct{}

var queryTraceKey = queryTraceContextKey{}

type queryTraceState struct {
	startTime time.Time
	sql       string
}

// queryTracer implements pgx.QueryTracer for slow-query and failure logging (guide §19.2).
type queryTracer struct {
	slowThreshold time.Duration
}

func (tracer *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	state := queryTraceState{
		startTime: time.Now(),
		sql:       data.SQL,
	}
	return context.WithValue(ctx, queryTraceKey, state)
}

func (tracer *queryTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	traceState, ok := ctx.Value(queryTraceKey).(queryTraceState)
	if !ok {
		return
	}

	duration := time.Since(traceState.startTime)
	sqlText := traceState.sql

	if data.Err != nil {
		slog.ErrorContext(ctx, "database query failed",
			"query", sqlText,
			"duration_ms", duration.Milliseconds(),
			"error", data.Err,
		)
		return
	}

	if duration > tracer.slowThreshold {
		slog.WarnContext(ctx, "slow database query",
			"query", sqlText,
			"duration_ms", duration.Milliseconds(),
			"rows_affected", data.CommandTag.RowsAffected(),
		)
		return
	}

	slog.DebugContext(ctx, "database query",
		"query", sqlText,
		"duration_ms", duration.Milliseconds(),
		"rows_affected", data.CommandTag.RowsAffected(),
	)
}
