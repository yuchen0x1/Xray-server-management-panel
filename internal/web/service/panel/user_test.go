package panel

import (
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/crypto"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
)

func TestUpdateUserCommitsCredentialsAndTwoFactorTogether(t *testing.T) {
	for _, tt := range []struct {
		name    string
		trigger string
		wantErr string
	}{
		{name: "successful update"},
		{name: "credentials write fails", trigger: `CREATE TRIGGER reject_user_update BEFORE UPDATE ON users BEGIN SELECT RAISE(FAIL, 'credentials write blocked'); END`, wantErr: "credentials write blocked"},
		{name: "two factor write fails", trigger: `CREATE TRIGGER reject_two_factor_update BEFORE UPDATE ON settings WHEN OLD.key = 'twoFactorToken' BEGIN SELECT RAISE(FAIL, 'two factor write blocked'); END`, wantErr: "two factor write blocked"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := database.InitDB(filepath.Join(t.TempDir(), "x-ui.db")); err != nil {
				t.Fatalf("InitDB: %v", err)
			}
			t.Cleanup(func() { _ = database.CloseDB() })
			settings := &service.SettingService{}
			const originalSecret = "JBSWY3DPEHPK3PXP"
			if err := settings.SetTwoFactorEnable(true); err != nil {
				t.Fatalf("seed two factor enable: %v", err)
			}
			if err := settings.SetTwoFactorToken(originalSecret); err != nil {
				t.Fatalf("seed two factor token: %v", err)
			}
			svc := &UserService{}
			original, err := svc.GetFirstUser()
			if err != nil {
				t.Fatalf("GetFirstUser: %v", err)
			}
			if tt.trigger != "" {
				if err := database.GetDB().Exec(tt.trigger).Error; err != nil {
					t.Fatalf("create failure trigger: %v", err)
				}
			}
			err = svc.UpdateUser(original.Id, "updated-admin", "updated-password")
			if tt.wantErr == "" && err != nil {
				t.Fatalf("UpdateUser: %v", err)
			}
			if tt.wantErr != "" && (err == nil || err.Error() != tt.wantErr) {
				t.Fatalf("UpdateUser error = %v, want %q", err, tt.wantErr)
			}
			stored, err := svc.GetFirstUser()
			if err != nil {
				t.Fatalf("GetFirstUser after update: %v", err)
			}
			enabled, err := settings.GetTwoFactorEnable()
			if err != nil {
				t.Fatalf("GetTwoFactorEnable: %v", err)
			}
			secret, err := settings.GetTwoFactorToken()
			if err != nil {
				t.Fatalf("GetTwoFactorToken: %v", err)
			}
			if tt.wantErr != "" {
				if !enabled || secret != originalSecret {
					t.Fatalf("failed update changed two factor settings: enabled = %v, secret preserved = %v", enabled, secret == originalSecret)
				}
				if stored.Username != original.Username || stored.Password != original.Password || stored.LoginEpoch != original.LoginEpoch {
					t.Fatal("failed update changed persisted credentials or login epoch")
				}
				return
			}
			if enabled || secret != "" {
				t.Fatalf("successful credential change did not reset two factor settings: enabled = %v, secret empty = %v", enabled, secret == "")
			}
			if stored.Username != "updated-admin" || stored.LoginEpoch != original.LoginEpoch+1 || !crypto.CheckPasswordHash(stored.Password, "updated-password") {
				t.Fatal("successful update did not persist the new credentials and increment login epoch")
			}
			var count int64
			if err := database.GetDB().Model(model.User{}).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("user count = %d, error = %v; want one user", count, err)
			}
		})
	}
}
