FROM python:3.14-slim

WORKDIR /app

COPY requirements.txt .

RUN sed -i 's/\r$//' requirements.txt && sed -i 's/opencv-python/opencv-python-headless/g' requirements.txt

# RUN apt-get update && apt-get install -y --no-install-recommends \
#     gcc \
#     cmake \
#     swig \
#     libssl-dev \
#     pkg-config \
#     python3-dev \
#     && pip install --no-cache-dir -r requirements.txt \
#     && apt-get purge -y --auto-remove gcc cmake swig libssl-dev pkg-config python3-dev \
#     && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir --only-binary=:all: -r requirements.txt

# purge the desktop version and override with opencv-python-headless
RUN pip uninstall -y opencv-python opencv-contrib-python || true
RUN pip install --no-cache-dir --force-reinstall --no-deps --only-binary=:all: opencv-python-headless==5.0.0.93

COPY /models ./models
COPY proto/ ./proto/
COPY *.py ./

ENV PYTHONUNBUFFERED=1

CMD ["python", "main.py"]
