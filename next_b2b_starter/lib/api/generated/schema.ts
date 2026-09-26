// Generated from go-b2b-starter/apicontract/openapi.json; run pnpm api:generate.
export interface paths {
    "/auth/profile/me": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** getProfile */
        get: operations["getProfile"];
        /** updateProfile */
        put: operations["updateProfile"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/members": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** listMembers */
        get: operations["listMembers"];
        put?: never;
        /** inviteMember */
        post: operations["inviteMember"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/members/{member_id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        /** updateMemberRole */
        put: operations["updateMemberRole"];
        post?: never;
        /** removeMember */
        delete: operations["removeMember"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/auth/members/{member_id}/resend-invitation": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** resendInvitation */
        post: operations["resendInvitation"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/organizations": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** getOrganization */
        get: operations["getOrganization"];
        /** updateOrganization */
        put: operations["updateOrganization"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/subscriptions/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** getBillingStatus */
        get: operations["getBillingStatus"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/subscriptions/verify-payment": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** verifyPayment */
        post: operations["verifyPayment"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        /** @enum {string} */
        Role: "admin" | "manager" | "member";
        Success: {
            /** @constant */
            success: true;
        };
        Error: {
            error: string;
            message?: string;
            code?: string;
            success?: boolean;
        } | {
            error?: string;
            message: string;
            code?: string;
            success?: boolean;
        };
        NameInput: {
            name: string;
        };
        RoleInput: {
            role: components["schemas"]["Role"];
        };
        InviteInput: {
            /** Format: email */
            email: string;
            name: string;
            role_slug?: components["schemas"]["Role"];
        };
        VerifyPaymentInput: {
            /** Format: uuid */
            session_id: string;
        };
        Organization: {
            id: string;
            name: string;
            slug: string;
        };
        ProfileOrganization: {
            organization_id: string;
            name: string;
            slug: string;
            /** @constant */
            status: "active";
        };
        Profile: {
            member_id: string;
            email: string;
            name: string;
            roles: components["schemas"]["Role"][];
            permissions: ("org:view" | "org:manage")[];
            email_verified: boolean;
            /** @constant */
            status: "active";
            organization: components["schemas"]["ProfileOrganization"];
            account_id: number;
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            updated_at: string;
        };
        Member: {
            member_id: string;
            email: string;
            name: string;
            roles: components["schemas"]["Role"][];
            /** @enum {string} */
            status: "active" | "pending";
            email_verified: boolean;
            /** Format: date-time */
            created_at: string;
            /** Format: date-time */
            updated_at: string;
        };
        MemberList: {
            members: components["schemas"]["Member"][];
            total: number;
        };
        InvitationResult: {
            member_id: string;
            invite_sent: boolean;
        };
        DeliveryResult: {
            invite_sent: boolean;
        };
        ProfileResponse: {
            /** @constant */
            success: true;
            data: components["schemas"]["Profile"];
        };
        OrganizationResponse: {
            /** @constant */
            success: true;
            data: components["schemas"]["Organization"];
        };
        MemberListResponse: {
            /** @constant */
            success: true;
            data: components["schemas"]["MemberList"];
        };
        InvitationResponse: {
            /** @constant */
            success: true;
            data: components["schemas"]["InvitationResult"];
        };
        DeliveryResponse: {
            /** @constant */
            success: true;
            data: components["schemas"]["DeliveryResult"];
        };
        BillingStatus: {
            BillingEnabled: boolean;
            OrganizationID: number;
            ExternalID: string;
            HasActiveSubscription: boolean;
            SubscriptionID: string;
            SubscriptionStatus: string;
            ProductID: string;
            /** Format: date-time */
            CurrentPeriodEnd: string | null;
            CancelAtPeriodEnd: boolean;
            Reason: string;
            /** Format: date-time */
            CheckedAt: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    getProfile: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProfileResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    updateProfile: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["NameInput"];
            };
        };
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Success"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    listMembers: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["MemberListResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    inviteMember: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["InviteInput"];
            };
        };
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["InvitationResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    updateMemberRole: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Membership ID or invitation:<id>; scoped to the verified active organization. */
                member_id: string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RoleInput"];
            };
        };
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Success"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    removeMember: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Membership ID or invitation:<id>; scoped to the verified active organization. */
                member_id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Success"];
                };
            };
            /** @description Successful deletion without a response body (supported by the client). */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    resendInvitation: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                /** @description Membership ID or invitation:<id>; scoped to the verified active organization. */
                member_id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["DeliveryResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    getOrganization: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OrganizationResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    updateOrganization: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["NameInput"];
            };
        };
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["OrganizationResponse"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    getBillingStatus: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["BillingStatus"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
    verifyPayment: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["VerifyPaymentInput"];
            };
        };
        responses: {
            /** @description Successful operation */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["BillingStatus"];
                };
            };
            /** @description Permission denied. CORS middleware can reject before the JSON handler and return an empty body. */
            403: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
            /** @description Validation, authentication, authorization or dependency failure */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Error"];
                };
            };
        };
    };
}
