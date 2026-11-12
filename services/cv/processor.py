from seaweed import SeaweedClient
from detector import YOLODetector
from proto import downloads_pb2
import cv2
import numpy as np


class StreamProcessor:
    def __init__(self, seaweed: SeaweedClient, detector: YOLODetector):
        self.seaweed = seaweed
        self.detector = detector

    def process_message(
        self, message_body: dict, confidence_threshold: float | None = None
    ):
        """This method handles incoming messages from your queue."""

        event = downloads_pb2.DownloadCompleteEvent()
        event.ParseFromString(message_body)
        seaweed_object_key = f"{event.prefix}/{event.object_name}{event.extension}"
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
        print(
            f"seaweed_object_key: {seaweed_object_key} dimensions: {img_width}x{img_height}"
        )

        #     # execute object detection
        #     detections = detector.detect(frame, confidence_threshold=0.6)

        predictions = self.detector.detect(
            frame, confidence_threshold=confidence_threshold
        )

        self._handle_results(predictions, seaweed_object_key)

    def _handle_results(self, detections, seaweed_object_key):
        """Helper method for printing out detection results."""

        print(f"detection found {len(detections)} objects:")
        for item in detections:
            x_left = int(item["box"][0])
            y_left = int(item["box"][1])
            x_width = int(item["box"][2])
            y_height = int(item["box"][3])

            print(
                f"  - {item['class_name']} ({item['confidence']:.2f}) at Box: [{x_left}, {y_left}, {x_width}, {y_height}]",
                flush=True,
            )
