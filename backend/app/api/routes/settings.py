from fastapi import APIRouter, HTTPException, status
from sqlalchemy import select

from app.api.deps import CurrentAdmin, SessionDep
from app.models import Setting
from app.schemas import SettingOut, SettingUpdate
from app.services.audit import audit

router = APIRouter(prefix="/settings", tags=["settings"])


@router.get("", response_model=list[SettingOut])
async def list_settings(session: SessionDep, _admin: CurrentAdmin) -> list[SettingOut]:
    result = await session.execute(select(Setting).order_by(Setting.key))
    items = []
    for setting in result.scalars():
        items.append(
            SettingOut(
                key=setting.key,
                value=None if setting.is_secret else setting.value,
                is_secret=setting.is_secret,
            )
        )
    return items


@router.patch("/{key}", response_model=SettingOut)
async def update_setting(
    key: str,
    payload: SettingUpdate,
    session: SessionDep,
    admin: CurrentAdmin,
) -> SettingOut:
    setting = await session.get(Setting, key)
    if setting is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Setting not found")
    setting.value = payload.value
    await audit(session, action="settings.update", actor=admin, target_type="setting", target_id=key)
    await session.commit()
    return SettingOut(key=setting.key, value=None if setting.is_secret else setting.value, is_secret=setting.is_secret)
