/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  async rewrites() {
    const apiBase = process.env.SING_GROK_API_BASE_URL || "http://127.0.0.1:8081";
    return [{ source: "/api/:path*", destination: `${apiBase}/api/:path*` }];
  }
};

export default nextConfig;
