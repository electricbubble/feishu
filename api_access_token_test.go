package feishu

import (
	"testing"
)

func Test_app_GetAppAccessTokenInternal(t *testing.T) {
	fsApp := testNewCustomApp()
	fsApp.opt.debug = true
	accessToken, err := fsApp.GetAppAccessTokenInternal()
	requireNil(t, err)

	logIndent(t, accessToken)
}

func Test_app_GetTenantAccessTokenInternal(t *testing.T) {
	fsApp := testNewCustomApp()
	fsApp.opt.debug = true

	tenantAccessToken, err := fsApp.GetTenantAccessTokenInternal()
	requireNil(t, err)

	logIndent(t, tenantAccessToken)
}
