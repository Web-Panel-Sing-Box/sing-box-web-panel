import asyncio
import hashlib
import json
import os
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from tempfile import NamedTemporaryFile
from typing import Any

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from app.core.config import Settings
from app.models import ConfigRevision, Inbound, Setting


@dataclass(slots=True)
class ConfigGenerationResult:
    path: Path
    checksum: str
    revision_id: int
    validated: bool


class ConfigValidationError(RuntimeError):
    pass


class LocalConfigGenerator:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings

    async def generate(self, session: AsyncSession, *, apply: bool = True) -> ConfigGenerationResult:
        config = await self._build_config(session)
        encoded = json.dumps(config, ensure_ascii=False, indent=2, sort_keys=True).encode()
        checksum = hashlib.sha256(encoded).hexdigest()
        revision = ConfigRevision(checksum=checksum, config_json=config)
        session.add(revision)
        await session.flush()

        target = self.settings.sing_box_config_path
        target.parent.mkdir(parents=True, exist_ok=True)
        tmp_path = await asyncio.to_thread(self._write_temp, target.parent, encoded)
        validated = False
        try:
            await self._validate(tmp_path)
            validated = True
            if apply:
                await asyncio.to_thread(self._atomic_replace, tmp_path, target)
                revision.status = "applied"
                revision.applied_at = datetime.now(UTC)
            else:
                revision.status = "validated"
        except Exception as exc:
            revision.status = "failed"
            revision.error = str(exc)
            try:
                Path(tmp_path).unlink(missing_ok=True)
            finally:
                raise
        finally:
            await session.commit()
        return ConfigGenerationResult(target, checksum, revision.id, validated)

    async def _build_config(self, session: AsyncSession) -> dict[str, Any]:
        settings = await self._settings_map(session)
        result = await session.execute(
            select(Inbound)
            .where(Inbound.status == "active")
            .options(selectinload(Inbound.users))
            .order_by(Inbound.id)
        )
        inbounds = [self._inbound_to_config(inbound) for inbound in result.scalars().unique()]
        clash_port = int(settings.get("clash_api_port", str(self.settings.clash_api_port)))
        v2ray_port = int(settings.get("v2ray_api_port", str(self.settings.v2ray_api_port)))
        clash_secret = settings.get("clash_api_secret", self.settings.clash_api_secret)

        config: dict[str, Any] = {
            "log": {"level": settings.get("log_level", "info"), "timestamp": True},
            "inbounds": inbounds,
            "outbounds": [{"type": "direct", "tag": "direct"}],
            "route": {"final": "direct"},
            "experimental": {
                "cache_file": {
                    "enabled": True,
                    "path": str(self.settings.data_dir / "sing-box-cache.db"),
                },
                "clash_api": {
                    "external_controller": f"127.0.0.1:{clash_port}",
                    "secret": clash_secret,
                    "access_control_allow_origin": ["http://127.0.0.1", "http://localhost"],
                    "access_control_allow_private_network": False,
                },
            },
        }
        if self.settings.enable_v2ray_api:
            config["experimental"]["v2ray_api"] = {
                "listen": f"127.0.0.1:{v2ray_port}",
                "stats": {
                    "enabled": True,
                    "inbounds": [inbound["tag"] for inbound in inbounds],
                    "users": [
                        user["name"]
                        for inbound in inbounds
                        for user in inbound.get("users", [])
                        if "name" in user
                    ],
                },
            }
        return config

    async def _settings_map(self, session: AsyncSession) -> dict[str, str]:
        result = await session.execute(select(Setting))
        return {setting.key: setting.value for setting in result.scalars()}

    def _inbound_to_config(self, inbound: Inbound) -> dict[str, Any]:
        active_users = [user for user in inbound.users if user.status == "active"]
        config: dict[str, Any] = {
            "type": inbound.protocol,
            "tag": inbound.tag,
            "listen": inbound.listen,
            "listen_port": inbound.port,
        }
        if inbound.protocol == "vless":
            config["users"] = [
                {
                    "name": user.username,
                    "uuid": user.uuid,
                    **({"flow": "xtls-rprx-vision"} if inbound.reality_enabled else {}),
                }
                for user in active_users
            ]
        elif inbound.protocol == "trojan":
            config["users"] = [
                {"name": user.username, "password": user.password or user.uuid} for user in active_users
            ]
        elif inbound.protocol == "shadowsocks":
            config.update(
                {
                    "method": inbound.options.get("method", "2022-blake3-aes-128-gcm"),
                    "users": [
                        {"name": user.username, "password": user.password or user.uuid}
                        for user in active_users
                    ],
                }
            )
        else:
            config["users"] = [
                {"name": user.username, "password": user.password or user.uuid} for user in active_users
            ]
        if inbound.tls_enabled or inbound.reality_enabled:
            config["tls"] = self._tls_config(inbound)
        if inbound.transport:
            config["transport"] = inbound.transport
        if inbound.multiplex:
            config["multiplex"] = inbound.multiplex
        for key, value in inbound.options.items():
            if key not in config and key not in {"method"}:
                config[key] = value
        return config

    def _tls_config(self, inbound: Inbound) -> dict[str, Any]:
        tls: dict[str, Any] = {"enabled": True}
        if inbound.server_name:
            tls["server_name"] = inbound.server_name
        if inbound.alpn:
            tls["alpn"] = inbound.alpn
        if inbound.reality_enabled:
            tls["reality"] = {
                "enabled": True,
                "handshake": {
                    "server": inbound.reality_handshake_server or inbound.server_name or "www.cloudflare.com",
                    "server_port": inbound.reality_handshake_port or 443,
                },
                "private_key": inbound.reality_private_key,
                "short_id": [inbound.reality_short_id or ""],
                "max_time_difference": "1m",
            }
        return tls

    def _write_temp(self, directory: Path, encoded: bytes) -> Path:
        with NamedTemporaryFile("wb", dir=directory, delete=False, prefix=".config.", suffix=".json") as tmp:
            tmp.write(encoded)
            tmp.flush()
            os.fsync(tmp.fileno())
            return Path(tmp.name)

    async def _validate(self, config_path: Path) -> None:
        proc = await asyncio.create_subprocess_exec(
            self.settings.sing_box_binary,
            "check",
            "-c",
            str(config_path),
            "-D",
            str(self.settings.config_dir),
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await proc.communicate()
        if proc.returncode != 0:
            output = (stderr or stdout).decode(errors="replace").strip()
            raise ConfigValidationError(output or "sing-box check failed")

    def _atomic_replace(self, tmp_path: Path, target: Path) -> None:
        if target.exists():
            backup = target.with_name(
                f"{target.name}.bak-{datetime.now(UTC).strftime('%Y%m%d%H%M%S')}"
            )
            target.replace(backup)
        os.replace(tmp_path, target)
        os.chmod(target, 0o640)
