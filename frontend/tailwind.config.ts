import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-sans)", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "SFMono-Regular", "Menlo", "monospace"]
      },
      colors: {
        canvas: "#171717",
        surface: "#212121",
        elevated: "#2f2f2f",
        hover: "#2a2a2a",
        brand: {
          DEFAULT: "#10a37f",
          soft: "#10a37f1a"
        },
        success: "#19c37d",
        danger: "#ef4444",
        cyan: "#22d3ee",
        violet: "#a78bfa",
        amber: "#facc15",
        ink: {
          primary: "#ffffff",
          secondary: "#b4b4b4",
          tertiary: "#8a8a8a"
        }
      },
      borderColor: {
        subtle: "rgba(255,255,255,0.08)",
        strong: "rgba(255,255,255,0.14)"
      },
      boxShadow: {
        card: "inset 0 1px 0 rgba(255,255,255,0.04), 0 8px 24px rgba(0,0,0,0.25)",
        pop: "0 12px 32px rgba(0,0,0,0.45)"
      },
      keyframes: {
        spin: {
          to: { transform: "rotate(360deg)" }
        }
      },
      animation: {
        spin: "spin 0.8s linear infinite"
      }
    }
  },
  plugins: [require("tailwindcss-animate")]
};

export default config;
