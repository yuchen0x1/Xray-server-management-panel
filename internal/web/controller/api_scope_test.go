package controller

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
)

// Every route the master panel calls on a node, transcribed from
// internal/web/runtime/remote.go and the heartbeat probe in
// internal/web/service/node.go. If a node sync RPC is added there without a
// matching nodeSyncScopeAllow entry, the master starts receiving 403s that
// surface only as failed cluster syncs — this table is what makes that a build
// failure instead.
var nodeSyncRequiredRoutes = []struct {
	method string
	path   string
}{
	{http.MethodGet, "/panel/api/server/status"},
	{http.MethodGet, "/panel/api/inbounds/list"},
	{http.MethodPost, "/panel/api/inbounds/add"},
	{http.MethodPost, "/panel/api/inbounds/del/:id"},
	{http.MethodPost, "/panel/api/inbounds/update/:id"},
	{http.MethodPost, "/panel/api/inbounds/:id/resetTraffic"},
	{http.MethodPost, "/panel/api/inbounds/resetAllTraffics"},
	{http.MethodPost, "/panel/api/inbounds/pushClientTraffics"},
	{http.MethodPost, "/panel/api/clients/add"},
	{http.MethodPost, "/panel/api/clients/del/:email"},
	{http.MethodPost, "/panel/api/clients/:email/detach"},
	{http.MethodPost, "/panel/api/clients/update/:email"},
	{http.MethodPost, "/panel/api/clients/resetTraffic/:email"},
	{http.MethodPost, "/panel/api/clients/onlines"},
	{http.MethodPost, "/panel/api/clients/onlinesByGuid"},
	{http.MethodPost, "/panel/api/clients/lastOnline"},
	{http.MethodPost, "/panel/api/clients/clientIpsByGuid"},
	{http.MethodGet, "/panel/api/server/clientIps"},
	{http.MethodPost, "/panel/api/server/clientIps"},
	{http.MethodPost, "/panel/api/server/restartXrayService"},
	{http.MethodGet, "/panel/api/server/getWebCertFiles"},
	{http.MethodGet, "/panel/api/server/descendants"},
	{http.MethodGet, "/panel/api/hosts/list"},
}

func TestNodeSyncScopeAllowsEveryClusterRPC(t *testing.T) {
	for _, route := range nodeSyncRequiredRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			if !scopeAllows(nodeSyncScopeAllow, relAPIPath(route.path), route.method) {
				t.Fatalf("node-sync token would be refused on %s %s — cluster sync would break with a silent 403", route.method, route.path)
			}
		})
	}
}

// Node-admin is a superset, so the same table must pass for it.
func TestNodeAdminScopeAllowsEveryClusterRPC(t *testing.T) {
	for _, route := range nodeSyncRequiredRoutes {
		rel := relAPIPath(route.path)
		allowed := scopeAllows(nodeSyncScopeAllow, rel, route.method) || scopeAllows(nodeAdminScopeAllow, rel, route.method)
		if !allowed {
			t.Fatalf("node-admin token would be refused on %s %s", route.method, route.path)
		}
	}
}

// A node-sync token is cluster plumbing, not an operator credential: it must
// not reach settings, xray config, other tokens, or panel updates.
func TestNodeSyncScopeDeniesOperatorSurfaces(t *testing.T) {
	denied := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/panel/api/server/updatePanel"},
		{http.MethodGet, "/panel/api/setting/all"},
		{http.MethodPost, "/panel/api/setting/apiTokens/create"},
		{http.MethodPost, "/panel/api/setting/update"},
		{http.MethodPost, "/panel/api/xray/update"},
		{http.MethodPost, "/panel/api/nodes/add"},
		{http.MethodGet, "/panel/api/server/getDb"},
		{http.MethodPost, "/panel/api/server/importDB"},
	}
	for _, route := range denied {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			if scopeAllows(nodeSyncScopeAllow, relAPIPath(route.path), route.method) {
				t.Fatalf("node-sync scope is allowed to reach %s %s", route.method, route.path)
			}
		})
	}
}

