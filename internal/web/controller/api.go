package controller

import (
	"net/http"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/middleware"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/tgbot"
	"github.com/mhsanaei/3x-ui/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

// APIController handles the main API routes for the 3x-ui panel, including inbounds and server management.
type APIController struct {
	BaseController
	inboundController     *InboundController
	serverController      *ServerController
	nodeController        *NodeController
	hostController        *HostController
	settingController     *SettingController
	xraySettingController *XraySettingController
	userService           panel.UserService
	apiTokenService       panel.ApiTokenService
	Tgbot                 tgbot.Tgbot
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

func (a *APIController) checkAPIAuth(c *gin.Context) {
	setAuthenticated := func(scope string) {
		if u, err := a.userService.GetFirstUser(); err == nil {
			session.SetAPIAuthUser(c, u)
		}
		c.Set("api_authed", true)
		c.Set("api_token_scope", scope)
	}
	matchedScope := ""
	if after, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); ok {
		if row, matched := a.apiTokenService.MatchToken(after); matched {
			matchedScope = row.Scope
		}
	}
	// A verified client certificate (a completed mTLS handshake) authenticates
	// the caller, equivalent to a valid bearer token. api_authed must be set so
	// the CSRF middleware lets cert-authed mutations through. Verified mTLS keeps
	// its historical node-sync identity; only an explicit node-admin bearer
	// elevates it to software-update authority.
	if c.Request.TLS != nil && len(c.Request.TLS.VerifiedChains) > 0 {
		scope := model.ApiScopeNodeSync
		if matchedScope == model.ApiScopeNodeAdmin {
			scope = model.ApiScopeNodeAdmin
		}
		setAuthenticated(scope)
		c.Next()
		return
	}
	if matchedScope != "" {
		setAuthenticated(matchedScope)
		c.Next()
		return
	}
	if !session.IsLogin(c) {
		// A presented Bearer token is not an anonymous scan: return 401 so
		// callers can distinguish a bad/disabled token from a wrong base path
		// (NoRoute still 404s). XHR keeps 401; bare unauthenticated stays 404.
		authHdr := c.GetHeader("Authorization")
		if strings.HasPrefix(authHdr, "Bearer ") || c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
		return
	}
	c.Next()
}

// monitorScopeAllow exposes only status/metrics reads. Keys are Gin route
// patterns relative to /panel/api, so a restricted token can never widen its
// reach by inventing a path parameter.
var monitorScopeAllow = map[string]map[string]struct{}{
	"/server/status":                              {http.MethodGet: {}},
	"/server/cpuHistory/:bucket":                  {http.MethodGet: {}},
	"/server/history/:metric/:bucket":             {http.MethodGet: {}},
	"/server/xrayMetricsState":                    {http.MethodGet: {}},
	"/server/xrayMetricsHistory/:metric/:bucket":  {http.MethodGet: {}},
	"/server/xrayObservatory":                     {http.MethodGet: {}},
	"/server/xrayObservatoryHistory/:tag/:bucket": {http.MethodGet: {}},
	"/server/getXrayVersion":                      {http.MethodGet: {}},
	"/server/getPanelUpdateInfo":                  {http.MethodGet: {}},
	"/nodes/history/:id/:metric/:bucket":          {http.MethodGet: {}},
}

// nodeSyncScopeAllow is the node-sync surface every master calls on a node.
// Each entry here is exercised by internal/web/runtime/remote.go or the node
// heartbeat probe; dropping one silently breaks cluster sync with a 403.
var nodeSyncScopeAllow = map[string]map[string]struct{}{
	"/server/status":               {http.MethodGet: {}},
	"/inbounds/list":               {http.MethodGet: {}},
	"/inbounds/add":                {http.MethodPost: {}},
	"/inbounds/del/:id":            {http.MethodPost: {}},
	"/inbounds/update/:id":         {http.MethodPost: {}},
	"/inbounds/:id/resetTraffic":   {http.MethodPost: {}},
	"/inbounds/resetAllTraffics":   {http.MethodPost: {}},
	"/inbounds/pushClientTraffics": {http.MethodPost: {}},
	"/clients/add":                 {http.MethodPost: {}},
	"/clients/del/:email":          {http.MethodPost: {}},
	"/clients/:email/detach":       {http.MethodPost: {}},
	"/clients/update/:email":       {http.MethodPost: {}},
	"/clients/resetTraffic/:email": {http.MethodPost: {}},
	"/clients/onlines":             {http.MethodPost: {}},
	"/clients/onlinesByGuid":       {http.MethodPost: {}},
	"/clients/lastOnline":          {http.MethodPost: {}},
	"/clients/clientIpsByGuid":     {http.MethodPost: {}},
	"/server/clientIps":            {http.MethodGet: {}, http.MethodPost: {}},
	"/server/restartXrayService":   {http.MethodPost: {}},
	"/server/getWebCertFiles":      {http.MethodGet: {}},
	"/server/descendants":          {http.MethodGet: {}},
	"/hosts/list":                  {http.MethodGet: {}},
}

// nodeAdminScopeAllow is the slice of node-sync that reaches beyond cluster
// plumbing into host-level software changes.
var nodeAdminScopeAllow = map[string]map[string]struct{}{
	"/server/updatePanel": {http.MethodPost: {}},
}

func scopeAllows(allow map[string]map[string]struct{}, path, method string) bool {
	methods, exists := allow[path]
	if !exists {
		return false
	}
	_, exists = methods[method]
	return exists
}

// enforceTokenScope applies explicit allowlists to restricted API tokens.
// Admin tokens and session-login users retain their existing behavior.
func (a *APIController) enforceTokenScope(c *gin.Context) {
	scopeVal, ok := c.Get("api_token_scope")
	if !ok {
		c.Next()
		return
	}
	scope, _ := scopeVal.(string)
	if scope == model.ApiScopeAdmin {
		c.Next()
		return
	}
	rel := relAPIPath(c.FullPath())
	switch scope {
	case model.ApiScopeMonitor:
		if scopeAllows(monitorScopeAllow, rel, c.Request.Method) {
			c.Next()
			return
		}
	case model.ApiScopeNodeSync:
		if scopeAllows(nodeSyncScopeAllow, rel, c.Request.Method) {
			c.Next()
			return
		}
	case model.ApiScopeNodeAdmin:
		if scopeAllows(nodeSyncScopeAllow, rel, c.Request.Method) || scopeAllows(nodeAdminScopeAllow, rel, c.Request.Method) {
			c.Next()
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"success": false,
		"msg":     "this API token is not permitted to access this endpoint",
	})
}

func relAPIPath(fullPath string) string {
	const marker = "/panel/api"
	_, after, ok := strings.Cut(fullPath, marker)
	if !ok {
		return ""
	}
	return after
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)
	api.Use(a.enforceTokenScope)
	// Decode + verify the node config envelope (zstd + X-Config-Sha256) and
	// advertise support, before CSRF/handlers read the body.
	api.Use(middleware.ConfigEnvelopeMiddleware())
	api.Use(middleware.CSRFMiddleware())

	api.GET("/openapi.json", ServeOpenAPISpec)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	clients := api.Group("/clients")
	NewClientController(clients)
	NewGroupController(clients)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Nodes API — multi-panel management
	nodes := api.Group("/nodes")
	a.nodeController = NewNodeController(nodes)

	// Hosts API — per-inbound override endpoints for subscription links
	hosts := api.Group("/hosts")
	a.hostController = NewHostController(hosts)

	// Settings + Xray config management live under the API surface too, so the
	// same API token drives them. Paths are /panel/api/setting/* and
	// /panel/api/xray/*.
	a.settingController = NewSettingController(api)
	a.xraySettingController = NewXraySettingController(api)

	// Extra routes
	api.POST("/backuptotgbot", a.BackuptoTgbot)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
