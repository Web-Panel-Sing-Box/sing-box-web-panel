from fastapi import APIRouter, Request
from sqlalchemy import func, select

from app.api.deps import CurrentAdmin, SessionDep, SettingsDep
from app.models import User
from app.schemas import CoreStatus, DashboardMetrics
from app.services.process_manager import get_process_manager

router = APIRouter(prefix="/dashboard", tags=["dashboard"])


@router.get("/metrics", response_model=DashboardMetrics)
async def dashboard_metrics(
    request: Request,
    session: SessionDep,
    settings: SettingsDep,
    _admin: CurrentAdmin,
) -> DashboardMetrics:
    try:
        import psutil

        cpu_percent = float(psutil.cpu_percent(interval=None))
        memory_percent = float(psutil.virtual_memory().percent)
    except Exception:
        cpu_percent = 0.0
        memory_percent = 0.0

    active_count = await session.scalar(select(func.count()).select_from(User).where(User.status == "active"))
    worker = getattr(request.app.state, "traffic_worker", None)
    snapshot = worker.snapshot if worker else None
    core = await get_process_manager(settings).status()
    return DashboardMetrics(
        cpu_percent=cpu_percent,
        memory_percent=memory_percent,
        upload_bps=int(getattr(snapshot, "upload_bps", 0)),
        download_bps=int(getattr(snapshot, "download_bps", 0)),
        active_users=int(active_count or 0),
        core=core if isinstance(core, CoreStatus) else CoreStatus(mode=settings.process_mode, running=False),
    )
