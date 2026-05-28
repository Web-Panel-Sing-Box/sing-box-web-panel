import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}", "./lib/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        mono: ["var(--font-mono)", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"]
      },
      colors: {
        void: "#000000",
        panel: "#050505",
        panel2: "#0a0a0a",
        line: "#171717",
        glow: "#00e5ff",
        pulse: "#b7ff35"
      },
      boxShadow: {
        neon: "0 0 24px rgba(0,229,255,.18)",
        lime: "0 0 24px rgba(183,255,53,.16)"
      }
    }
  },
  plugins: [require("tailwindcss-animate")]
};

export default config;
