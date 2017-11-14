package finish

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"
	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/cihub/seelog"
)

func AtEnd() error {
	timeString := time.Now().Format("2006-01-02-15-04-05")
	err := outputJSONAndClean(timeString)
	if err != nil {
		return err
	}
	seelog.Info("输出记录文件")
	err = packageIndexAndClean(timeString)
	if err != nil {
		return err
	}
	seelog.Info("保存index.htm")
	err = os.RemoveAll(data.TempPath + "/cache")
	if err != nil {
		return err
	}
	seelog.Info("清除index缓存")
	err = os.RemoveAll(data.TempPath + "/pic")
	if err != nil {
		return err
	}
	seelog.Info("清除图片缓存")
	return nil
}

func outputJSONAndClean(timeString string) error {
	err := utils.CheckFolderAndMake(data.TempPath + "/json")
	if err != nil {
		return err
	}
	dataJSON, err := json.MarshalIndent(&data.NewData, "", "\t")
	if err != nil {
		return err
	}
	file, err2 := os.Create(data.TempPath + "/json/" + timeString + ".json")
	if err2 != nil {
		return err2
	}
	_, err3 := fmt.Fprintf(file, "%s", dataJSON)
	if err3 != nil {
		return err3
	}
	seelog.Info("输出记录文件：" + data.TempPath + "/json/" + timeString + ".json")
	jsonDir, err4 := ioutil.ReadDir(data.TempPath + "/json")
	if err4 != nil {
		return err4
	}
	if len(jsonDir) > 5 {
		seelog.Info("清理多余记录")
		names := make([]string, 0)
		for _, jsonFile := range jsonDir {
			names = append(names, jsonFile.Name())
		}
		sort.Strings(names)
		for i := 0; i < len(jsonDir)-5; i++ {
			err := os.Remove(data.TempPath + "/json/" + names[i])
			if err != nil {
				return err
			}
			seelog.Info("删除多余记录：" + data.TempPath + "/json/" + names[i])
		}
	}
	return nil
}

func packageIndexAndClean(timeString string) error {
	err := utils.CheckFolderAndMake(data.TempPath + "/index")
	if err != nil {
		return err
	}
	indexDir, err2 := ioutil.ReadDir(data.TempPath + "/cache")
	if err2 != nil {
		if os.IsNotExist(err2) {
			seelog.Info("无index文件")
			return nil
		}
		return err2
	}
	zipFile, err3 := os.Create(data.TempPath + "/index/" + timeString + ".zip")
	if err3 != nil {
		return err3
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()
	for _, indexFile := range indexDir {
		seelog.Debug("添加文件" + data.TempPath + "/cache/" + indexFile.Name())
		fw, err := writer.Create(indexFile.Name())
		if err != nil {
			return err
		}
		in, err := ioutil.ReadFile(data.TempPath + "/cache/" + indexFile.Name())
		if err != nil {
			return err
		}
		fw.Write(in)
		seelog.Debug("归档index.htm成功")
	}
	seelog.Info("建立index.htm压缩包：" + data.TempPath + "/index/" + timeString + ".zip")
	jsonDir, err4 := ioutil.ReadDir(data.TempPath + "/index")
	if err4 != nil {
		return err4
	}
	if len(jsonDir) > 5 {
		seelog.Info("清理多余index包")
		names := make([]string, 0)
		for _, jsonFile := range jsonDir {
			names = append(names, jsonFile.Name())
		}
		sort.Strings(names)
		for i := 0; i < len(jsonDir)-5; i++ {
			err := os.Remove(data.TempPath + "/index/" + names[i])
			if err != nil {
				return err
			}
			seelog.Info("删除多余index包：" + data.TempPath + "/json/" + names[i])
		}
	}
	return nil
}
