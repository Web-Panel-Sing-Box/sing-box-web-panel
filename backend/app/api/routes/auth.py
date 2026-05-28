from datetime import UTC, datetime, timedelta

from fastapi import APIRouter, HTTPException, Request, Response, status
from sqlalchemy import select

from app.api.deps import CurrentAdmin, SessionDep, SettingsDep
from app.core.security import create_token, login_limiter, verify_password
from app.models import Admin
from app.schemas import AdminOut, LoginRequest, LoginResponse, Message
from app.services.audit import audit

router = APIRouter(prefix="/auth", tags=["auth"])


@router.post("/login", response_model=LoginResponse)
async def login(
    payload: LoginRequest,
    request: Request,
    response: Response,
    session: SessionDep,
    settings: SettingsDep,
) -> LoginResponse:
    client_ip = request.client.host if request.client else "unknown"
    limiter_key = f"{client_ip}:{payload.username}"
    if not login_limiter.check(limiter_key):
        raise HTTPException(status_code=status.HTTP_429_TOO_MANY_REQUESTS, detail="Too many attempts")

    result = await session.execute(select(Admin).where(Admin.username == payload.username))
    admin = result.scalar_one_or_none()
    if admin is None or not admin.is_active or not verify_password(payload.password, admin.password_hash):
        login_limiter.failure(limiter_key)
        await audit(session, action="auth.login_failed", ip_address=client_ip, metadata={"username": payload.username})
        await session.commit()
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid credentials")

    login_limiter.success(limiter_key)
    admin.last_login_at = datetime.now(UTC)
    access = create_token(
        subject=str(admin.id),
        token_type="access",
        settings=settings,
        expires_delta=timedelta(minutes=settings.access_token_minutes),
    )
    refresh = create_token(
        subject=str(admin.id),
        token_type="refresh",
        settings=settings,
        expires_delta=timedelta(days=settings.refresh_token_days),
    )
    response.set_cookie(
        "sg_access_token",
        access,
        httponly=True,
        secure=settings.cookie_secure,
        samesite="lax",
        max_age=settings.access_token_minutes * 60,
    )
    response.set_cookie(
        "sg_refresh_token",
        refresh,
        httponly=True,
        secure=settings.cookie_secure,
        samesite="lax",
        max_age=settings.refresh_token_days * 86400,
    )
    await audit(session, action="auth.login", actor=admin, ip_address=client_ip)
    await session.commit()
    return LoginResponse(admin=AdminOut.model_validate(admin))


@router.post("/logout", response_model=Message)
async def logout(response: Response, admin: CurrentAdmin, session: SessionDep) -> Message:
    response.delete_cookie("sg_access_token")
    response.delete_cookie("sg_refresh_token")
    await audit(session, action="auth.logout", actor=admin)
    await session.commit()
    return Message(message="Logged out")


@router.get("/me", response_model=AdminOut)
async def me(admin: CurrentAdmin) -> AdminOut:
    return AdminOut.model_validate(admin)
