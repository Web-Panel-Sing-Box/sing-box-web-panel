from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.config import Settings
from app.core.security import hash_password
from app.models import Admin, Setting


DEFAULT_SETTINGS: dict[str, tuple[str, bool]] = {
    "log_level": ("info", False),
    "public_host": ("", False),
    "clash_api_port": ("9090", False),
    "v2ray_api_port": ("8080", False),
}


async def ensure_defaults(session: AsyncSession, settings: Settings) -> None:
    for key, (value, is_secret) in DEFAULT_SETTINGS.items():
        existing = await session.get(Setting, key)
        if existing is None:
            session.add(Setting(key=key, value=value, is_secret=is_secret))

    if settings.clash_api_secret and settings.clash_api_secret != "change-me":
        existing = await session.get(Setting, "clash_api_secret")
        if existing is None:
            session.add(Setting(key="clash_api_secret", value=settings.clash_api_secret, is_secret=True))

    if settings.bootstrap_admin_password:
        result = await session.execute(select(Admin).where(Admin.username == settings.bootstrap_admin_username))
        if result.scalar_one_or_none() is None:
            session.add(
                Admin(
                    username=settings.bootstrap_admin_username,
                    password_hash=hash_password(settings.bootstrap_admin_password),
                    role="owner",
                )
            )
    await session.commit()
