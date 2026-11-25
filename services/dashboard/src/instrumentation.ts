import { NodeSDK } from "@opentelemetry/sdk-node";
import { AmqplibInstrumentation } from "@opentelemetry/instrumentation-amqplib";
import { HttpInstrumentation } from "@opentelemetry/instrumentation-http";
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http";
import { resourceFromAttributes } from "@opentelemetry/resources";
import { ATTR_SERVICE_NAME } from "@opentelemetry/semantic-conventions";

// Only export over OTLP when an endpoint is explicitly configured
// (prevents retry noise/hangs when running offline without a collector)
const traceExporter = process.env.OTEL_EXPORTER_OTLP_ENDPOINT
    ? new OTLPTraceExporter()
    : undefined;

const sdk = new NodeSDK({
    resource: resourceFromAttributes({
        [ATTR_SERVICE_NAME]:
            process.env.OTEL_SERVICE_NAME || "cctv-dashboard-bff",
    }),
    traceExporter,
    instrumentations: [new AmqplibInstrumentation(), new HttpInstrumentation()],
});

sdk.start();

process.on("SIGTERM", () => {
    sdk.shutdown().finally(() => process.exit(0));
});

