# Sing Grok Web Panel

Status: Planning and early scaffold.

Sing Grok is a local-first web panel for managing a local `sing-box` process through `127.0.0.1`. It is designed to run on a user's own machine or VPS without a required external control plane.

## Stack

- Backend: Python 3.11+, FastAPI, asyncio, SQLite, SQLAlchemy Async.
- Frontend: Next.js App Router, TypeScript, Tailwind CSS, Framer Motion, Recharts.
- Core: local `sing-box` binary and local `config.json`.
- Ops: Bash installer, systemd units, local console menu.

## Safety Default

The panel must keep sing-box management APIs bound to `127.0.0.1`. Public exposure, if needed, belongs only to the web panel and must be configured explicitly by the operator.

## Local Development

### Frontend

The frontend is a Vite + React Single Page Application (SPA) located in the `frontend/` directory. 

To run the frontend locally:

1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the Vite development server:
   ```bash
   npm run dev
   ```

The application will be available at `http://127.0.0.1:3000/` and uses mock data for the UI by default when the backend is not connected.
