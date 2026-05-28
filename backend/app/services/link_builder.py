from base64 import urlsafe_b64encode
from urllib.parse import urlencode, quote

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.models import Inbound, Subscription, User


def _host(public_host: str, inbound: Inbound) -> str:
    if public_host:
        return public_host.removeprefix("https://").removeprefix("http://").split("/")[0]
    return inbound.server_name or "127.0.0.1"


def build_user_link(user: User, inbound: Inbound, public_host: str = "") -> str:
    host = _host(public_host, inbound)
    label = quote(user.username)
    if inbound.protocol == "vless":
        params = {
            "encryption": "none",
            "type": inbound.transport.get("type", "tcp") if inbound.transport else "tcp",
        }
        if inbound.reality_enabled:
            params.update(
                {
                    "security": "reality",
                    "sni": inbound.server_name or inbound.reality_handshake_server or "",
                    "pbk": inbound.reality_public_key or "",
                    "sid": inbound.reality_short_id or "",
                    "flow": "xtls-rprx-vision",
                }
            )
        elif inbound.tls_enabled:
            params.update({"security": "tls", "sni": inbound.server_name or ""})
        return f"vless://{user.uuid}@{host}:{inbound.port}?{urlencode(params)}#{label}"

    if inbound.protocol == "trojan":
        params = {"security": "tls" if inbound.tls_enabled else "none", "sni": inbound.server_name or ""}
        return f"trojan://{quote(user.password or user.uuid)}@{host}:{inbound.port}?{urlencode(params)}#{label}"

    if inbound.protocol == "shadowsocks":
        method = inbound.options.get("method", "2022-blake3-aes-128-gcm")
        secret = user.password or user.uuid
        encoded = urlsafe_b64encode(f"{method}:{secret}".encode()).decode().rstrip("=")
        return f"ss://{encoded}@{host}:{inbound.port}#{label}"

    return f"{inbound.protocol}://{quote(user.password or user.uuid)}@{host}:{inbound.port}#{label}"


async def build_links_for_user(
    session: AsyncSession,
    user: User,
    *,
    public_host: str,
    panel_base_url: str | None,
) -> tuple[list[str], str | None]:
    inbound = await session.get(Inbound, user.inbound_id)
    if inbound is None:
        return [], None
    result = await session.execute(
        select(Subscription).where(Subscription.user_id == user.id, Subscription.status == "active")
    )
    subscription = result.scalar_one_or_none()
    subscription_url = None
    if subscription and panel_base_url:
        subscription_url = f"{panel_base_url.rstrip('/')}/api/subscriptions/{subscription.token}"
    return [build_user_link(user, inbound, public_host)], subscription_url
