// Included by tsc, never executed. These errors must remain errors after generation.
import type { ApiClient } from "../lib/api/api/client/api-client.ts";
import type { components } from "../lib/api/generated/schema.ts";
export function checkContractTypes(client: ApiClient) {
  // @ts-expect-error Unmounted path.
  client.GET("/auth/members/missing-route");
  // @ts-expect-error Missing required profile body.
  client.PUT("/auth/profile/me");
  // @ts-expect-error Privileged role is not part of the API.
  client.PUT("/auth/members/{member_id}", { params: { path: { member_id: "one" } }, body: { role: "owner" } });
  // @ts-expect-error Tenant selection is not an invitation input.
  const invalid: components["schemas"]["InviteInput"] = { email: "one@example.test", name: "One", organization_id: "other" };
  void invalid;
}
