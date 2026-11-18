package messaging

import (
	"context"
	"reflect"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type pkgAnchor struct{}

var tracer = otel.Tracer(reflect.TypeOf(pkgAnchor{}).PkgPath())

// ExtractDeliveryContext reads traceparent from AMQP headers and starts a child span.
func ExtractDeliveryContext(parentCtx context.Context, delivery amqp.Delivery, spanName string) (context.Context, trace.Span) {
	headers := delivery.Headers
	if headers == nil {
		headers = amqp.Table{}
	}

	ctx := otel.GetTextMapPropagator().Extract(parentCtx, AMQPHeaderCarrier(headers))
	return tracer.Start(ctx, spanName)
}
