import os
from dataclasses import dataclass

@dataclass(frozen=True)
class AppConfig:
    model_path: str
    model_names_path: str
    rmq_host: str
    rmq_port: str
    rmq_vhost: str
    rmq_user: str
    rmq_pass: str
    rmq_queue: str
    seaweed_host: str
    seaweed_port: str
    seaweed_access_key: str
    seaweed_secret_key: str
    seaweed_bucket: str
    confidence_threshold: float = 0.6


def load_config() -> AppConfig:
    return AppConfig(
        model_path = os.getenv("CV_MODEL_PATH"),
        model_names_path = os.getenv("CV_MODEL_NAMES_PATH"),
        rmq_host = os.getenv('RMQ_HOST', 'localhost'),
        rmq_port = os.getenv('RMQ_PORT', '5672'),
        rmq_vhost = os.getenv('RMQ_VHOST', 'cctv-vf'),
        rmq_user = os.getenv('CV_RMQ_USER', 'a'),
        rmq_pass = os.getenv('CV_RMQ_PASS', 'b'),
        rmq_queue = os.getenv('RMQ_QUEUE_CV_FEED', 'image_refs'),
        seaweed_host = os.getenv("SEAWEED_HOST", "localhost"),
        seaweed_port = os.getenv("SEAWEED_API_PORT", "8333"),
        seaweed_bucket = os.getenv("SEAWEED_BUCKET", "downloads"),
        seaweed_access_key = os.getenv("CV_SEAWEED_USER", "a"),
        seaweed_secret_key = os.getenv("CV_SEAWEED_PASS", "b"),
    )
