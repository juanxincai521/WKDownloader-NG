package page

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/cihub/seelog"
)

var (
	lastPageInfo map[int]*data.Page
	newPageInfo  map[int]*data.Page
	needCompare  bool
	pool         *utils.SimplePool
	mutex        sync.Mutex
	fail         bool
)

func downloadPage(no int) {
	defer pool.Done()
	bookNo := newPageInfo[no].BookNo
	chapterNo := no
	seelog.Tracef("开始下载%d.htm，编号%d", chapterNo, bookNo)
	a := bookNo / 1000
	url := fmt.Sprintf("http://www.wenku8.com/novel/%d/%d/%d.htm", a, bookNo, chapterNo)
	filePath := fmt.Sprintf("%s/%s/%d/%d.htm", data.TempPath, "page", bookNo, chapterNo)
	_, err := os.Stat(filePath)
	if err == nil {
		if needCompare {
			lastPage, ok := lastPageInfo[chapterNo]
			newPage, _ := newPageInfo[chapterNo]
			if ok {
				if lastPage.ChapterName == newPage.ChapterName && lastPage.BookName == newPage.BookName {
					seelog.Tracef("跳过下载%d.htm，编号%d", chapterNo, bookNo)
					return
				}
			}
		} else {
			os.Remove(filePath)
			seelog.Debugf("清除残留文件：%s", filePath)
			// seelog.Debugf("跳过下载%d.htm，编号%d", chapterNo, bookNo)
			// return
		}
	}
	for i := 1; i <= 100; i++ {
		err := utils.Download(url, filePath)
		if err != nil {
			seelog.Warnf("%d.htm，编号%d，尝试第%d次失败，错误原因：%s", chapterNo, bookNo, i, err.Error())
		} else {
			seelog.Debugf("%d.htm，编号%d，尝试第%d次，下载成功", chapterNo, bookNo, i)
			return
		}
	}
	os.Remove(filePath)
	seelog.Errorf("%s", url)
	seelog.Warnf("%d.htm，编号%d，下载失败", chapterNo, bookNo)
	fail = true
}

func parsePages(bookNo int) {
	defer pool.Done()
	pages, err := ioutil.ReadDir(fmt.Sprintf("%s/%s/%d", data.TempPath, "page", bookNo))
	if err != nil {
		seelog.Warnf("列出页面失败，错误原因：%s", err.Error())
		fail = true
	}
	for _, page := range pages {
		chapterPath := strings.Split(page.Name(), ".")[0]
		chapterNo, err2 := strconv.Atoi(chapterPath)
		if err2 != nil {
			seelog.Warnf("列出页面失败，错误原因：%s", err.Error())
			fail = true
		}
		parsePage(bookNo, chapterNo)
	}
}

func parsePage(bookNo int, chapterNo int) {
	seelog.Tracef("开始解析%d.htm，编号%d", chapterNo, bookNo)
	filePath := fmt.Sprintf("%s/%s/%d/%d.htm", data.TempPath, "page", bookNo, chapterNo)
	file, err := os.Open(filePath)
	if err != nil {
		seelog.Warnf("文件打开失败，错误原因：%s", err.Error())
		fail = true
		return
	}
	contentGBK, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		seelog.Warnf("文件打开失败，错误原因：%s", err2.Error())
		fail = true
		return
	}
	content, err3 := utils.GBKToUTF8(contentGBK)
	if err3 != nil {
		seelog.Warnf("文件打开失败，错误原因：%s", err3.Error())
		fail = true
		return
	}
	fileStat, _ := file.Stat()
	reader, err4 := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err4 != nil {
		seelog.Warnf("htm解析失败，错误原因：%s", err4.Error())
		fail = true
		return
	}
	if fileStat.Size() < 10240 {
		if bytes.Contains(content, []byte("版权")) {
			page := &data.Page{}
			page.BookNo = bookNo
			page.ChapterNo = chapterNo
			page.ChapterName = reader.Find("title").Text()
			mutex.Lock()
			defer mutex.Unlock()
			data.NewData.CopyrightChapters[chapterNo] = page
			seelog.Tracef("解析%d.htm成功，编号%d", chapterNo, bookNo)
			return
		}
	}
	picNos := make([]*data.Img, 0)
	reader.Find("img").Each(
		func(index int, node *goquery.Selection) {
			src, exist := node.Attr("src")
			if exist {
				if strings.Contains(src, "pic") {
					params := strings.Split(src, "/")
					picFile := strings.Split(params[len(params)-1], ".")
					picNo, err5 := strconv.Atoi(picFile[0])
					if err5 != nil {
						seelog.Warnf("%d图片链接解析失败，错误原因：%s", chapterNo, err5.Error())
						//fail = true
						return
					}
					img := &data.Img{ImgNo: picNo, Extend: picFile[1]}
					picNos = append(picNos, img)
				}
			}
		})
	if len(picNos) > 0 {
		chapterPic := &data.Pic{BookNo: bookNo, ChapterNo: chapterNo, PicNo: picNos}
		mutex.Lock()
		defer mutex.Unlock()
		data.NewData.Pics[chapterNo] = chapterPic
	}
	seelog.Tracef("解析%d.htm成功，编号%d", chapterNo, bookNo)
}

func GetAndParsePage() error {
	fail = false
	pool = utils.NewPool(10)
	newPageInfo = data.BVCToPage(data.NewData.Books)
	if data.OldData == nil {
		needCompare = false
	} else {
		needCompare = true
		lastPageInfo = data.BVCToPage(data.OldData.Books)
	}
	err := utils.CheckFolderAndMake(data.TempPath + "/page")
	if err != nil {
		return err
	}
	seelog.Info("开始下载页面")
	for pageNo := range newPageInfo {
		pool.Add(1)
		go downloadPage(pageNo)
	}
	pool.Wait()
	if fail {
		return errors.New("下载页面失败")
	}
	seelog.Info("下载页面完成")
	seelog.Info("开始解析页面")
	pool = utils.NewPool(runtime.NumCPU())
	data.NewData.Pics = make(map[int]*data.Pic)
	data.NewData.CopyrightChapters = make(map[int]*data.Page)
	for _, book := range data.BookList {
		pool.Add(1)
		go parsePages(book)
	}
	pool.Wait()
	seelog.Info("解析页面完成")
	if fail {
		return errors.New("解析页面失败")
	}
	return nil
}
