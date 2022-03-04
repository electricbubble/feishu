package feishu

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

var (
	appID             = os.Getenv("FeiShu-App-ID")
	appSecret         = os.Getenv("FeiShu-App-Secret")
	encryptKey        = os.Getenv("FeiShu-App-Encrypt-Key")
	verificationToken = os.Getenv("FeiShu-App-Verification-Token")
)

var _ Logger = (*debugLogger)(nil)

type debugLogger struct {
	*log.Logger
}

func (l *debugLogger) Debug(s string) {
	l.Println(s)
}

func requireNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func logIndent(t *testing.T, v interface{}) {
	t.Helper()

	bs, err := json.MarshalIndent(v, "", "  ")
	requireNil(t, err)

	t.Logf("\n%s", string(bs))
}

func testNewCustomApp() *app {
	fsApp := newApp(appID, appSecret,
		WithAppDebugLogger(&debugLogger{log.New(os.Stderr, "", log.LstdFlags)}),
		WithAppEventEncryptKey(encryptKey),
		WithAppEventVerificationToken(verificationToken),
	)
	fsApp.isCustomApp = true
	return fsApp
}
