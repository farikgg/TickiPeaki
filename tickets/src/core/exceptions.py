class BaseExceptionError(Exception):
    def __init__(self, error: str, detail: str | None = None):
        self.error = error
        self.detail = detail
        super().__init__(error)


class S3StorageServiceError(BaseExceptionError):
    def __init__(self, error: str, detail: str | None = None, filename: str | None = None):
        super().__init__(error, detail)
        self.filename = filename


class PDFGenerationServiceError(BaseExceptionError):
    def __init__(self, error: str, detail: str | None = None):
        super().__init__(error, detail)


class RedisStatusServiceError(BaseExceptionError):
    def __init__(self, error: str, detail: str | None = None):
        super().__init__(error, detail)
