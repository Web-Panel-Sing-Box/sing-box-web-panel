from fastapi import APIRouter

from app.api.routes import auth, core, dashboard, inbounds, logs, settings, subscriptions, users

api_router = APIRouter(prefix="/api")
api_router.include_router(auth.router)
api_router.include_router(users.router)
api_router.include_router(inbounds.router)
api_router.include_router(core.router)
api_router.include_router(dashboard.router)
api_router.include_router(logs.router)
api_router.include_router(settings.router)
api_router.include_router(subscriptions.router)
