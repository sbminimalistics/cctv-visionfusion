import time
import pika
from opentelemetry import trace, propagate


class RabbitMQClient:
    """Manages connection, channel registration, and listening to RabbitMQ."""

    def __init__(
        self, host: str, port: str, vhost: str, user: str, password: str, queue: str
    ):
        self.host = host
        self.port = port
        self.vhost = vhost
        self.user = user
        self.password = password
        self.queue = queue

        self.connection = None
        self.channel = None

    def connect(self):
        """Establishes connection and declares target queue."""
        credentials = pika.PlainCredentials(self.user, self.password)
        parameters = pika.ConnectionParameters(
            host=self.host,
            port=self.port,
            virtual_host=self.vhost,
            credentials=credentials,
            heartbeat=60,
        )

        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()

        self.channel.queue_declare(queue=self.queue, passive=True)

        # give only one message to this worker at a time
        self.channel.basic_qos(prefetch_count=1)

    def _internal_callback(self, ch, method, properties, body: bytes):
        """Internal bridge method to catch delivery states and invoke external logic."""
        headers = (properties.headers or {}) if properties else {}
        carrier = {
            k: v.decode("utf-8") if isinstance(v, bytes) else str(v)
            for k, v in headers.items()
        }
        parent_ctx = propagate.extract(carrier)
        tracer = trace.get_tracer(__name__)

        with tracer.start_as_current_span("rmq.consume", context=parent_ctx):
            try:
                self.on_message_received(body)
                # acknowledge completion back to RMQ
                ch.basic_ack(delivery_tag=method.delivery_tag)
            except Exception as e:
                print(f"execution error, rejecting message: {e}")
                # reject message and put it back in queue for a retry
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

    def publish(self, exchange: str, routing_key: str, body: bytes):
        """Publishes a message with the active OpenTelemetry traceparent header."""
        headers: dict[str, str] = {}
        propagate.inject(headers)

        self.channel.basic_publish(
            exchange=exchange,
            routing_key=routing_key,
            body=body,
            properties=pika.BasicProperties(
                headers=headers,
                delivery_mode=2,
            ),
        )

    def start_listening(self, callback):
        """Blocks and listens for incoming queue payloads."""
        self.on_message_received = callback
        print(f"waiting for messages in queue: '{self.queue}'")
        self.channel.basic_consume(
            queue=self.queue, on_message_callback=self._internal_callback
        )
        self.channel.start_consuming()

    def close(self):
        """Gracefully closes infrastructure sockets."""
        if self.connection and self.connection.is_open:
            self.connection.close()
