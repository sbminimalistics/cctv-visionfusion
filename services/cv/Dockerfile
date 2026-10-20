FROM python:3.15-slim

WORKDIR /app

COPY requirements.txt .

RUN sed -i 's/\r$//' requirements.txt && sed -i 's/opencv-python/opencv-python-headless/g' requirements.txt

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    cmake \
    swig \
    libssl-dev \
    pkg-config \
    python3-dev \
    && pip install --no-cache-dir -r requirements.txt \
    && apt-get purge -y --auto-remove gcc cmake swig libssl-dev pkg-config python3-dev \
    && rm -rf /var/lib/apt/lists/*

COPY main.py .

ENV PYTHONUNBUFFERED=1

CMD ["python", "main.py"]
