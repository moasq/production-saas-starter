import { SettingsContent } from "./components/settings-content";
import { isPolarEnabled } from "@/lib/polar/config";
export default function SettingsPage() { return <SettingsContent billingEnabled={isPolarEnabled()} />; }
