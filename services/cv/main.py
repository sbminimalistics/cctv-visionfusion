import sys
from rabbitmq import RabbitMQClient
from seaweed import SeaweedClient
from detector import YOLODetector
from config import load_config
from processor import StreamProcessor


# def process_download_event(body: bytes):
#     print("process_download_event", flush=True)
#     event = downloads_pb2.DownloadCompleteEvent()
#     event.ParseFromString(body)

#     seaweed_object_key = f"{event.prefix}/{event.object_name}{event.extension}"

#     # download directly into an in-memory byte array
#     file_bytes = seaweed.download_file_to_memory(object_name=seaweed_object_key)

#     # convert raw bytes into a numpy array framework for opencv-python
#     nparr = np.frombuffer(file_bytes, np.uint8)

#     # decode the array into an OpenCV matrix using opencv-python
#     frame = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
#     # print(f"donwloaded frame size: {frame.size}", flush=True)

#     if frame is None:
#         raise RuntimeError("opencv-python failed to decode file from memory buffer.")

#     img_height, img_width, _ = frame.shape

#     # execute object detection
#     detections = detector.detect(frame, confidence_threshold=0.6)

#     # print clean parsed results
#     print(f"seaweed_object_key: {seaweed_object_key} dimensions: {img_width}x{img_height}")
#     print(f"detection found {len(detections)} objects:")
#     for item in detections:
#         x_left = int(item['box'][0])
#         y_left = int(item['box'][1])
#         x_width = int(item['box'][2])
#         y_height = int(item['box'][3])

#         print(
#             f"  - {item['class_name']} ({item['confidence']:.2f}) at Box: [{x_left}, {y_left}, {x_width}, {y_height}]"
#         )


def main():
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

    processor = StreamProcessor(seaweed=seaweed, detector=detector)

    # instantiate consumer and inject the processing logic callback
    rmq_client = RabbitMQClient(
        host=config.rmq_host,
        port=config.rmq_port,
        vhost=config.rmq_vhost,
        user=config.rmq_user,
        password=config.rmq_pass,
        queue=config.rmq_queue,
    )


    try:
        rmq_client.connect()
        rmq_client.start_listening(callback=processor.process_message)
    except KeyboardInterrupt:
        print("\nshutting down services...", flush=True)
        rmq_client.close()
        sys.exit(0)
    except Exception as e:
        print(f"critical application failure: {e}", flush=True)
        sys.exit(1)


if __name__ == "__main__":
    main()