// The only thing node-admin adds over node-sync is the panel software update,
// which rewrites the node's own binary and cannot be undone remotely.
func TestNodeAdminScopeAddsOnlyPanelUpdate(t *testing.T) {
	if !scopeAllows(nodeAdminScopeAllow, relAPIPath("/panel/api/server/updatePanel"), http.MethodPost) {
		t.Fatal("node-admin must be able to trigger a panel update")
	}
	if scopeAllows(nodeSyncScopeAllow, relAPIPath("/panel/api/server/updatePanel"), http.MethodPost) {
		t.Fatal("node-sync must not be able to trigger a panel update")
	}
	if len(nodeAdminScopeAllow) != 1 {
		t.Fatalf("nodeAdminScopeAllow has %d entries, want exactly the updatePanel route; widening it silently grants cluster tokens more power", len(nodeAdminScopeAllow))
	}
}

func TestMonitorScopeIsReadOnly(t *testing.T) {
	// FullPath() hands the middleware Gin route patterns, so the allowlist keys
	// are patterns too — a concrete path would never match in production.
	reads := []string{
		"/panel/api/server/status",
		"/panel/api/server/history/:metric/:bucket",
		"/panel/api/server/xrayObservatory",
		"/panel/api/nodes/history/:id/:metric/:bucket",
	}
	for _, path := range reads {
		if !scopeAllows(monitorScopeAllow, relAPIPath(path), http.MethodGet) {
			t.Fatalf("monitor scope should allow GET %s", path)
		}
	}

	mutations := []string{
		"/panel/api/server/restartXrayService",
		"/panel/api/inbounds/add",
		"/panel/api/clients/add",
		"/panel/api/setting/update",
	}
	for _, path := range mutations {
		if scopeAllows(monitorScopeAllow, relAPIPath(path), http.MethodPost) {
			t.Fatalf("monitor scope must not allow POST %s", path)
		}
	}

	// Even a read-only endpoint is refused when reached with a mutating verb,
	// so the allowlist cannot be bypassed by method confusion.
	if scopeAllows(monitorScopeAllow, "/server/status", http.MethodPost) {
		t.Fatal("monitor scope allowed POST on a GET-only route")
	}
}

func TestScopeAllowsRejectsUnlistedPathExactly(t *testing.T) {
	if scopeAllows(monitorScopeAllow, "/server/status/extra", http.MethodGet) {
		t.Fatal("allowlist matched a longer path than it declares")
	}
	if scopeAllows(monitorScopeAllow, "", http.MethodGet) {
		t.Fatal("allowlist matched an empty path (relAPIPath found no /panel/api marker)")
	}
}

// newScopeTestEngine mirrors production wiring: checkAPIAuth populates the
// scope, then enforceTokenScope gates the route. It returns the issuer of a
// monitor-scoped token plus that token's plaintext.
func newScopeTestEngine(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	if err := database.InitDB(filepath.Join(dbDir, "x-ui.db")); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })

	const plaintext = "scope-test-monitor-token"
	if err := database.GetDB().Create(&model.ApiToken{
		Name:    "scope-test",
		Token:   crypto.HashTokenSHA256(plaintext),
		Enabled: true,
		Scope:   model.ApiScopeMonitor,
	}).Error; err != nil {
		t.Fatalf("seed token: %v", err)
	}

	engine := gin.New()
	engine.Use(sessions.Sessions("3x-ui", cookie.NewStore([]byte("scope-test-secret"))))
	a := &APIController{}

	api := engine.Group("/panel/api")
	api.Use(a.checkAPIAuth)
	api.Use(a.enforceTokenScope)
	api.GET("/server/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"allowed": true})
	})
	api.POST("/inbounds/add", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"allowed": true})
	})
	return engine, plaintext
}

func TestMonitorTokenReachesReadRouteAndIsDeniedOnWrite(t *testing.T) {
	engine, plaintext := newScopeTestEngine(t)

	bearer := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+plaintext)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		return w
	}

	if got := bearer(http.MethodGet, "/panel/api/server/status").Code; got != http.StatusOK {
		t.Fatalf("monitor GET status = %d, want 200", got)
	}
	if got := bearer(http.MethodPost, "/panel/api/inbounds/add").Code; got != http.StatusForbidden {
		t.Fatalf("monitor POST inbound = %d, want 403", got)
	}
}
