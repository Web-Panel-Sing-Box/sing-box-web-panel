from sqlalchemy.ext.asyncio import AsyncSession

from app.models import Admin, AuditLog


async def audit(
    session: AsyncSession,
    *,
    action: str,
    actor: Admin | None = None,
    target_type: str | None = None,
    target_id: str | None = None,
    ip_address: str | None = None,
    metadata: dict | None = None,
) -> None:
    session.add(
        AuditLog(
            actor_admin_id=actor.id if actor else None,
            action=action,
            target_type=target_type,
            target_id=target_id,
            ip_address=ip_address,
            metadata_json=metadata or {},
        )
    )
