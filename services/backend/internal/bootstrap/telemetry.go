package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"sbminimalistics.com/cctv-visionfusion/internal/config"
)

type traceHandler struct {
	slog.Handler
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{Handler: h.Handler.WithGroup(name)}
}

func InitTelemetry(serviceName string) func() {
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLogLevel(config.Config.LogLevel),
	})
	logger := slog.New(&traceHandler{Handler: jsonHandler}).With("service", serviceName)
	slog.SetDefault(logger)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(ctx); err != nil {
			slog.Error("failed to flush telemetry on shutdown", "error", err)
		}
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	default: // "warn" or unset
		return slog.LevelWarn
	}
}
