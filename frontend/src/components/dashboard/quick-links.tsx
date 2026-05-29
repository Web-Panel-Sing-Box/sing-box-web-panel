
import { Link } from "react-router-dom";
import { m } from "framer-motion";
import { ChevronRight, Network, ScrollText, Users } from "lucide-react";

const ITEMS = [
  {
    href: "/inbounds",
    title: "Manage inbounds",
    description: "Create, toggle, and clone protocol endpoints",
    Icon: Network
  },
  {
    href: "/clients",
    title: "Manage clients",
    description: "View user quotas, links, and connectivity",
    Icon: Users
  },
  {
    href: "/logs",
    title: "Open logs",
    description: "Stream sing-box runtime output in real time",
    Icon: ScrollText
  }
];

export function QuickLinks() {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
      {ITEMS.map(({ href, title, description, Icon }) => (
        <Link key={href} to={href} className="block">
          <m.div
            whileHover={{}}
            className="group flex h-full items-center gap-4 rounded-xl border border-subtle bg-surface p-5 shadow-card transition-colors duration-200 hover:bg-hover"
          >
            <div className="grid size-10 shrink-0 place-items-center rounded-lg border border-subtle bg-canvas text-ink-secondary transition-colors duration-200 group-hover:text-ink-primary">
              <Icon size={18} />
            </div>
            <m.div className="min-w-0 flex-1" whileHover={{ x: 4 }} transition={{ duration: 0.18 }}>
              <div className="truncate text-sm font-medium text-ink-primary">{title}</div>
              <div className="truncate text-xs text-ink-tertiary">{description}</div>
            </m.div>
            <ChevronRight size={16} className="shrink-0 text-ink-tertiary transition-colors duration-200 group-hover:text-ink-primary" />
          </m.div>
        </Link>
      ))}
    </div>
  );
}
