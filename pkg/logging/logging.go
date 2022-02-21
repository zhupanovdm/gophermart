package logging

import (
	"context"
	"github.com/zhupanovdm/gophermart/pkg/app"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Option func(zerolog.Logger) zerolog.Logger

const (
	// ctxKeyLogger identifies logger instance bound within request context.
	ctxKeyLogger = app.ContextKey("Logger")

	// ctxKeyCorrelationID identifies request's correlation ID.
	ctxKeyCorrelationID = app.ContextKey("CorrelationID")
)

// GetOrCreateLogger returns context bound logger.
// Creates a new one with correlation ID field than binds it to context.
func GetOrCreateLogger(ctx context.Context, options ...Option) (context.Context, zerolog.Logger) {
	if ctx == nil {
		ctx = context.Background()
	}
	if value := ctx.Value(ctxKeyLogger); value != nil {
		if logger, ok := value.(zerolog.Logger); ok {
			return ctx, ApplyOptions(logger, options...)
		}
	}

	logger := NewLogger(options...)
	return SetLogger(ctx, logger), logger
}

func WithComponent(service string) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		return logger.With().Str(ComponentKey, service).Logger()
	}
}

func WithCID(ctx context.Context) Option {
	return func(logger zerolog.Logger) zerolog.Logger {
		if value := ctx.Value(ctxKeyCorrelationID); value != nil {
			if correlationID, ok := value.(string); ok {
				return logger.With().Str(CorrelationIDKey, correlationID).Logger()
			}
		}
		return logger
	}
}

func SetIfAbsentCID(ctx context.Context, cid string) (context.Context, string) {
	if value := ctx.Value(ctxKeyCorrelationID); value != nil {
		if cid, ok := value.(string); ok {
			return ctx, cid
		}
	}
	return SetCID(ctx, cid)
}

func SetCID(ctx context.Context, cid string) (context.Context, string) {
	return context.WithValue(ctx, ctxKeyCorrelationID, cid), cid
}

func NewCID() string {
	cid, _ := uuid.NewUUID()
	return cid.String()
}

// SetLogger binds specified logger to the context.
func SetLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, logger)
}

func NewLogger(options ...Option) zerolog.Logger {
	logger := zerolog.New(os.Stdout).
		Output(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Logger()
	return ApplyOptions(logger, options...)
}

func ApplyOptions(logger zerolog.Logger, options ...Option) zerolog.Logger {
	for _, opt := range options {
		logger = opt(logger)
	}
	return logger
}
