package image

import (
	"os"
	"testing"

	"github.com/ptrost/mosaic2go/test"
)

func TestS3FileStorageSaveDelete(t *testing.T) {
	cfg := getConfig()
	storage := NewS3FileStorage(cfg.Get("s3_region"), cfg.Get("s3_images_bucket_test"))
	filename := "test.jpg"
	file, err := os.Open("../test_fixtures/test.jpg")
	test.AssertNotErr("osOpen", err, t)
	_, errSave := storage.Save(file, filename)
	test.AssertNotErr("S3FileStorage.Save", errSave, t)
	errDelete := storage.Delete(filename)
	test.AssertNotErr("S3FileStorage.Delete", errDelete, t)
}
