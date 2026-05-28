import secrets
from io import BytesIO

import qrcode
from fastapi import APIRouter, HTTPException, Request, status
from fastapi.responses import StreamingResponse
from sqlalchemy import select
from sqlalchemy.orm import selectinload

from app.api.deps import CurrentAdmin, SessionDep, SettingsDep
from app.models import Inbound, Subscription, User
from app.schemas import Message, UserCreate, UserLinks, UserOut, UserUpdate
from app.services.audit import audit
from app.services.link_builder import build_links_for_user

router = APIRouter(prefix="/users", tags=["users"])


@router.get("", response_model=list[UserOut])
async def list_users(session: SessionDep, _admin: CurrentAdmin) -> list[UserOut]:
    result = await session.execute(select(User).order_by(User.id))
    return [UserOut.model_validate(user) for user in result.scalars()]


@router.post("", response_model=UserOut, status_code=status.HTTP_201_CREATED)
async def create_user(
    payload: UserCreate,
    request: Request,
    session: SessionDep,
    admin: CurrentAdmin,
) -> UserOut:
    inbound = await session.get(Inbound, payload.inbound_id)
    if inbound is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Inbound not found")
    user = User(**payload.model_dump(exclude_none=True))
    if payload.uuid:
        user.uuid = payload.uuid
    session.add(user)
    await session.flush()
    session.add(Subscription(user_id=user.id, token=secrets.token_urlsafe(32)))
    await audit(
        session,
        action="users.create",
        actor=admin,
        target_type="user",
        target_id=str(user.id),
        ip_address=request.client.host if request.client else None,
    )
    await session.commit()
    await session.refresh(user)
    return UserOut.model_validate(user)


@router.patch("/{user_id}", response_model=UserOut)
async def update_user(
    user_id: int,
    payload: UserUpdate,
    request: Request,
    session: SessionDep,
    admin: CurrentAdmin,
) -> UserOut:
    user = await session.get(User, user_id)
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    updates = payload.model_dump(exclude_unset=True)
    if "inbound_id" in updates and await session.get(Inbound, updates["inbound_id"]) is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Inbound not found")
    for key, value in updates.items():
        setattr(user, key, value)
    await audit(
        session,
        action="users.update",
        actor=admin,
        target_type="user",
        target_id=str(user.id),
        ip_address=request.client.host if request.client else None,
    )
    await session.commit()
    await session.refresh(user)
    return UserOut.model_validate(user)


@router.delete("/{user_id}", response_model=Message)
async def delete_user(user_id: int, session: SessionDep, admin: CurrentAdmin) -> Message:
    user = await session.get(User, user_id)
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    await session.delete(user)
    await audit(session, action="users.delete", actor=admin, target_type="user", target_id=str(user_id))
    await session.commit()
    return Message(message="User deleted")


@router.post("/{user_id}/reset-traffic", response_model=UserOut)
async def reset_traffic(user_id: int, session: SessionDep, admin: CurrentAdmin) -> UserOut:
    user = await session.get(User, user_id)
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    user.used_traffic = 0
    if user.status == "limited":
        user.status = "active"
    await audit(session, action="users.reset_traffic", actor=admin, target_type="user", target_id=str(user.id))
    await session.commit()
    await session.refresh(user)
    return UserOut.model_validate(user)


@router.post("/{user_id}/disable", response_model=UserOut)
async def disable_user(user_id: int, session: SessionDep, admin: CurrentAdmin) -> UserOut:
    user = await session.get(User, user_id)
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    user.status = "disabled"
    await audit(session, action="users.disable", actor=admin, target_type="user", target_id=str(user.id))
    await session.commit()
    await session.refresh(user)
    return UserOut.model_validate(user)


@router.get("/{user_id}/links", response_model=UserLinks)
async def user_links(
    user_id: int,
    request: Request,
    session: SessionDep,
    settings: SettingsDep,
    _admin: CurrentAdmin,
) -> UserLinks:
    result = await session.execute(
        select(User).where(User.id == user_id).options(selectinload(User.subscriptions))
    )
    user = result.scalar_one_or_none()
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    links, subscription_url = await build_links_for_user(
        session,
        user,
        public_host=settings.public_host,
        panel_base_url=str(request.base_url),
    )
    return UserLinks(user_id=user.id, links=links, subscription_url=subscription_url)


@router.get("/{user_id}/qr")
async def user_qr(
    user_id: int,
    request: Request,
    session: SessionDep,
    settings: SettingsDep,
    _admin: CurrentAdmin,
) -> StreamingResponse:
    result = await session.execute(select(User).where(User.id == user_id))
    user = result.scalar_one_or_none()
    if user is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="User not found")
    links, _subscription_url = await build_links_for_user(
        session,
        user,
        public_host=settings.public_host,
        panel_base_url=str(request.base_url),
    )
    if not links:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="No link available")
    image = qrcode.make(links[0])
    buffer = BytesIO()
    image.save(buffer, format="PNG")
    buffer.seek(0)
    return StreamingResponse(buffer, media_type="image/png")
