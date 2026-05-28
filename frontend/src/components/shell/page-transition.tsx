
import { useLocation } from "react-router-dom";
import { AnimatePresence, motion } from "framer-motion";

import { pageVariants } from "@/lib/motion";

export function PageTransition({ children }: { children: React.ReactNode }) {
  const pathname = useLocation().pathname;
  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.div
        key={pathname}
        variants={pageVariants}
        initial="initial"
        animate="animate"
        exit="exit"
        className="min-h-[calc(100vh-56px)]"
      >
        {children}
      </motion.div>
    </AnimatePresence>
  );
}
