from aioboto3 import Session
from botocore.exceptions import ClientError

from src.core.config import AppSettings, get_settings
from src.core.exceptions import S3StorageServiceError
from src.core.logger import logger


class S3StorageService:
    def __init__(self):
        s3_settings: AppSettings = get_settings()
        self.access_key: str = s3_settings.access_key
        self.bucket: str = s3_settings.bucket
        self.endpoint_url: str = s3_settings.endpoint_url
        self.secret_key: str = s3_settings.secret_key
        self.session: Session = Session()
        self.public_url: str = s3_settings.s3_public_url

    async def upload_file(self, file: bytes, filename: str):
        async with self.session.client(
            service_name="s3",
            endpoint_url=self.endpoint_url,
            aws_access_key_id=self.access_key,
            aws_secret_access_key=self.secret_key,
        ) as s3_client:
            try:
                await s3_client.put_object(Bucket=self.bucket, Key=filename, Body=file)
                logger.info(f"Файл загружен в S3. Файл: {filename}")

                response = await s3_client.generate_presigned_url(
                    'get_object',
                    Params={
                        'Bucket': self.bucket,
                        'Key': filename,
                    },
                    ExpiresIn=3600,
                )
                logger.info(f"Сгенерированный URL: {response}")
            except ClientError as error:
                logger.error(f"Ошибка S3: {error}")
                raise S3StorageServiceError(
                    error=str(error),
                    detail="Ошибка с клиентом S3",
                    filename=filename
                )
        url = response.replace(self.endpoint_url, self.public_url)

        return url
