import argparse
import asyncio
import getpass

from sqlalchemy import select

from app.core.config import get_settings
from app.core.security import hash_password
from app.db.base import Base
from app.db.session import SessionLocal, engine
from app.models import Admin, Setting


async def _init_schema() -> None:
    async with engine.begin() as connection:
        await connection.run_sync(Base.metadata.create_all)


async def reset_admin(username: str, password: str) -> None:
    await _init_schema()
    async with SessionLocal() as session:
        result = await session.execute(select(Admin).where(Admin.username == username))
        admin = result.scalar_one_or_none()
        if admin is None:
            session.add(Admin(username=username, password_hash=hash_password(password), role="owner"))
        else:
            admin.password_hash = hash_password(password)
            admin.is_active = True
        await session.commit()


async def set_setting(key: str, value: str, is_secret: bool = False) -> None:
    await _init_schema()
    async with SessionLocal() as session:
        setting = await session.get(Setting, key)
        if setting is None:
            session.add(Setting(key=key, value=value, is_secret=is_secret))
        else:
            setting.value = value
            setting.is_secret = is_secret
        await session.commit()


def main() -> None:
    parser = argparse.ArgumentParser(prog="sing-grok-admin")
    sub = parser.add_subparsers(dest="command", required=True)
    reset = sub.add_parser("reset-admin")
    reset.add_argument("--username", default=get_settings().bootstrap_admin_username)
    reset.add_argument("--password", default="")
    setting = sub.add_parser("set-setting")
    setting.add_argument("key")
    setting.add_argument("value")
    setting.add_argument("--secret", action="store_true")
    args = parser.parse_args()

    if args.command == "reset-admin":
        password = args.password or getpass.getpass("New admin password: ")
        asyncio.run(reset_admin(args.username, password))
        print(f"Admin {args.username} updated")
    elif args.command == "set-setting":
        asyncio.run(set_setting(args.key, args.value, args.secret))
        print(f"Setting {args.key} updated")


if __name__ == "__main__":
    main()
