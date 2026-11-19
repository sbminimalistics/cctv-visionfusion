from seaweed import SeaweedClient
from detector import YOLODetector
from rabbitmq import RabbitMQClient
from proto import downloads_pb2, detections_pb2
import cv2
import numpy as np
import logging
from google.protobuf.timestamp_pb2 import Timestamp


class StreamProcessor:
    def __init__(self, seaweed: SeaweedClient, detector: YOLODetector, rmq_client):
        self.seaweed = seaweed
        self.detector = detector
        self.rmq_client = rmq_client

    def process_message(
        self, message_body: dict, confidence_threshold: float | None = None
    ):
        """This method handles incoming messages from rmq queue."""

        event = downloads_pb2.DownloadCompleteEvent()
        event.ParseFromString(message_body)
        seaweed_object_key = f"{event.prefix.rstrip('/')}/{event.object_name}{event.extension}"
        file_bytes = self.seaweed.download_file_to_memory(seaweed_object_key)

        # convert raw bytes into a numpy array framework for opencv-python
        nparr = np.frombuffer(file_bytes, np.uint8)

        # decode the array into an OpenCV matrix using opencv-python
        frame = cv2.imdecode(nparr, cv2.IMREAD_COLOR)
        # print(f"donwloaded frame size: {frame.size}", flush=True)

        if frame is None:
            raise RuntimeError(
                "opencv-python failed to decode file from memory buffer."
            )

        img_height, img_width, _ = frame.shape
        logging.info(
            f"seaweed_object_key: {seaweed_object_key} dimensions: {img_width}x{img_height}",
        )

        #     # execute object detection
        #     detections = detector.detect(frame, confidence_threshold=0.6)

        predictions = self.detector.detect(
            frame, confidence_threshold=confidence_threshold
        )

        self._handle_results(
            event=event,
            detections=predictions,
            seaweed_object_key=seaweed_object_key,
            img_width=img_width,
            img_height=img_height,
        )

    def _handle_results(
        self,
        event: downloads_pb2.DownloadCompleteEvent,
        detections: list[dict],
        seaweed_object_key: str,
        img_width: int,
        img_height: int,
    ):
        """Serializes detection results to protobuf and publishes to RabbitMQ."""
        logging.info(
            f"detection found {len(detections)} objects for {seaweed_object_key}"
        )

        proto_detections = []
        for item in detections:
            x_left = int(item["box"][0])
            y_left = int(item["box"][1])
            x_width = int(item["box"][2])
            y_height = int(item["box"][3])

            logging.info(
                f"detected {item['class_name']} ({item['confidence']:.2f}) at [{x_left}, {y_left}, {x_width}, {y_height}]"
            )

            proto_detections.append(
                detections_pb2.DetectedObject(
                    class_id=int(item.get("class_id", 0)),
                    class_name=item["class_name"],
                    confidence=float(item["confidence"]),
                    box=detections_pb2.BoundingBox(
                        x_left=x_left,
                        y_left=y_left,
                        width=x_width,
                        height=y_height,
                    ),
                )
            )

        processed_at = Timestamp()
        processed_at.GetCurrentTime()

        out_msg = detections_pb2.DetectionCompletedEvent(
            event_uuid=event.event_uuid,
            model_name=self.detector.model_name,
            detections=proto_detections,
            frame=detections_pb2.FrameMetadata(
                seaweed_object_key=seaweed_object_key,
                image_width=img_width,
                image_height=img_height,
            ),
            processed_at=processed_at,
        )

        self.rmq_client.publish(
            exchange="detections.completed",
            routing_key="backend.detections.completed",
            body=out_msg.SerializeToString(),
        )
        logging.info(f"published detection results for event {event.event_uuid}")
