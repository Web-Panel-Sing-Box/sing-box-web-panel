
import { useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";
import {
  LayoutDashboard,
  Network,
  PanelLeftClose,
  PanelLeftOpen,
  ScrollText,
  Settings,
  Users,
  type LucideIcon
} from "lucide-react";

import { cn } from "@/lib/utils";
import { useI18n } from "@/lib/i18n";

type NavItem = {
  labelKey: "nav.dashboard" | "nav.inbounds" | "nav.clients" | "nav.settings" | "nav.logs";
  href: string;
  icon: LucideIcon;
};

const NAV: NavItem[] = [
  { labelKey: "nav.dashboard", href: "/", icon: LayoutDashboard },
  { labelKey: "nav.inbounds", href: "/inbounds", icon: Network },
  { labelKey: "nav.clients", href: "/clients", icon: Users },
  { labelKey: "nav.settings", href: "/settings", icon: Settings },
  { labelKey: "nav.logs", href: "/logs", icon: ScrollText }
];

const COLLAPSED = 64;
const EXPANDED = 240;
const STORAGE_KEY = "sidebar:pinned";

type SidebarProps = {
  mobileOpen: boolean;
  onCloseMobile: () => void;
};

export function Sidebar({ mobileOpen, onCloseMobile }: SidebarProps) {
  const pathname = useLocation().pathname;
  const [pinned, setPinned] = useState(true);
  const [hover, setHover] = useState(false);
  const grace = useRef<number | null>(null);

  useEffect(() => {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (stored !== null) setPinned(stored === "1");
  }, []);

  useEffect(() => {
    window.localStorage.setItem(STORAGE_KEY, pinned ? "1" : "0");
  }, [pinned]);

  useEffect(() => {
    onCloseMobile();
    // close drawer on route change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [pathname]);

  const isExpanded = pinned || hover;
  const width = isExpanded ? EXPANDED : COLLAPSED;

  const onEnter = () => {
    if (grace.current) window.clearTimeout(grace.current);
    setHover(true);
  };
  const onLeave = () => {
    if (grace.current) window.clearTimeout(grace.current);
    grace.current = window.setTimeout(() => setHover(false), 120);
  };

  return (
    <>
      {/* Hot-zone for hover reveal (desktop) */}
      {!pinned ? (
        <div
          className="fixed left-0 top-0 z-30 hidden h-screen w-3 lg:block"
          onMouseEnter={onEnter}
        />
      ) : null}

      {/* Desktop sidebar */}
      <motion.aside
        onMouseEnter={onEnter}
        onMouseLeave={onLeave}
        initial={false}
        animate={{ width }}
        transition={{ duration: 0.2, ease: "easeOut" }}
        className="sticky top-0 z-30 hidden h-screen shrink-0 flex-col border-r border-subtle bg-canvas lg:flex"
        style={{ width }}
      >
        <SidebarContents
          pinned={pinned}
          expanded={isExpanded}
          pathname={pathname}
          onPinToggle={() => setPinned((v) => !v)}
        />
      </motion.aside>

      {/* Mobile drawer */}
      <AnimatePresence>
        {mobileOpen ? (
          <>
            <motion.div
              key="backdrop"
              className="fixed inset-0 z-40 bg-black/60 lg:hidden"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.18 }}
              onClick={onCloseMobile}
            />
            <motion.aside
              key="drawer"
              className="fixed inset-y-0 left-0 z-50 flex w-[260px] flex-col border-r border-subtle bg-canvas lg:hidden"
              initial={{ x: -260 }}
              animate={{ x: 0 }}
              exit={{ x: -260 }}
              transition={{ duration: 0.22, ease: "easeOut" }}
            >
              <SidebarContents
                pinned
                expanded
                pathname={pathname}
                onPinToggle={onCloseMobile}
                pinIcon="close"
              />
            </motion.aside>
          </>
        ) : null}
      </AnimatePresence>
    </>
  );
}

function SidebarContents({
  pinned,
  expanded,
  pathname,
  onPinToggle,
  pinIcon = "pin"
}: {
  pinned: boolean;
  expanded: boolean;
  pathname: string;
  onPinToggle: () => void;
  pinIcon?: "pin" | "close";
}) {
  const { t } = useI18n();
  return (
    <>
      <div className="flex h-14 items-center gap-3 px-4">
        <div className="grid size-8 shrink-0 place-items-center rounded-md bg-white/5 text-ink-primary">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none">
            <path d="M5 8.5 12 5l7 3.5v7L12 19l-7-3.5v-7Z" stroke="currentColor" strokeWidth="1.5" />
            <path d="M5 8.5 12 12l7-3.5M12 12v7" stroke="currentColor" strokeWidth="1.5" />
          </svg>
        </div>
        <AnimatePresence>
          {expanded ? (
            <motion.span
              key="brand"
              initial={{ opacity: 0, x: -4 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -4 }}
              transition={{ duration: 0.12, ease: "easeOut" }}
              className="truncate text-sm font-semibold text-ink-primary"
            >
              Sing Grok
            </motion.span>
          ) : null}
        </AnimatePresence>
      </div>

      <nav className="flex-1 px-2 py-2">
        <ul className="flex flex-col gap-1">
          {NAV.map((item) => {
            const active = item.href === "/" ? pathname === "/" : pathname.startsWith(item.href);
            const Icon = item.icon;
            return (
              <li key={item.href}>
                <Link
                  to={item.href}
                  className={cn(
                    "relative flex h-10 items-center rounded-lg text-sm transition-colors duration-200",
                    expanded ? "gap-3 px-3" : "justify-center px-0",
                    active ? "text-ink-primary" : "text-ink-secondary hover:bg-hover hover:text-ink-primary"
                  )}
                >
                  {active ? (
                    <motion.span
                      layoutId="sidebar-active-pill"
                      className="absolute inset-0 -z-10 rounded-lg bg-surface"
                      transition={{ type: "spring", stiffness: 500, damping: 40 }}
                    />
                  ) : null}
                  <span className="grid size-6 shrink-0 place-items-center">
                    <Icon size={16} className="shrink-0" />
                  </span>
                  <AnimatePresence>
                    {expanded ? (
                      <motion.span
                        initial={{ opacity: 0, x: -4 }}
                        animate={{ opacity: 1, x: 0 }}
                        exit={{ opacity: 0, x: -4 }}
                        transition={{ duration: 0.12 }}
                        className="truncate"
                      >
                        {t(item.labelKey)}
                      </motion.span>
                    ) : null}
                  </AnimatePresence>
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="border-t border-subtle p-2">
        <button
          type="button"
          onClick={onPinToggle}
          className={cn(
            "flex h-10 w-full items-center rounded-lg text-sm text-ink-secondary transition-colors duration-200 hover:bg-hover hover:text-ink-primary",
            expanded ? "gap-3 px-3" : "justify-center px-0"
          )}
          title={pinIcon === "close" ? t("nav.close") : pinned ? t("nav.unpin") : t("nav.pin")}
        >
          <span className="grid size-6 shrink-0 place-items-center">
            {pinned ? <PanelLeftClose size={16} /> : <PanelLeftOpen size={16} />}
          </span>
          <AnimatePresence>
            {expanded ? (
              <motion.span
                initial={{ opacity: 0, x: -4 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -4 }}
                transition={{ duration: 0.12 }}
              >
                {pinIcon === "close" ? t("nav.close") : pinned ? t("nav.unpin") : t("nav.pin")}
              </motion.span>
            ) : null}
          </AnimatePresence>
        </button>
      </div>
    </>
  );
}
