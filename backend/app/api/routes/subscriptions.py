from fastapi import APIRouter, HTTPException, Request, status
from fastapi.responses import PlainTextResponse
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from app.api.deps import SessionDep, SettingsDep
from app.models import Subscription
from app.services.link_builder import build_links_for_user

router = APIRouter(prefix="/subscriptions", tags=["subscriptions"])


@router.get("/{token}", response_class=PlainTextResponse)
async def read_subscription(
    token: str,
    request: Request,
    session: SessionDep,
    settings: SettingsDep,
) -> str:
    result = await session.execute(
        select(Subscription)
        .where(Subscription.token == token, Subscription.status == "active")
        .options(selectinload(Subscription.user))
    )
    subscription = result.scalar_one_or_none()
    if subscription is None or subscription.user.status != "active":
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Subscription not found")
    links, _ = await build_links_for_user(
        session,
        subscription.user,
        public_host=settings.public_host,
        panel_base_url=str(request.base_url),
    )
    return "\n".join(links) + "\n"
