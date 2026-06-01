import { useEffect } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { AnimatePresence, m } from "framer-motion";
import {
  LayoutDashboard,
  LogOut,
  Network,
  ScrollText,
  Settings,
  Users,
  X,
  type LucideIcon
} from "lucide-react";

function GithubMark({ size = 16 }: { size?: number }) {
  return (
    <svg viewBox="0 0 24 24" width={size} height={size} fill="currentColor" aria-hidden="true">
      <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.08 3.29 9.39 7.86 10.91.58.1.79-.25.79-.56v-2c-3.2.7-3.88-1.36-3.88-1.36-.52-1.34-1.27-1.7-1.27-1.7-1.04-.71.08-.7.08-.7 1.15.08 1.76 1.18 1.76 1.18 1.02 1.76 2.68 1.25 3.34.96.1-.74.4-1.25.73-1.54-2.55-.29-5.24-1.28-5.24-5.69 0-1.26.45-2.29 1.18-3.1-.12-.29-.51-1.46.11-3.05 0 0 .97-.31 3.17 1.18a11.04 11.04 0 0 1 5.78 0c2.2-1.49 3.17-1.18 3.17-1.18.62 1.59.23 2.76.11 3.05.74.81 1.18 1.84 1.18 3.1 0 4.42-2.7 5.4-5.27 5.68.41.36.78 1.07.78 2.16v3.2c0 .31.21.67.8.56C20.21 21.39 23.5 17.08 23.5 12 23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

import { cn } from "@/lib/utils";
import { useAuth } from "@/lib/auth";
import { useI18n } from "@/lib/i18n";

type NavLabel = "nav.dashboard" | "nav.inbounds" | "nav.clients" | "nav.settings" | "nav.logs";

type NavItem = {
  labelKey: NavLabel;
  href: string;
  icon: LucideIcon;
};

const NAV: NavItem[] = [
  { labelKey: "nav.dashboard", href: "/dashboard", icon: LayoutDashboard },
  { labelKey: "nav.inbounds", href: "/inbounds", icon: Network },
  { labelKey: "nav.clients", href: "/clients", icon: Users },
  { labelKey: "nav.settings", href: "/settings", icon: Settings },
  { labelKey: "nav.logs", href: "/logs", icon: ScrollText }
];

type SidebarProps = {
  mobileOpen: boolean;
  onCloseMobile: () => void;
};

export function Sidebar({ mobileOpen, onCloseMobile }: SidebarProps) {
  const pathname = useLocation().pathname;

  useEffect(() => {
    onCloseMobile();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname]);

  return (
    <>
      {/* Desktop sidebar — fixed at 64px, icons only, no expand */}
      <aside className="sticky top-0 z-30 hidden h-screen w-16 shrink-0 flex-col border-r border-subtle bg-canvas lg:flex">
        <SidebarContents expanded={false} pathname={pathname} />
      </aside>

      {/* Mobile drawer */}
      <AnimatePresence>
        {mobileOpen ? (
          <>
            <m.div
              key="backdrop"
              className="fixed inset-0 z-40 bg-black/60 lg:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.18 }}
              onClick={onCloseMobile}
            />
            <m.aside
              key="drawer"
              className="fixed inset-y-0 left-0 z-50 flex w-[260px] flex-col border-r border-subtle bg-canvas lg:hidden"
              initial={{ x: -260 }}
              animate={{ x: 0 }}
              exit={{ x: -260 }}
              transition={{ duration: 0.22, ease: "easeOut" }}
            >
              <SidebarContents expanded pathname={pathname} onClose={onCloseMobile} />
            </m.aside>
          </>
        ) : null}
      </AnimatePresence>
    </>
  );
}

function SidebarContents({
  expanded,
  pathname,
  onClose
}: {
  expanded: boolean;
  pathname: string;
  onClose?: () => void;
}) {
  const { t } = useI18n();
  const { logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    onClose?.();
    logout();
    navigate("/login", { replace: true });
  };

  return (
    <>
      <div className={cn("flex h-14 items-center", expanded ? "gap-3 px-4" : "justify-center")}>
        <div className="grid size-8 shrink-0 place-items-center rounded-md bg-white/5 text-ink-primary">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <path d="M5 8.5 12 5l7 3.5v7L12 19l-7-3.5v-7Z" stroke="currentColor" strokeWidth="1.5" />
            <path d="M5 8.5 12 12l7-3.5M12 12v7" stroke="currentColor" strokeWidth="1.5" />
          </svg>
        </div>
        {expanded ? (
          <span className="truncate text-sm font-semibold text-ink-primary">Sing box</span>
        ) : null}
        {expanded && onClose ? (
          <button
            type="button"
            onClick={onClose}
            className="ml-auto grid size-8 place-items-center rounded-md text-ink-secondary transition-colors duration-150 hover:bg-hover hover:text-ink-primary"
            aria-label={t("nav.close")}
          >
            <X size={16} />
          </button>
        ) : null}
      </div>

      <nav className="flex-1 px-2 py-2">
        <ul className="flex flex-col gap-1">
          {NAV.map((item) => {
            const active = pathname === item.href || (item.href !== "/dashboard" && pathname.startsWith(item.href));
            const Icon = item.icon;
            const label = t(item.labelKey);
            return (
              <li key={item.href}>
                <Link
                  to={item.href}
                  onClick={onClose}
                  title={label}
                  aria-label={label}
                  className={cn(
                    "relative flex h-10 items-center rounded-lg text-sm transition-colors duration-200",
                    expanded ? "gap-3 px-3" : "justify-center px-0",
                    active ? "text-ink-primary" : "text-ink-secondary hover:bg-hover hover:text-ink-primary"
                  )}
                >
                  {active ? (
                    <m.span
                      layoutId="sidebar-active-pill"
                      className="absolute inset-0 -z-10 rounded-lg bg-surface"
                      transition={{ type: "spring", stiffness: 500, damping: 40 }}
                    />
                  ) : null}
                  <span className="grid size-6 shrink-0 place-items-center">
                    <Icon size={16} className="shrink-0" />
                  </span>
                  {expanded ? <span className="truncate">{label}</span> : null}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="flex flex-col gap-1 px-2 py-3">
        <a
          href="https://github.com/Web-Panel-Sing-Box/sing-box-web-panel"
          target="_blank"
          rel="noopener noreferrer"
          title="GitHub"
          aria-label="GitHub repository"
          onClick={onClose}
          className={cn(
            "flex h-10 items-center rounded-lg text-sm text-ink-secondary transition-colors duration-200 hover:bg-hover hover:text-ink-primary",
            expanded ? "gap-3 px-3" : "justify-center px-0"
          )}
        >
          <span className="grid size-6 shrink-0 place-items-center">
            <GithubMark size={16} />
          </span>
          {expanded ? <span className="truncate">GitHub</span> : null}
        </a>
        <button
          type="button"
          onClick={handleLogout}
          title={t("nav.logout")}
          aria-label={t("nav.logout")}
          className={cn(
            "flex h-10 items-center rounded-lg text-sm text-ink-secondary transition-colors duration-200 hover:bg-hover hover:text-danger",
            expanded ? "gap-3 px-3" : "justify-center px-0"
          )}
        >
          <span className="grid size-6 shrink-0 place-items-center">
            <LogOut size={16} />
          </span>
          {expanded ? <span className="truncate">{t("nav.logout")}</span> : null}
        </button>
      </div>
    </>
  );
}
