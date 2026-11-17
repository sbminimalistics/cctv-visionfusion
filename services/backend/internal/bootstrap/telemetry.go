package bootstrap

import (
	"context"
	"log/slog"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"sbminimalistics.com/cctv-visionfusion/internal/messaging"
)

type traceHandler struct {
	slog.Handler
}

var logServiceName string

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.HasTraceID() {
		r.AddAttrs(slog.String("trace_id", spanCtx.TraceID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

func InitTelemetry(serviceName string) func() {
	logServiceName = serviceName
	tp := sdktrace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
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

// ExtractDeliveryContext reads traceparent from AMQP headers and starts a child span.
func ExtractDeliveryContext(parentCtx context.Context, delivery amqp.Delivery, spanName string) (context.Context, trace.Span) {
	headers := delivery.Headers
	if headers == nil {
		headers = amqp.Table{}
	}

	ctx := otel.GetTextMapPropagator().Extract(parentCtx, messaging.AMQPHeaderCarrier(headers))
	return otel.Tracer(logServiceName).Start(ctx, spanName)
}
