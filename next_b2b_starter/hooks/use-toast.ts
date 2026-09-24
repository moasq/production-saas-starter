import { toast as notify } from "sonner";
export function useToast() {
 return { toast: ({ title, description, variant }: { title?: string; description?: string; variant?: "default" | "destructive" }) => {
   if (variant === "destructive") notify.error(title || "Error", { description });
   else notify.success(title || "Done", { description });
 } };
}
