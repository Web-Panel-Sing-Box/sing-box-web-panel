import asyncio

from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from app.api.deps import CurrentAdmin, SettingsDep
from app.services.logs import tail_file

router = APIRouter(prefix="/logs", tags=["logs"])


@router.get("/sing-box", response_model=list[str])
async def read_sing_box_logs(settings: SettingsDep, _admin: CurrentAdmin, lines: int = 300) -> list[str]:
    return await tail_file(settings.sing_box_log_path, min(max(lines, 1), 1000))


@router.get("/stream")
async def stream_sing_box_logs(settings: SettingsDep, _admin: CurrentAdmin) -> StreamingResponse:
    async def events():
        offset = 0
        while True:
            path = settings.sing_box_log_path
            if path.exists():
                with path.open("r", encoding="utf-8", errors="replace") as handle:
                    handle.seek(offset)
                    for line in handle:
                        yield f"data: {line.rstrip()}\n\n"
                    offset = handle.tell()
            await asyncio.sleep(1)

    return StreamingResponse(events(), media_type="text/event-stream")
