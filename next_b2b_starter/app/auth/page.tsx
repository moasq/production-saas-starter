import Content from "./content";
import { AuthSetup } from "@/components/auth/auth-setup";
import { isAuthConfigured } from "@/lib/auth/config";
export default function Page() {
  return isAuthConfigured() ? <Content /> : <AuthSetup />;
}
