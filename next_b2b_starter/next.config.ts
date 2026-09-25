import type { NextConfig } from "next";
const development = process.env.NODE_ENV === "development";
const policy = [
 "default-src 'self'", `script-src 'self' 'unsafe-inline'${development ? " 'unsafe-eval'" : ""}`,
 "style-src 'self' 'unsafe-inline'", "img-src 'self' data:", "font-src 'self'", "connect-src 'self'", "object-src 'none'", "base-uri 'self'", "frame-ancestors 'none'", "form-action 'self'",
].join("; ");
const config: NextConfig = {
 output: "standalone", poweredByHeader: false,
 async headers() { return [{ source: "/:path*", headers: [
   { key: "X-Content-Type-Options", value: "nosniff" },
   { key: "X-Frame-Options", value: "DENY" },
   { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
   { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
   { key: "Content-Security-Policy", value: policy },
 ] }]; },
 async rewrites() { return development ? [{ source: "/api/:path*", destination: `${process.env.API_BASE_URL_INTERNAL || "http://localhost:8080/api"}/:path*` }] : []; },
};
export default config;
