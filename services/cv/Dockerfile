FROM python:3.14-slim

WORKDIR /app

COPY requirements.txt .

RUN sed -i 's/\r$//' requirements.txt && sed -i 's/opencv-python/opencv-python-headless/g' requirements.txt

RUN pip install --no-cache-dir --only-binary=:all: -r requirements.txt

# purge the desktop version and override with opencv-python-headless
RUN pip uninstall -y opencv-python opencv-contrib-python || true
RUN pip install --no-cache-dir --force-reinstall --no-deps --only-binary=:all: opencv-python-headless==5.0.0.93

# COPY /models ./models
RUN mkdir -p ./models
ADD https://huggingface.co/onnx-community/yolov10s/blob/main/onnx/model.onnx /app/models/yolov10s.onnx
ADD https://huggingface.co/onnx-community/yolov10s/blob/main/config.json /app/models/yolov10_config.json
COPY proto/ ./proto/
COPY *.py ./

ENV PYTHONUNBUFFERED=1

CMD ["python", "main.py"]
