import time
import pika


class RabbitMQClient:
    """Manages connection, channel registration, and listening to RabbitMQ."""

    def __init__(self, host: str, port: str, vhost: str, user: str, password: str, queue: str):
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
            heartbeat=60
        )

        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()

        self.channel.queue_declare(queue=self.queue, passive=True)

        # give only one message to this worker at a time
        self.channel.basic_qos(prefetch_count=1)

    def _internal_callback(self, ch, method, properties, body: bytes):
        """Internal bridge method to catch delivery states and invoke external logic."""
        try:
            self.on_message_received(body)
            # acknowledge completion back to RMQ
            ch.basic_ack(delivery_tag=method.delivery_tag)
        except Exception as e:
            print(f"execution error, rejecting message: {e}")
            # reject message and put it back in queue for a retry
            ch.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

    def start_listening(self, callback):
        """Blocks and listens for incoming queue payloads."""
        self.on_message_received = callback
        print(f"Waiting for messages in queue: '{self.queue}'")
        self.channel.basic_consume(
            queue=self.queue, on_message_callback=self._internal_callback
        )
        self.channel.start_consuming()

    def close(self):
        """Gracefully closes infrastructure sockets."""
        if self.connection and self.connection.is_open:
            self.connection.close()
