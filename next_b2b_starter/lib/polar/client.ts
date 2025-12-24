import { Polar } from "@polar-sh/sdk";

let cachedClient: Polar | null = null;

function createPolarClient(): Polar | null {
  if (typeof window !== "undefined") {
    throw new Error("Polar SDK client must only be instantiated on the server.");
  }

  const accessToken = process.env.POLAR_ACCESS_TOKEN;
  if (!accessToken) {
    return null;
  }

  const server = process.env.NODE_ENV === "production" ? "production" : "sandbox";

  return new Polar({
    accessToken,
    server,
  });
}

export function getPolarClient(): Polar | null {
  if (cachedClient) {
    return cachedClient;
  }

  const client = createPolarClient();
  if (!client) {
    return null;
  }

  cachedClient = client;
  return cachedClient;
}
