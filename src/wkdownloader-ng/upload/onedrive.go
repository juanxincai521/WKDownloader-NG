package upload

import (
	"os/exec"
)

func UploadToOnedrive() error {
	cmd := exec.Command("sh", "/home/jason/WKDownloader-NG/sync-to-onedrive.sh")
	return cmd.Run()
}
