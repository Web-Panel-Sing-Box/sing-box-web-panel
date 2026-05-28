import secrets
import time
from dataclasses import dataclass
from datetime import UTC, datetime, timedelta
from typing import Any

import jwt
from passlib.context import CryptContext

from app.core.config import Settings


pwd_context = CryptContext(schemes=["argon2"], deprecated="auto")


def hash_password(password: str) -> str:
    return pwd_context.hash(password)


def verify_password(password: str, password_hash: str) -> bool:
    return pwd_context.verify(password, password_hash)


def create_token(
    *,
    subject: str,
    token_type: str,
    settings: Settings,
    expires_delta: timedelta,
    extra_claims: dict[str, Any] | None = None,
) -> str:
    now = datetime.now(UTC)
    payload: dict[str, Any] = {
        "sub": subject,
        "typ": token_type,
        "iss": settings.jwt_issuer,
        "iat": int(now.timestamp()),
        "exp": int((now + expires_delta).timestamp()),
        "jti": secrets.token_urlsafe(16),
    }
    if extra_claims:
        payload.update(extra_claims)
    return jwt.encode(payload, settings.jwt_secret, algorithm="HS256")


def decode_token(token: str, settings: Settings) -> dict[str, Any]:
    return jwt.decode(token, settings.jwt_secret, algorithms=["HS256"], issuer=settings.jwt_issuer)


@dataclass
class LoginAttempt:
    failures: int = 0
    locked_until: float = 0.0


class LoginRateLimiter:
    """Small in-memory limiter. Production deployments should stay single-node/local."""

    def __init__(self, max_failures: int = 5, lock_seconds: int = 300) -> None:
        self.max_failures = max_failures
        self.lock_seconds = lock_seconds
        self._attempts: dict[str, LoginAttempt] = {}

    def check(self, key: str) -> bool:
        attempt = self._attempts.get(key)
        if not attempt:
            return True
        return time.monotonic() >= attempt.locked_until

    def success(self, key: str) -> None:
        self._attempts.pop(key, None)

    def failure(self, key: str) -> None:
        attempt = self._attempts.setdefault(key, LoginAttempt())
        attempt.failures += 1
        if attempt.failures >= self.max_failures:
            attempt.locked_until = time.monotonic() + self.lock_seconds


login_limiter = LoginRateLimiter()
