import asyncio
from collections import defaultdict
from dataclasses import dataclass
from datetime import UTC, datetime
from typing import Protocol

import httpx
from sqlalchemy import select, update
from sqlalchemy.dialects.sqlite import insert as sqlite_insert

from app.core.config import Settings
from app.db.session import SessionLocal
from app.models import TrafficLedger, User


@dataclass(slots=True)
class TrafficSnapshot:
    upload_bps: int = 0
    download_bps: int = 0


class PerUserTrafficSource(Protocol):
    async def collect(self) -> dict[str, tuple[int, int]]:
        """Return username -> (upload_delta, download_delta)."""


class ClashTrafficSource:
    def __init__(self, base_url: str, secret: str) -> None:
        self.base_url = base_url.rstrip("/")
        self.secret = secret

    async def read_once(self) -> TrafficSnapshot:
        headers = {"Authorization": f"Bearer {self.secret}"} if self.secret else {}
        async with httpx.AsyncClient(timeout=5) as client:
            response = await client.get(f"{self.base_url}/traffic", headers=headers)
            response.raise_for_status()
            payload = response.json()
        return TrafficSnapshot(
            upload_bps=int(payload.get("up", payload.get("upload", 0))),
            download_bps=int(payload.get("down", payload.get("download", 0))),
        )


class V2RayStatsSource:
    async def collect(self) -> dict[str, tuple[int, int]]:
        return {}


class ObservatorySource:
    def __init__(self, url: str) -> None:
        self.url = url

    async def collect(self) -> dict[str, tuple[int, int]]:
        async with httpx.AsyncClient(timeout=5) as client:
            response = await client.get(self.url)
        if response.status_code >= 400:
            return {}
        return {}


class TrafficBackgroundWorker:
    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self._task: asyncio.Task | None = None
        self._stop = asyncio.Event()
        self._deltas: defaultdict[str, list[int]] = defaultdict(lambda: [0, 0])
        self._pending_bytes = 0
        self.snapshot = TrafficSnapshot()
        self.clash = ClashTrafficSource(
            f"http://127.0.0.1:{settings.clash_api_port}",
            settings.clash_api_secret,
        )
        self.sources: list[PerUserTrafficSource] = [
            V2RayStatsSource(),
            ObservatorySource(settings.observatory_api_url),
        ]

    def start(self) -> None:
        if self._task is None:
            self._task = asyncio.create_task(self._run(), name="traffic-worker")

    async def stop(self) -> None:
        self._stop.set()
        if self._task:
            await self._task
        await self.flush()

    async def _run(self) -> None:
        last_flush = asyncio.get_running_loop().time()
        while not self._stop.is_set():
            await self._poll_once()
            now = asyncio.get_running_loop().time()
            if (
                now - last_flush >= self.settings.traffic_flush_seconds
                or self._pending_bytes >= self.settings.traffic_flush_bytes
            ):
                await self.flush()
                last_flush = now
            try:
                await asyncio.wait_for(self._stop.wait(), timeout=2)
            except TimeoutError:
                pass

    async def _poll_once(self) -> None:
        try:
            self.snapshot = await self.clash.read_once()
        except Exception:
            self.snapshot = TrafficSnapshot()

        for source in self.sources:
            try:
                deltas = await source.collect()
            except Exception:
                continue
            for username, (upload, download) in deltas.items():
                self._deltas[username][0] += max(0, upload)
                self._deltas[username][1] += max(0, download)
                self._pending_bytes += max(0, upload) + max(0, download)

    async def flush(self) -> None:
        if not self._deltas:
            return
        deltas = dict(self._deltas)
        self._deltas.clear()
        self._pending_bytes = 0
        window = datetime.now(UTC).replace(second=0, microsecond=0)
        async with SessionLocal() as session:
            result = await session.execute(select(User).where(User.username.in_(deltas.keys())))
            users_by_name = {user.username: user for user in result.scalars()}
            for username, (upload, download) in deltas.items():
                user = users_by_name.get(username)
                if user is None:
                    continue
                total_delta = upload + download
                stmt = (
                    sqlite_insert(TrafficLedger)
                    .values(
                        user_id=user.id,
                        window_start=window,
                        upload_bytes=upload,
                        download_bytes=download,
                    )
                    .on_conflict_do_update(
                        index_elements=["user_id", "window_start"],
                        set_={
                            "upload_bytes": TrafficLedger.upload_bytes + upload,
                            "download_bytes": TrafficLedger.download_bytes + download,
                        },
                    )
                )
                await session.execute(stmt)
                new_used = user.used_traffic + total_delta
                status = user.status
                if user.total_traffic and new_used >= user.total_traffic:
                    status = "limited"
                await session.execute(
                    update(User)
                    .where(User.id == user.id)
                    .values(used_traffic=new_used, status=status)
                )
            await session.commit()
