import json
import logging
import sys
from datetime import datetime, timezone
from opentelemetry import trace, propagate
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.trace.propagation.tracecontext import TraceContextTextMapPropagator


class JSONTraceFormatter(logging.Formatter):
    def __init__(self, service_name: str):
        super().__init__()
        self.service_name = service_name

    def format(self, record: logging.LogRecord) -> str:
        log_entry = {
            "time": datetime.now(timezone.utc).astimezone().isoformat(),
            "level": record.levelname,
            "msg": record.getMessage(),
            "service": self.service_name,
        }

        # automatically extract active trace_id from current OTel context!
        span_ctx = trace.get_current_span().get_span_context()
        if span_ctx and span_ctx.is_valid:
            log_entry["trace_id"] = f"{span_ctx.trace_id:032x}"

        # include any extra structured fields passed via extra={...}
        if hasattr(record, "extra_fields") and isinstance(record.extra_fields, dict):
            log_entry.update(record.extra_fields)

        if record.exc_info:
            log_entry["error"] = self.formatException(record.exc_info)

        return json.dumps(log_entry)


def init_telemetry(service_name: str = "cctv-CV") -> None:
    # initialize TracerProvider & W3C TraceContext propagator
    provider = TracerProvider()
    trace.set_tracer_provider(provider)
    propagate.set_global_textmap(TraceContextTextMapPropagator())

    # instruct root logger to output structured JSON with service & trace_id
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JSONTraceFormatter(service_name))

    root_logger = logging.getLogger()
    root_logger.handlers.clear()
    root_logger.addHandler(handler)
    root_logger.setLevel(logging.INFO)
