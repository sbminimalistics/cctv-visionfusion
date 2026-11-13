FROM python:3.14-slim

WORKDIR /app

COPY requirements.txt .

RUN sed -i 's/\r$//' requirements.txt && \
    sed -i 's/opencv-python/opencv-python-headless/g' requirements.txt && \
    pip install --no-cache-dir --only-binary=:all: -r requirements.txt

# COPY /models ./models
RUN mkdir -p ./models
ADD https://huggingface.co/onnx-community/yolov10s/resolve/main/onnx/model.onnx /app/models/yolov10s.onnx
ADD https://huggingface.co/onnx-community/yolov10s/resolve/main/config.json /app/models/yolov10_config.json

COPY proto/ ./proto/
COPY *.py ./

ENV PYTHONUNBUFFERED=1

CMD ["python", "main.py"]
