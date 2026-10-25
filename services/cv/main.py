import cv2
import numpy as np
import os
import sys
import time

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
    rabbitmq_vhost = os.getenv('RABBITMQ_VHOST', 'cctv-vf')
    queue_name = os.getenv('RABBITMQ_QUEUE', 'image_refs')
    rabbitmq_user = os.getenv('DOWNLOADER_RMQ_USER', 'a')
    rabbitmq_pass = os.getenv('DOWNLOADER_RMQ_PASS', 'b')

    print(f"Starting OpenCV 5.0 worker via AMQP 1.0...", flush=True)

    connection = None
    env = None

    try:
        # Initialize the AMQP 1.0 Environment and open a connection
        env = Environment(uri=f"amqp://{rabbitmq_user}:{rabbitmq_pass}@{rabbitmq_host}:{rabbitmq_port}/{rabbitmq_vhost}")

        connection = env.connection()
        connection.dial()
        
        # Keep a reference to the consumer object if you need to close it explicitly
        consumer = connection.consumer(f"/queues/{queue_name}", message_handler=ImageMessageHandler())
        print(f" [*] Listening on AMQP 1.0 Queue: '{queue_name}'... Press Ctrl+C to exit.", flush=True)

        # Main worker retention block
        while True:
            time.sleep(1)

    except KeyboardInterrupt:
        print("Shutdown signal received (Ctrl+C). Cleaning up resources...", flush=True)
    except Exception as e:
        print(f" [!] AMQP Engine Error: {e}", flush=True)
        sys.exit(1)
    finally:
        # Graceful connection context tear-down 
        if connection:
            try:
                print(" [->] Closing AMQP connection...", flush=True)
                connection.close()
            except Exception:
                pass
        print(" [✓] Worker stopped cleanly.", flush=True)

if __name__ == '__main__':
    main()
