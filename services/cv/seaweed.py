import os
import boto3
from botocore.config import Config


class SeaweedClient:
    """Manages the lifecycle and operations for the SeaweedFS S3 API endpoint."""

    def __init__(self):
        self.host = os.getenv("SEAWEED_HOST", "localhost")
        self.port = os.getenv("SEAWEED_API_PORT", "8333")
        self.endpoint_url = f"http://{self.host}:{self.port}"
        self.access_key = os.getenv("CV_SEAWEED_USER", "a")
        self.secret_key = os.getenv("CV_SEAWEED_PASS", "any_secret_if_unconfigured")
        self.default_bucket = os.getenv("SEAWEED_BUCKET", "downloads")

        self.client = boto3.client(
            "s3",
            endpoint_url=self.endpoint_url,
            aws_access_key_id=self.access_key,
            aws_secret_access_key=self.secret_key,
            config=Config(signature_version="s3v4", s3={"addressing_style": "path"}),
        )

    def upload_file(self, local_path: str, object_name: str, bucket_name: str = None) -> bool:
        """Uploads a local file directly into SeaweedFS."""
        bucket = bucket_name or self.default_bucket
        try:
            self.client.upload_file(local_path, bucket, object_name)
            print(f"successfully uploaded {local_path} to {bucket}/{object_name}")
            return True
        except Exception as e:
            print(f"failed to upload file to SeaweedFS: {e}")
            return False

    def download_file(self, object_name: str, local_path: str, bucket_name: str = None) -> bool:
        """Downloads an object out of SeaweedFS onto the local filesystem."""
        bucket = bucket_name or self.default_bucket
        try:
            self.client.download_file(bucket, object_name, local_path)
            print(f"successfully downloaded {bucket}/{object_name} to {local_path}")
            return True
        except Exception as e:
            print(f"failed to download file from SeaweedFS: {e}")
            return False

    def download_file_to_memory(self, object_name: str, bucket_name: str = None) -> bytes:
        """Downloads an object out of SeaweedFS directly into a memory byte string."""
        bucket = bucket_name or self.default_bucket
        try:
            response = self.client.get_object(Bucket=bucket, Key=object_name)
            return response['Body'].read()
        except Exception as e:
            print(f"failed to read file from SeaweedFS memory stream: {e}")
            raise e
