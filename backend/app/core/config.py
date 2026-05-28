from functools import lru_cache
from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """Runtime settings. Defaults are local-host first and installer-friendly."""

    model_config = SettingsConfigDict(
        env_prefix="SING_GROK_",
        env_file=".env",
        env_file_encoding="utf-8",
        extra="ignore",
    )

    app_name: str = "Sing Grok"
    api_host: str = "127.0.0.1"
    api_port: int = 8081
    web_host: str = "0.0.0.0"
    web_port: int = 3000

    runtime_dir: Path = Path("/opt/sing-grok")
    config_dir: Path = Path("/etc/sing-grok")
    data_dir: Path = Path("/var/lib/sing-grok")
    log_dir: Path = Path("/var/log/sing-grok")

    database_url: str = "sqlite+aiosqlite:////var/lib/sing-grok/panel.db"
    sing_box_binary: str = "sing-box"
    sing_box_config_path: Path = Path("/etc/sing-grok/config.json")
    sing_box_log_path: Path = Path("/var/log/sing-grok/sing-box.log")
    sing_box_pid_path: Path = Path("/var/run/sing-grok-singbox.pid")
    process_mode: str = Field(default="systemd", pattern="^(systemd|subprocess)$")
    sing_box_service_name: str = "sing-grok-singbox"

    jwt_secret: str = "change-me-before-first-run"
    jwt_issuer: str = "sing-grok"
    access_token_minutes: int = 15
    refresh_token_days: int = 14
    cookie_secure: bool = False

    clash_api_port: int = 9090
    clash_api_secret: str = "change-me"
    v2ray_api_port: int = 8080
    enable_v2ray_api: bool = True
    observatory_api_url: str = "http://127.0.0.1:9090/v1/services/observatory"
    traffic_flush_seconds: int = 30
    traffic_flush_bytes: int = 1024 * 1024

    public_host: str = ""
    bootstrap_admin_username: str = "admin"
    bootstrap_admin_password: str = ""


@lru_cache
def get_settings() -> Settings:
    return Settings()
