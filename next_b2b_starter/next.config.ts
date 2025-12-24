import type { NextConfig } from "next";

const withDefault = (value: string | undefined, fallback: string) => {
  const trimmed = value?.trim();
  return trimmed && trimmed.length > 0 ? trimmed : fallback;
};

const isDevelopment = process.env.NODE_ENV === 'development';

const ContentSecurityPolicy = [
  "default-src 'self';",
  // Note: Add 'unsafe-eval' back if Google Tag Manager is needed in production
  "script-src 'self' 'unsafe-inline' https://www.youtube.com https://youtube.com https://www.youtube-nocookie.com https://youtube-nocookie.com https://www.googletagmanager.com;",
  "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;",
  "img-src 'self' data: https://www.youtube.com https://youtube.com https://img.youtube.com https://i.ytimg.com https://fonts.gstatic.com;",
  "font-src 'self' https://fonts.gstatic.com;",
  isDevelopment
    ? "connect-src 'self' http://localhost:8080 https://test.stytch.com https://api.stytch.com https://www.youtube.com https://youtube.com https://www.youtube-nocookie.com;"
    : "connect-src 'self' https://api.stytch.com https://www.youtube.com https://youtube.com https://www.youtube-nocookie.com;",
  "frame-src https://www.youtube.com https://youtube.com https://www.youtube-nocookie.com https://youtube-nocookie.com;",
  "media-src 'self' https://www.youtube.com https://youtube.com https://www.youtube-nocookie.com;",
  "object-src 'none';",
  "base-uri 'self';",
  "frame-ancestors 'self';",
].join(" ");

const nextConfig: NextConfig = {
  // Enable standalone output for Docker optimization
  // This reduces the Docker image size by ~75% by including only required files
  output: "standalone",

  compress: true,
  images: {
    formats: ["image/avif", "image/webp"],
  },
  experimental: {
    optimizePackageImports: ["lucide-react"],
  },

  env: {
    NEXT_PUBLIC_STYTCH_LOGIN_PATH: withDefault(
      process.env.NEXT_PUBLIC_STYTCH_LOGIN_PATH,
      "/auth"
    ),
    NEXT_PUBLIC_STYTCH_REDIRECT_PATH: withDefault(
      process.env.NEXT_PUBLIC_STYTCH_REDIRECT_PATH,
      "/authenticate"
    ),
    NEXT_PUBLIC_STYTCH_SESSION_DURATION_MINUTES: withDefault(
      process.env.NEXT_PUBLIC_STYTCH_SESSION_DURATION_MINUTES,
      "2880"
    ),
    NEXT_PUBLIC_STYTCH_PROJECT_ENV: withDefault(
      process.env.NEXT_PUBLIC_STYTCH_PROJECT_ENV,
      process.env.STYTCH_PROJECT_ENV || "test"
    ),
    NEXT_PUBLIC_APP_BASE_URL: withDefault(
      process.env.NEXT_PUBLIC_APP_BASE_URL,
      process.env.APP_BASE_URL || "http://localhost:3000"
    ),
  },

  async headers() {
    return [
      // Security Headers
      {
        source: "/:path*",
        headers: [
          {
            key: "X-DNS-Prefetch-Control",
            value: "on",
          },
          {
            key: "Strict-Transport-Security",
            value: "max-age=63072000; includeSubDomains; preload",
          },
          {
            key: "X-Content-Type-Options",
            value: "nosniff",
          },
          {
            key: "X-Frame-Options",
            value: "SAMEORIGIN",
          },
          {
            key: "X-XSS-Protection",
            value: "1; mode=block",
          },
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=(), interest-cohort=()",
          },
          {
            key: "Content-Security-Policy",
            value: ContentSecurityPolicy.replace(/\s{2,}/g, " ").trim(),
          },
          {
            key: "Cross-Origin-Resource-Policy",
            value: "same-site",
          },
          {
            key: "Cross-Origin-Opener-Policy",
            value: "same-origin-allow-popups",
          },
        ],
      },
      // SEO Image Caching
      {
        source: "/opengraph-image.png",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      {
        source: "/twitter-image.png",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      {
        source: "/icon.png",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      {
        source: "/favicon.ico",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      {
        source: "/apple-touch-icon.png",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      {
        source: "/screenshot.webp",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=0, s-maxage=31536000, immutable",
          },
        ],
      },
      // Static Assets Caching
      {
        source: "/fonts/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
      {
        source: "/_next/static/:path*",
        headers: [
          {
            key: "Cache-Control",
            value: "public, max-age=31536000, immutable",
          },
        ],
      },
    ];
  },

  async redirects() {
    return [
      // Redirects have been removed for the starter kit
    ];
  },
};

export default nextConfig;
