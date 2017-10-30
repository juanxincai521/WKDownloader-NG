package utils

import (
	"os"
	"strings"

	"github.com/cihub/seelog"
)

func CheckFolderAndMake(dirPath string) error {
	seelog.Debugf("尝试建立文件夹%s", dirPath)
	_, check := os.Stat(dirPath)
	if check != nil {
		if !os.IsNotExist(check) {
			return check
		}
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
		seelog.Debugf("建立文件夹%s", dirPath)
	}
	seelog.Debugf("已存在文件夹%s", dirPath)
	return nil
}

func CleanFileName(s string) string {
	s = strings.Replace(s, "/", "／", -1)
	s = strings.Replace(s, ":", "：", -1)
	s = strings.Replace(s, ":", "：", -1)
	s = strings.Replace(s, "?", "？", -1)
	s = strings.Replace(s, "<", "《", -1)
	s = strings.Replace(s, ">", "》", -1)
	s = strings.Replace(s, "*", "", -1)
	s = strings.Replace(s, "|", "", -1)
	s = strings.Replace(s, `\`, "", -1)
	s = strings.Replace(s, `"`, "", -1)
	return s
}
