import pino from "pino";
import { trace } from "@opentelemetry/api";

export const logger = pino(
    {
        level: "debug", //process.env.LOG_LEVEL || "info",
        mixin() {
            const span = trace.getActiveSpan();
            if (!span) return {};
            const { traceId, spanId, traceFlags } = span.spanContext();
            return {
                trace_id: traceId,
                span_id: spanId,
                trace_flags: `0${traceFlags.toString(16)}`,
            };
        },
    },
    pino.destination({ sync: true }),
);
