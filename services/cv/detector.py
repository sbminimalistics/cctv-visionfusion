import os
import sys
import json
import cv2
import numpy as np
import onnxruntime as ort


class YOLODetector:
    """Universal detector configured for explicit manual toggling between YOLOv10
    (with coco.names) and YOLO26 (with config.json).
    """

    def __init__(self, model_path: str, model_names_path: str, confidence_threshold: float):
        self.model_path = model_path
        self.model_names_path = model_names_path
        self.confidence_threshold = confidence_threshold

        if not self.model_path or not self.model_names_path:
            print("critical error: CV_MODEL_PATH or CV_MODEL_NAMES_PATH env variables are not set.", flush=True)
            sys.exit(1)

        if not os.path.exists(self.model_path):
            print(f"critical error: model file not found at {self.model_path}", flush=True)
            sys.exit(1)

        if not os.path.exists(self.model_names_path):
            print(f"critical error: names config file not found at {self.model_names_path}", flush=True)
            sys.exit(1)

        print(f"loading unified model: {self.model_path}", flush=True)
        print(f"loading target names map: {self.model_names_path}", flush=True)

        # dynamically parse labels based on the file type (.json vs standard text list)
        self.names = {}
        if self.model_names_path.lower().endswith('.json'):
            with open(self.model_names_path, "r") as f:
                config_data = json.load(f)
            id2label = config_data.get("id2label", {})
            self.names = {int(k): v for k, v in id2label.items()}
        else:
            # fallback for plain-text lists like coco.names
            with open(self.model_names_path, "r") as f:
                lines = [line.strip() for line in f if line.strip()]
            self.names = {i: name for i, name in enumerate(lines)}

        # initialize the interchangeable ONNX Runtime session
        self.session = ort.InferenceSession(
            self.model_path, providers=["CPUExecutionProvider"]
        )

        # extract graph requirements dynamically
        self.input_name = self.session.get_inputs()[0].name
        self.output_names = [out.name for out in self.session.get_outputs()]

    def detect(self, frame: np.ndarray, confidence_threshold: float | None = None) -> list:
        """Runs object detection on the provided frame, adapting structure dynamically
        based on how the ONNX graph builds its output nodes.
        """

        if confidence_threshold is None:
            confidence_threshold = self.confidence_threshold

        orig_h, orig_w = frame.shape[:2]

        # preprocessing: Resize to match the 640x640 static graph constraint
        input_img = cv2.resize(frame, (640, 640))
        input_img = cv2.cvtColor(input_img, cv2.COLOR_BGR2RGB)
        input_img = input_img.astype(np.float32) / 255.0
        input_img = np.transpose(input_img, (2, 0, 1))
        input_tensor = np.expand_dims(input_img, axis=0)

        # execute Graph Model Inference
        outputs = self.session.run(
            self.output_names, {self.input_name: input_tensor}
        )
        parsed_results = []

        # FORMAT A: End-to-End Packed Matrix (e.g. yolov10s.onnx outputs 1 single tensor array)
        if len(outputs) == 1:
            predictions = np.squeeze(outputs[0])

            if len(predictions.shape) == 2 and predictions.shape[1] < 6:
                predictions = predictions.T

            for pred in predictions:
                if len(pred) >= 6:
                    confidence = float(pred[4])
                    if confidence < confidence_threshold:
                        continue

                    class_id = int(pred[5])
                    class_name = self.names.get(class_id, f"class-{class_id}")

                    x1, y1, x2, y2 = pred[0:4]

                    if max(x1, y1, x2, y2) <= 1.01:
                        x_min = int(x1 * orig_w)
                        y_min = int(y1 * orig_h)
                        box_w = int((x2 - x1) * orig_w)
                        box_h = int((y2 - y1) * orig_h)
                    else:
                        x_min = int(x1 * (orig_w / 640.0))
                        y_min = int(y1 * (orig_h / 640.0))
                        box_w = int((x2 - x1) * (orig_w / 640.0))
                        box_h = int((y2 - y1) * (orig_h / 640.0))

                    parsed_results.append(
                        {
                            "class_id": class_id,
                            "class_name": class_name,
                            "confidence": confidence,
                            "box": [x_min, y_min, box_w, box_h],
                        }
                    )

        # FORMAT B: Split Matrix Output Array Structure (e.g. yolo26n.onnx outputs 2 split tensors)
        else:
            logits = np.squeeze(outputs[0])  # Shape: (300, 80)
            boxes = np.squeeze(outputs[1])  # Shape: (300, 4)

            # convert raw logits to probabilities (0.0 to 1.0)
            scores = 1 / (1 + np.exp(-logits))

            for i in range(len(boxes)):
                class_id = int(np.argmax(scores[i]))
                confidence = float(scores[i][class_id])

                if confidence < confidence_threshold:
                    continue

                class_name = self.names.get(class_id, f"class-{class_id}")
                cx, cy, w, h = boxes[i]

                x_min = int((cx - w / 2) * orig_w)
                y_min = int((cy - h / 2) * orig_h)
                box_w = int(w * orig_w)
                box_h = int(h * orig_h)

                parsed_results.append(
                    {
                        "class_id": class_id,
                        "class_name": class_name,
                        "confidence": confidence,
                        "box": [x_min, y_min, box_w, box_h],
                    }
                )

        return parsed_results
