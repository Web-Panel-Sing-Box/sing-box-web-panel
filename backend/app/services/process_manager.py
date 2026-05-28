import asyncio
import os
import signal
from abc import ABC, abstractmethod
from pathlib import Path

from app.core.config import Settings
from app.schemas import CoreStatus


class ProcessManager(ABC):
    def __init__(self, settings: Settings) -> None:
        self.settings = settings

    @abstractmethod
    async def start(self) -> CoreStatus:
        raise NotImplementedError

    @abstractmethod
    async def stop(self) -> CoreStatus:
        raise NotImplementedError

    async def restart(self) -> CoreStatus:
        await self.stop()
        return await self.start()

    async def reload(self) -> CoreStatus:
        return await self.restart()

    @abstractmethod
    async def status(self) -> CoreStatus:
        raise NotImplementedError


class SystemdProcessManager(ProcessManager):
    async def _systemctl(self, command: str) -> tuple[int, str]:
        proc = await asyncio.create_subprocess_exec(
            "systemctl",
            command,
            self.settings.sing_box_service_name,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
        )
        stdout, stderr = await proc.communicate()
        return proc.returncode, (stdout or stderr).decode(errors="replace").strip()

    async def start(self) -> CoreStatus:
        code, output = await self._systemctl("start")
        return CoreStatus(mode="systemd", running=code == 0, detail=output)

    async def stop(self) -> CoreStatus:
        code, output = await self._systemctl("stop")
        return CoreStatus(mode="systemd", running=False if code == 0 else True, detail=output)

    async def restart(self) -> CoreStatus:
        code, output = await self._systemctl("restart")
        return CoreStatus(mode="systemd", running=code == 0, detail=output)

    async def status(self) -> CoreStatus:
        code, output = await self._systemctl("is-active")
        return CoreStatus(mode="systemd", running=code == 0 and output == "active", detail=output)


class SubprocessProcessManager(ProcessManager):
    @property
    def pid_path(self) -> Path:
        return self.settings.sing_box_pid_path

    async def start(self) -> CoreStatus:
        existing = await self.status()
        if existing.running:
            return existing
        self.pid_path.parent.mkdir(parents=True, exist_ok=True)
        log_path = self.settings.sing_box_log_path
        log_path.parent.mkdir(parents=True, exist_ok=True)
        log_file = open(log_path, "ab", buffering=0)  # noqa: SIM115
        proc = await asyncio.create_subprocess_exec(
            self.settings.sing_box_binary,
            "run",
            "-c",
            str(self.settings.sing_box_config_path),
            stdout=log_file,
            stderr=log_file,
            start_new_session=True,
        )
        self.pid_path.write_text(str(proc.pid), encoding="utf-8")
        os.chmod(self.pid_path, 0o640)
        return CoreStatus(mode="subprocess", running=True, detail=f"pid={proc.pid}")

    async def stop(self) -> CoreStatus:
        pid = self._read_pid()
        if pid is None:
            return CoreStatus(mode="subprocess", running=False, detail="not running")
        try:
            os.kill(pid, signal.SIGTERM)
        except ProcessLookupError:
            self.pid_path.unlink(missing_ok=True)
            return CoreStatus(mode="subprocess", running=False, detail="stale pid removed")
        for _ in range(30):
            await asyncio.sleep(0.2)
            if not self._pid_alive(pid):
                self.pid_path.unlink(missing_ok=True)
                return CoreStatus(mode="subprocess", running=False, detail="stopped")
        os.kill(pid, signal.SIGKILL)
        self.pid_path.unlink(missing_ok=True)
        return CoreStatus(mode="subprocess", running=False, detail="killed after timeout")

    async def status(self) -> CoreStatus:
        pid = self._read_pid()
        running = pid is not None and self._pid_alive(pid)
        return CoreStatus(
            mode="subprocess",
            running=running,
            detail=f"pid={pid}" if running else "not running",
        )

    def _read_pid(self) -> int | None:
        try:
            return int(self.pid_path.read_text(encoding="utf-8").strip())
        except (FileNotFoundError, ValueError):
            return None

    def _pid_alive(self, pid: int) -> bool:
        try:
            os.kill(pid, 0)
        except ProcessLookupError:
            return False
        return True


def get_process_manager(settings: Settings) -> ProcessManager:
    if settings.process_mode == "subprocess":
        return SubprocessProcessManager(settings)
    return SystemdProcessManager(settings)
