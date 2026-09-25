package organizations

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations/app/services"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"net/http/httptest"
	"strings"
	"testing"
)

type memberServiceSpy struct {
	services.MemberService
	calls int
	org   string
}

func (s *memberServiceSpy) AddMemberDirect(_ context.Context, req *services.AddMemberRequest) (*services.AddMemberResponse, error) {
	s.calls++
	s.org = req.OrgID
	return &services.AddMemberResponse{MemberID: "created"}, nil
}

type routeResolver struct{ permissions []auth.Permission }

func (r routeResolver) Get(string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := &auth.Identity{OrganizationID: "verified-organization", Permissions: r.permissions, Roles: []auth.Role{auth.RoleAdmin}}
		auth.SetIdentity(c, identity)
		auth.SetRequestContext(c, &auth.RequestContext{Identity: identity, ProviderOrgID: identity.OrganizationID, OrganizationID: 1})
		c.Next()
	}
}

func TestMemberCreationRequiresProviderPermissionAndUsesVerifiedTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, allowed := range []bool{false, true} {
		t.Run(map[bool]string{false: "denied", true: "allowed"}[allowed], func(t *testing.T) {
			service := &memberServiceSpy{}
			router := gin.New()
			resolver := routeResolver{}
			if allowed {
				resolver.permissions = []auth.Permission{auth.PermOrgManage}
			}
			NewRoutes(&OrganizationHandler{}, &AccountHandler{}, NewMemberHandler(service, logger.New())).RegisterRoutes(router.Group("/api"), resolver)
			req := httptest.NewRequest("POST", "/api/auth/members", strings.NewReader(`{"email":"new@example.com","name":"New member","role_slug":"member","org_id":"attacker-organization"}`))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if allowed {
				if rec.Code != 201 || service.calls != 1 || service.org != "verified-organization" {
					t.Fatalf("unexpected create: %d calls=%d org=%s", rec.Code, service.calls, service.org)
				}
			} else if rec.Code != 403 || service.calls != 0 {
				t.Fatalf("unprivileged create: %d calls=%d", rec.Code, service.calls)
			}
		})
	}
}
