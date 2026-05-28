import asyncio
from pathlib import Path


async def tail_file(path: Path, lines: int = 300) -> list[str]:
    def _read() -> list[str]:
        if not path.exists():
            return []
        with path.open("rb") as handle:
            handle.seek(0, 2)
            end = handle.tell()
            block_size = 4096
            data = b""
            while end > 0 and data.count(b"\n") <= lines:
                read_size = min(block_size, end)
                end -= read_size
                handle.seek(end)
                data = handle.read(read_size) + data
            return data.decode(errors="replace").splitlines()[-lines:]

    return await asyncio.to_thread(_read)
