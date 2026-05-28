from datetime import datetime
from typing import Any, Literal

from pydantic import BaseModel, ConfigDict, Field


class AdminOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    id: int
    username: str
    role: str
    is_active: bool


class LoginRequest(BaseModel):
    username: str = Field(min_length=1, max_length=64)
    password: str = Field(min_length=1, max_length=256)


class LoginResponse(BaseModel):
    admin: AdminOut


class InboundBase(BaseModel):
    protocol: str = Field(default="vless", max_length=32)
    tag: str = Field(max_length=128)
    listen: str = Field(default="::", max_length=128)
    port: int = Field(gt=0, le=65535)
    status: Literal["active", "disabled"] = "active"
    tls_enabled: bool = False
    server_name: str | None = Field(default=None, max_length=255)
    alpn: list[str] = Field(default_factory=list)
    reality_enabled: bool = False
    reality_handshake_server: str | None = Field(default=None, max_length=255)
    reality_handshake_port: int | None = Field(default=None, gt=0, le=65535)
    reality_private_key: str | None = None
    reality_public_key: str | None = None
    reality_short_id: str | None = Field(default=None, max_length=32)
    transport: dict[str, Any] = Field(default_factory=dict)
    multiplex: dict[str, Any] = Field(default_factory=dict)
    options: dict[str, Any] = Field(default_factory=dict)


class InboundCreate(InboundBase):
    pass


class InboundUpdate(BaseModel):
    protocol: str | None = Field(default=None, max_length=32)
    tag: str | None = Field(default=None, max_length=128)
    listen: str | None = Field(default=None, max_length=128)
    port: int | None = Field(default=None, gt=0, le=65535)
    status: Literal["active", "disabled"] | None = None
    tls_enabled: bool | None = None
    server_name: str | None = Field(default=None, max_length=255)
    alpn: list[str] | None = None
    reality_enabled: bool | None = None
    reality_handshake_server: str | None = Field(default=None, max_length=255)
    reality_handshake_port: int | None = Field(default=None, gt=0, le=65535)
    reality_private_key: str | None = None
    reality_public_key: str | None = None
    reality_short_id: str | None = Field(default=None, max_length=32)
    transport: dict[str, Any] | None = None
    multiplex: dict[str, Any] | None = None
    options: dict[str, Any] | None = None


class InboundOut(InboundBase):
    model_config = ConfigDict(from_attributes=True)

    id: int
    created_at: datetime
    updated_at: datetime


class UserBase(BaseModel):
    inbound_id: int
    username: str = Field(min_length=1, max_length=128)
    uuid: str | None = Field(default=None, max_length=64)
    password: str | None = Field(default=None, max_length=255)
    total_traffic: int = Field(default=0, ge=0)
    used_traffic: int = Field(default=0, ge=0)
    expire_time: datetime | None = None
    status: Literal["active", "disabled", "expired", "limited"] = "active"
    ip_limit: int = Field(default=0, ge=0)


class UserCreate(UserBase):
    pass


class UserUpdate(BaseModel):
    inbound_id: int | None = None
    username: str | None = Field(default=None, min_length=1, max_length=128)
    uuid: str | None = Field(default=None, max_length=64)
    password: str | None = Field(default=None, max_length=255)
    total_traffic: int | None = Field(default=None, ge=0)
    used_traffic: int | None = Field(default=None, ge=0)
    expire_time: datetime | None = None
    status: Literal["active", "disabled", "expired", "limited"] | None = None
    ip_limit: int | None = Field(default=None, ge=0)


class UserOut(UserBase):
    model_config = ConfigDict(from_attributes=True)

    id: int
    uuid: str
    created_at: datetime
    updated_at: datetime


class CoreStatus(BaseModel):
    mode: str
    running: bool
    detail: str = ""


class DashboardMetrics(BaseModel):
    cpu_percent: float
    memory_percent: float
    upload_bps: int
    download_bps: int
    active_users: int
    core: CoreStatus


class UserLinks(BaseModel):
    user_id: int
    links: list[str]
    subscription_url: str | None = None


class Message(BaseModel):
    message: str


class SettingOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)

    key: str
    value: str | None
    is_secret: bool


class SettingUpdate(BaseModel):
    value: str
