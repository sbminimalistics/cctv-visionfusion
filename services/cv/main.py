import sys
from rabbitmq import RabbitMQClient
from seaweed import SeaweedClient
from detector import YOLODetector
from config import load_config
from processor import StreamProcessor
from telemetry import init_telemetry
import logging


def main():
    init_telemetry("cctv-CV")
    logging.info("starting CCTV VisionFusion CV worker")

    config = load_config()

    # Initialize the SeaweedFS client container once
    seaweed = SeaweedClient(
        host=config.seaweed_host,
        port=config.seaweed_port,
        default_bucket=config.seaweed_bucket,
        access_key=config.seaweed_access_key,
        secret_key=config.seaweed_secret_key,
    )
    detector = YOLODetector(
        model_path=config.model_path,
        model_names_path=config.model_names_path,
        confidence_threshold=config.confidence_threshold,
    )

    # instantiate consumer and inject the processing logic callback
    rmq_client = RabbitMQClient(
        host=config.rmq_host,
        port=config.rmq_port,
        vhost=config.rmq_vhost,
        user=config.rmq_user,
        password=config.rmq_pass,
        queue=config.rmq_queue,
    )

    processor = StreamProcessor(
        seaweed=seaweed,
        detector=detector,
        rmq_client=rmq_client,
    )

    try:
        rmq_client.connect()
        rmq_client.start_listening(callback=processor.process_message)
    except KeyboardInterrupt:
        logging.info("shutting down services...")
        rmq_client.close()
        sys.exit(0)
    except Exception as e:
        logging.error(f"critical application failure: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
