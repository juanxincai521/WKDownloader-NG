package config

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/cihub/seelog"
)

func parseBookList() error {
	f, err := os.Open(data.ConfPath + "/booklist.txt")
	if err != nil {
		return err
	}
	data.BookList = make([]int, 0)
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		no, err2 := strconv.Atoi(line)
		if err2 != nil {
			return err2
		}
		data.BookList = append(data.BookList, no)
	}
}

func parseOldData() error {
	files, err := ioutil.ReadDir(data.TempPath + "/json")
	if err != nil {
		if os.IsNotExist(err) {
			seelog.Warn("无json数据目录")
			return nil
		}
		return err
	}
	infoFile := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			infoFile = append(infoFile, file.Name())
		}
	}
	length := len(infoFile)
	if length > 0 {
		sort.Strings(infoFile)
		jsonPath := data.TempPath + "/json/" + infoFile[length-1]
		seelog.Infof("解析旧数据Json文件：%s", jsonPath)
		infoFile, err := os.Open(jsonPath)
		if err != nil {
			return err
		}
		defer infoFile.Close()
		content, err2 := ioutil.ReadAll(infoFile)
		if err2 != nil {
			seelog.Warnf("列表文件读取失败，错误原因：%s", err2.Error())
			return err2
		}
		// // 用于恢复现场
		// data.NewData = &data.WKData{}
		// err3 := json.Unmarshal(content, data.NewData)
		data.OldData = &data.WKData{}
		err3 := json.Unmarshal(content, data.OldData)
		if err3 != nil {
			seelog.Warnf("列表文件解析失败，错误原因：%s", err3.Error())
			return err3
		}
	}
	return nil
}

func ParseConfig() error {
	err := parseBookList()
	if err != nil {
		return err
	}
	seelog.Infof("解析到%d个小说编号", len(data.BookList))
	err2 := parseOldData()
	if err2 != nil {
		return err2
	}
	err3 := utils.CheckFolderAndMake(data.DataPath)
	if err3 != nil {
		return err3
	}
	err4 := utils.CheckFolderAndMake(data.TempPath)
	if err4 != nil {
		return err4
	}
	return nil
}
