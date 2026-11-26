package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

func TestUpdateUserPreservesCurrentSessionAndRevokesOldSessions(t *testing.T) {
	engine, auth := newAPIAuthTestEngine(t)
	user, err := auth.userService.GetFirstUser()
	if err != nil {
		t.Fatalf("GetFirstUser: %v", err)
	}
	password, err := crypto.HashPasswordAsBcrypt("original-password")
	if err != nil {
		t.Fatalf("HashPasswordAsBcrypt: %v", err)
	}
	if err := database.GetDB().Model(user).Updates(map[string]any{
		"username": "original-admin", "password": password, "login_epoch": 7,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	settings := &SettingController{}
	engine.POST("/panel/api/setting/updateUser", auth.checkAPIAuth, settings.updateUser)

	login := httptest.NewRecorder()
	engine.ServeHTTP(login, httptest.NewRequest(http.MethodGet, "/test-login", nil))
	if login.Code != http.StatusOK {
		t.Fatalf("login status = %d, want 200", login.Code)
	}
	oldCookies := login.Result().Cookies()
	if len(oldCookies) == 0 {
		t.Fatal("login did not issue a session cookie")
	}

	req := httptest.NewRequest(http.MethodPost, "/panel/api/setting/updateUser", strings.NewReader(`{"oldUsername":"original-admin","oldPassword":"original-password","newUsername":"updated-admin","newPassword":"updated-password"}`))
	req.Header.Set("Content-Type", "application/json")
	for _, cookie := range oldCookies {
		req.AddCookie(cookie)
	}
	updated := httptest.NewRecorder()
	engine.ServeHTTP(updated, req)
	if updated.Code != http.StatusOK || !strings.Contains(updated.Body.String(), `"success":true`) {
		t.Fatalf("update status = %d, body = %s, want success", updated.Code, updated.Body.String())
	}
	newCookies := updated.Result().Cookies()
	if len(newCookies) == 0 {
		t.Fatal("update did not refresh the session cookie")
	}

	for _, tt := range []struct {
		name    string
		cookies []*http.Cookie
		want    int
	}{
		{name: "current session", cookies: newCookies, want: http.StatusOK},
		{name: "old session", cookies: oldCookies, want: http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/panel/api/ping", nil)
			req.Header.Set("X-Requested-With", "XMLHttpRequest")
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			if resp.Code != tt.want {
				t.Fatalf("status = %d, want %d; body = %s", resp.Code, tt.want, resp.Body.String())
			}
		})
	}
	stored, err := auth.userService.GetFirstUser()
	if err != nil {
		t.Fatalf("GetFirstUser after update: %v", err)
	}
	if stored.Username != "updated-admin" || stored.LoginEpoch != 8 || !crypto.CheckPasswordHash(stored.Password, "updated-password") {
		t.Fatalf("persisted credentials or login epoch do not match the update: username = %q, epoch = %d", stored.Username, stored.LoginEpoch)
	}
}

func TestValidateRegex(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewSettingController(router.Group("/panel/api"))

	tests := []struct {
		name    string
		body    string
		success bool
	}{
		{name: "Go RE2 inline flag", body: `{"regex":"(?m)^general-purpose$"}`, success: true},
		{name: "invalid expression", body: `{"regex":"["}`, success: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/panel/api/setting/validateRegex", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body=%s", resp.Code, resp.Body.String())
			}
			needle := `"success":true`
			if !tt.success {
				needle = `"success":false`
			}
			if !strings.Contains(resp.Body.String(), needle) {
				t.Fatalf("body = %s, want %s", resp.Body.String(), needle)
			}
		})
	}
}

// GHSA-xqqw-jqqv-99h6: saving settings with 2FA already on must not rebind the
// authenticator unless the request carries a code the current secret verifies.
func TestUpdateSettingRequiresCodeToReplaceTwoFactorToken(t *testing.T) {
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	if err := database.InitDB(filepath.Join(dbDir, "x-ui.db")); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { _ = database.CloseDB() })

	settingService := service.SettingService{}
	if err := settingService.SetTwoFactorToken("ORIGINALSECRET234567"); err != nil {
		t.Fatalf("seed token: %v", err)
	}
	if err := settingService.SetTwoFactorEnable(true); err != nil {
		t.Fatalf("seed enable: %v", err)
	}

	post := func(t *testing.T, mutate func(map[string]any)) string {
		t.Helper()
		base, err := settingService.GetAllSetting()
		if err != nil {
			t.Fatalf("GetAllSetting: %v", err)
		}
		raw, err := json.Marshal(base)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		body := map[string]any{}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		mutate(body)
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}

		gin.SetMode(gin.TestMode)
		router := gin.New()
		NewSettingController(router.Group("/panel/api"))
		req := httptest.NewRequest(http.MethodPost, "/panel/api/setting/update", strings.NewReader(string(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		return resp.Body.String()
	}

	t.Run("rebind without code is rejected", func(t *testing.T) {
		got := post(t, func(body map[string]any) {
			body["twoFactorEnable"] = true
			body["twoFactorToken"] = "ATTACKERSECRET567890"
		})
		if !strings.Contains(got, `"success":false`) {
			t.Fatalf("rebind without a 2FA code was accepted: %s", got)
		}
		stored, err := settingService.GetTwoFactorToken()
		if err != nil {
			t.Fatalf("GetTwoFactorToken: %v", err)
		}
		if stored != "ORIGINALSECRET234567" {
			t.Fatalf("stored token = %q, want the original secret", stored)
		}
	})

	t.Run("save with blank secret represerves the stored one", func(t *testing.T) {
		got := post(t, func(body map[string]any) {
			body["twoFactorEnable"] = true
			body["twoFactorToken"] = ""
		})
		if !strings.Contains(got, `"success":true`) {
			t.Fatalf("unrelated save was rejected: %s", got)
		}
		stored, err := settingService.GetTwoFactorToken()
		if err != nil {
			t.Fatalf("GetTwoFactorToken: %v", err)
		}
		if stored != "ORIGINALSECRET234567" {
			t.Fatalf("stored token = %q, want it preserved", stored)
		}
	})
}
