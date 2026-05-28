import uuid as uuid_lib
from datetime import UTC, datetime

from sqlalchemy import BigInteger, Boolean, DateTime, ForeignKey, Integer, String, Text, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship
from sqlalchemy.types import JSON

from app.db.base import Base


def utcnow() -> datetime:
    return datetime.now(UTC)


class TimestampMixin:
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True), default=utcnow, onupdate=utcnow
    )


class Admin(Base, TimestampMixin):
    __tablename__ = "admins"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    username: Mapped[str] = mapped_column(String(64), unique=True, index=True)
    password_hash: Mapped[str] = mapped_column(String(255))
    role: Mapped[str] = mapped_column(String(32), default="owner")
    is_active: Mapped[bool] = mapped_column(Boolean, default=True)
    last_login_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)


class Inbound(Base, TimestampMixin):
    __tablename__ = "inbounds"
    __table_args__ = (UniqueConstraint("tag", name="uq_inbounds_tag"),)

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    protocol: Mapped[str] = mapped_column(String(32), default="vless", index=True)
    tag: Mapped[str] = mapped_column(String(128), unique=True)
    listen: Mapped[str] = mapped_column(String(128), default="::")
    port: Mapped[int] = mapped_column(Integer, index=True)
    status: Mapped[str] = mapped_column(String(24), default="active", index=True)

    tls_enabled: Mapped[bool] = mapped_column(Boolean, default=False)
    server_name: Mapped[str | None] = mapped_column(String(255), nullable=True)
    alpn: Mapped[list[str]] = mapped_column(JSON, default=list)
    reality_enabled: Mapped[bool] = mapped_column(Boolean, default=False)
    reality_handshake_server: Mapped[str | None] = mapped_column(String(255), nullable=True)
    reality_handshake_port: Mapped[int | None] = mapped_column(Integer, nullable=True)
    reality_private_key: Mapped[str | None] = mapped_column(String(255), nullable=True)
    reality_public_key: Mapped[str | None] = mapped_column(String(255), nullable=True)
    reality_short_id: Mapped[str | None] = mapped_column(String(32), nullable=True)

    transport: Mapped[dict] = mapped_column(JSON, default=dict)
    multiplex: Mapped[dict] = mapped_column(JSON, default=dict)
    options: Mapped[dict] = mapped_column(JSON, default=dict)

    users: Mapped[list["User"]] = relationship(
        back_populates="inbound", cascade="all, delete-orphan", lazy="selectin"
    )


class User(Base, TimestampMixin):
    __tablename__ = "users"
    __table_args__ = (UniqueConstraint("inbound_id", "username", name="uq_users_inbound_username"),)

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    inbound_id: Mapped[int] = mapped_column(ForeignKey("inbounds.id", ondelete="CASCADE"), index=True)
    username: Mapped[str] = mapped_column(String(128), index=True)
    uuid: Mapped[str] = mapped_column(String(64), default=lambda: str(uuid_lib.uuid4()), index=True)
    password: Mapped[str | None] = mapped_column(String(255), nullable=True)
    total_traffic: Mapped[int] = mapped_column(BigInteger, default=0)
    used_traffic: Mapped[int] = mapped_column(BigInteger, default=0)
    expire_time: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    status: Mapped[str] = mapped_column(String(24), default="active", index=True)
    ip_limit: Mapped[int] = mapped_column(Integer, default=0)

    inbound: Mapped[Inbound] = relationship(back_populates="users")
    subscriptions: Mapped[list["Subscription"]] = relationship(
        back_populates="user", cascade="all, delete-orphan", lazy="selectin"
    )


class TrafficLedger(Base):
    __tablename__ = "traffic_ledger"
    __table_args__ = (
        UniqueConstraint("user_id", "window_start", name="uq_traffic_ledger_user_window"),
    )

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id", ondelete="CASCADE"), index=True)
    window_start: Mapped[datetime] = mapped_column(DateTime(timezone=True), index=True)
    upload_bytes: Mapped[int] = mapped_column(BigInteger, default=0)
    download_bytes: Mapped[int] = mapped_column(BigInteger, default=0)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)


class Setting(Base, TimestampMixin):
    __tablename__ = "settings"

    key: Mapped[str] = mapped_column(String(128), primary_key=True)
    value: Mapped[str] = mapped_column(Text)
    is_secret: Mapped[bool] = mapped_column(Boolean, default=False)


class ConfigRevision(Base):
    __tablename__ = "config_revisions"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    checksum: Mapped[str] = mapped_column(String(64), index=True)
    config_json: Mapped[dict] = mapped_column(JSON)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)
    applied_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    status: Mapped[str] = mapped_column(String(24), default="generated")
    error: Mapped[str | None] = mapped_column(Text, nullable=True)


class AuditLog(Base):
    __tablename__ = "audit_logs"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    actor_admin_id: Mapped[int | None] = mapped_column(ForeignKey("admins.id"), nullable=True)
    action: Mapped[str] = mapped_column(String(128), index=True)
    target_type: Mapped[str | None] = mapped_column(String(64), nullable=True)
    target_id: Mapped[str | None] = mapped_column(String(64), nullable=True)
    ip_address: Mapped[str | None] = mapped_column(String(64), nullable=True)
    metadata_json: Mapped[dict] = mapped_column(JSON, default=dict)
    created_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)


class ActiveIP(Base):
    __tablename__ = "active_ips"
    __table_args__ = (UniqueConstraint("user_id", "ip_address", name="uq_active_ips_user_ip"),)

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id", ondelete="CASCADE"), index=True)
    ip_address: Mapped[str] = mapped_column(String(64), index=True)
    first_seen_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)
    last_seen_at: Mapped[datetime] = mapped_column(DateTime(timezone=True), default=utcnow)


class Subscription(Base, TimestampMixin):
    __tablename__ = "subscriptions"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    user_id: Mapped[int] = mapped_column(ForeignKey("users.id", ondelete="CASCADE"), index=True)
    token: Mapped[str] = mapped_column(String(128), unique=True, index=True)
    status: Mapped[str] = mapped_column(String(24), default="active")
    last_used_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)

    user: Mapped[User] = relationship(back_populates="subscriptions")
