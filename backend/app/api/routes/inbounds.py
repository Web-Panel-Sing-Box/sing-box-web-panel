from fastapi import APIRouter, HTTPException, status
from sqlalchemy import select

from app.api.deps import CurrentAdmin, SessionDep, SettingsDep
from app.models import Inbound
from app.schemas import InboundCreate, InboundOut, InboundUpdate, Message
from app.services.audit import audit
from app.services.config_generator import LocalConfigGenerator

router = APIRouter(prefix="/inbounds", tags=["inbounds"])


@router.get("", response_model=list[InboundOut])
async def list_inbounds(session: SessionDep, _admin: CurrentAdmin) -> list[InboundOut]:
    result = await session.execute(select(Inbound).order_by(Inbound.id))
    return [InboundOut.model_validate(inbound) for inbound in result.scalars()]


@router.post("", response_model=InboundOut, status_code=status.HTTP_201_CREATED)
async def create_inbound(
    payload: InboundCreate,
    session: SessionDep,
    admin: CurrentAdmin,
) -> InboundOut:
    inbound = Inbound(**payload.model_dump())
    session.add(inbound)
    await session.flush()
    await audit(session, action="inbounds.create", actor=admin, target_type="inbound", target_id=str(inbound.id))
    await session.commit()
    await session.refresh(inbound)
    return InboundOut.model_validate(inbound)


@router.patch("/{inbound_id}", response_model=InboundOut)
async def update_inbound(
    inbound_id: int,
    payload: InboundUpdate,
    session: SessionDep,
    admin: CurrentAdmin,
) -> InboundOut:
    inbound = await session.get(Inbound, inbound_id)
    if inbound is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Inbound not found")
    for key, value in payload.model_dump(exclude_unset=True).items():
        setattr(inbound, key, value)
    await audit(
        session,
        action="inbounds.update",
        actor=admin,
        target_type="inbound",
        target_id=str(inbound.id),
    )
    await session.commit()
    await session.refresh(inbound)
    return InboundOut.model_validate(inbound)


@router.delete("/{inbound_id}", response_model=Message)
async def delete_inbound(inbound_id: int, session: SessionDep, admin: CurrentAdmin) -> Message:
    inbound = await session.get(Inbound, inbound_id)
    if inbound is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Inbound not found")
    await session.delete(inbound)
    await audit(session, action="inbounds.delete", actor=admin, target_type="inbound", target_id=str(inbound_id))
    await session.commit()
    return Message(message="Inbound deleted")


@router.post("/{inbound_id}/apply", response_model=Message)
async def apply_inbound(
    inbound_id: int,
    session: SessionDep,
    settings: SettingsDep,
    admin: CurrentAdmin,
) -> Message:
    if await session.get(Inbound, inbound_id) is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="Inbound not found")
    result = await LocalConfigGenerator(settings).generate(session, apply=True)
    await audit(
        session,
        action="inbounds.apply",
        actor=admin,
        target_type="inbound",
        target_id=str(inbound_id),
        metadata={"checksum": result.checksum, "revision_id": result.revision_id},
    )
    await session.commit()
    return Message(message=f"Config applied: {result.checksum}")
