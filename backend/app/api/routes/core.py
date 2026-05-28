from fastapi import APIRouter

from app.api.deps import CurrentAdmin, SessionDep, SettingsDep
from app.schemas import CoreStatus, Message
from app.services.audit import audit
from app.services.config_generator import LocalConfigGenerator
from app.services.process_manager import get_process_manager

router = APIRouter(prefix="/core", tags=["core"])


@router.get("/status", response_model=CoreStatus)
async def core_status(settings: SettingsDep, _admin: CurrentAdmin) -> CoreStatus:
    return await get_process_manager(settings).status()


@router.post("/start", response_model=CoreStatus)
async def start_core(settings: SettingsDep, session: SessionDep, admin: CurrentAdmin) -> CoreStatus:
    result = await get_process_manager(settings).start()
    await audit(session, action="core.start", actor=admin, metadata=result.model_dump())
    await session.commit()
    return result


@router.post("/stop", response_model=CoreStatus)
async def stop_core(settings: SettingsDep, session: SessionDep, admin: CurrentAdmin) -> CoreStatus:
    result = await get_process_manager(settings).stop()
    await audit(session, action="core.stop", actor=admin, metadata=result.model_dump())
    await session.commit()
    return result


@router.post("/restart", response_model=CoreStatus)
async def restart_core(settings: SettingsDep, session: SessionDep, admin: CurrentAdmin) -> CoreStatus:
    result = await get_process_manager(settings).restart()
    await audit(session, action="core.restart", actor=admin, metadata=result.model_dump())
    await session.commit()
    return result


@router.post("/reload", response_model=Message)
async def reload_core(settings: SettingsDep, session: SessionDep, admin: CurrentAdmin) -> Message:
    generated = await LocalConfigGenerator(settings).generate(session, apply=True)
    result = await get_process_manager(settings).reload()
    await audit(
        session,
        action="core.reload",
        actor=admin,
        metadata={"checksum": generated.checksum, "core": result.model_dump()},
    )
    await session.commit()
    return Message(message=f"Reloaded with config {generated.checksum}")
