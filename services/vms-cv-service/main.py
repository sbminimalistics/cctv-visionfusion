import cv2
import numpy as np
import os
import sys

from rabbitmq_amqp_python_client import Environment, AMQPMessagingHandler, AddressHelper

class ImageMessageHandler(AMQPMessagingHandler):
    def __init__(self):
        super().__init__()

    def on_message(self, event):
        try:
            # Extract raw binary message body bytes from the event frame
            body = event.message.body
            print(" [x] Received raw image payload. Decoding...", flush=True)
            
            # Convert binary data directly to numpy format
            nparr = np.frombuffer(body, np.uint8)
            img = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
            
            if img is not None:
                h, w, c = img.shape
                print(f" [✓] AMQP 1.0 Parsed JPEG: {w}x{h} ({c} channels)", flush=True)
                # Put your OpenCV 5.0 logic here (e.g., cv2.imwrite, analytics)
            else:
                print(" [!] Failed to decode incoming byte stream.", flush=True)
                
            # Settle/Acknowledge the message receipt explicitly back to the broker
            self.delivery_context.accept(event)
            
        except Exception as e:
            print(f" [!] Error processing message inside callback: {e}", flush=True)


def main():
    rabbitmq_host = os.getenv('RABBITMQ_HOST', 'localhost')
    rabbitmq_port = os.getenv('RABBITMQ_PORT', '5672')
    rabbitmq_vhost = os.getenv('RABBITMQ_VHOST', 'vms')
    queue_name = os.getenv('RABBITMQ_QUEUE', 'image_queue')
    rabbitmq_user = os.getenv('DOWNLOADER_RMQ_USER', 'a')
    rabbitmq_pass = os.getenv('DOWNLOADER_RMQ_PASS', 'b')

    print(rabbitmq_host, flush=True)
    print(rabbitmq_port, flush=True)
    print(rabbitmq_vhost, flush=True)
    print(queue_name, flush=True)
    print(rabbitmq_user, flush=True)
    print(rabbitmq_pass, flush=True)

    print(f"Starting OpenCV 5.0 worker via AMQP 1.0...", flush=True)

    try:
        # Initialize the AMQP 1.0 Environment and open a connection
        env = Environment(uri=f"amqp://{rabbitmq_user}:{rabbitmq_pass}@{rabbitmq_host}:{rabbitmq_port}/{rabbitmq_vhost}")

        connection = env.connection()
        connection.dial()
        
        consumer = connection.consumer(f"/queues/{queue_name}", message_handler=ImageMessageHandler())
        print(f" [*] Listening on AMQP 1.0 Queue: '{queue_name}'...", flush=True)

    except Exception as e:
        print(f" [!] AMQP 1.1/1.0 Engine Error: {e}", flush=True)
        sys.exit(1)

if __name__ == '__main__':
    main()
