import { handleIdentityRequest } from "@/lib/auth/public-handler";
export function GET(request: Request) { return handleIdentityRequest(request); }
export function POST(request: Request) { return handleIdentityRequest(request); }
