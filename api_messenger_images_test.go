package feishu

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_app_UploadImage(t *testing.T) {
	fsApp := testNewCustomApp()

	_, err := fsApp.getAppAccessTokenWithContext(context.Background())
	requireNil(t, err)

	fsApp.opt.debug = true

	uhDir, err := os.UserHomeDir()
	requireNil(t, err)

	// f, err := os.Open(filepath.Join(uhDir, "/Pictures/IMG_0806.jpg"))
	// requireNil(t, err)
	// defer f.Close()
	//
	// imgKey, err := fsApp.UploadImage(WithUploadImageViaReader("test.jpg", f))
	// requireNil(t, err)

	imgKey, err := fsApp.UploadImage(WithUploadImage(filepath.Join(uhDir, "/Pictures/dot-23s.png")))
	requireNil(t, err)

	t.Log(imgKey)
}
