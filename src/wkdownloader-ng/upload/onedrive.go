package upload

import (
	"os/exec"
)

func UploadToOnedrive() error {
	cmd := exec.Command("/home/jason/WKDownloader-NG/sync-to-onedrive.sh >/home/jason/WKDownloader-NG/logs/sync-to-onedrive.log 2>&1 &")
	return cmd.Run()
}
