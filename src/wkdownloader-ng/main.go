package main

import (
	"runtime"

	"wkdownloader-ng/config"
	"wkdownloader-ng/copyright"
	"wkdownloader-ng/data"
	"wkdownloader-ng/download"
	"wkdownloader-ng/finish"
	"wkdownloader-ng/index"
	"wkdownloader-ng/page"
	"wkdownloader-ng/rename"
	"wkdownloader-ng/upload"

	"github.com/cihub/seelog"
)

func main() {
	// 配置运行时
	runtime.GOMAXPROCS(runtime.NumCPU())

	// 配置logger
	logger, err := seelog.LoggerFromConfigAsFile(data.ConfPath + "/seelog.xml")
	if err != nil {
		panic(err)
	}
	seelog.ReplaceLogger(logger)
	defer logger.Flush()
	seelog.Info("任务开始")

	// 获取初始配置
	seelog.Info("开始解析配置文件")
	err = config.ParseConfig()
	if err != nil {
		seelog.Error("解析配置文件失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("解析配置文件完成")

	// seelog.Info("设置代理")
	// os.Setenv("http_proxy", "http://127.0.0.1:8118")
	// os.Setenv("https_proxy", "http://127.0.0.1:8118")

	// 下载index.htm
	seelog.Info("开始下载index.htm")
	err = index.GetAndParseIndex()
	if err != nil {
		seelog.Error("下载index.htm失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("下载index.htm结束")

	// 下载并解析分页
	seelog.Info("开始下载并解析分页")
	err = page.GetAndParsePage()
	if err != nil {
		seelog.Error("下载并解析分页失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("下载并解析分页完成")

	// os.Unsetenv("http_proxy")
	// os.Unsetenv("https_proxy")
	// seelog.Info("解除代理")

	// 获取版权插图
	seelog.Info("开始获取版权插图")
	err = copyright.FetchCopyrightPic()
	if err != nil {
		seelog.Error("获取版权插图失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("获取版权插图完成")

	// // 额外记录
	// data, _ := json.Marshal(data.NewData)
	// seelog.Debug(string(data))

	// 下载TXT
	seelog.Info("开始下载TXT")
	err = download.DownloadTXT()
	if err != nil {
		seelog.Error("下载TXT失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("下载TXT完成")

	// 下载插图
	seelog.Info("开始下载插图")
	err = download.DownloadPic()
	if err != nil {
		seelog.Error("下载插图失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("下载插图完成")

	// 重命名TXT
	seelog.Info("开始重命名TXT")
	err = rename.RenameTXT()
	if err != nil {
		seelog.Error("重命名TXT失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("重命名TXT完成")

	// 归档插图
	seelog.Info("开始归档插图")
	err = rename.RenameAndPackagePic()
	if err != nil {
		seelog.Error("归档插图失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("归档插图完成")

	// 数据整理
	seelog.Info("开始数据整理")
	err = finish.AtEnd()
	if err != nil {
		seelog.Error("数据整理失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("数据整理完成")

	// 上传到Onedrive
	seelog.Info("开始上传到Onedrive")
	err = upload.UploadToOnedrive()
	if err != nil {
		seelog.Error("上传到Onedrive失败：" + err.Error())
		panic(err.Error())
	}
	seelog.Info("上传到Onedrive完成")

	seelog.Info("任务结束")
}
