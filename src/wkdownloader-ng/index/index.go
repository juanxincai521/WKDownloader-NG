package index

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/cihub/seelog"
)

var (
	pool  *utils.SimplePool
	mutex sync.Mutex
	fail  bool
)

func downloadIndex(no int) {
	seelog.Debugf("开始下载index.htm，编号%d", no)
	a := no / 1000
	url := fmt.Sprintf("http://www.wenku8.net/novel/%d/%d/index.htm", a, no)
	filePath := fmt.Sprintf("%s/%s/%d.htm", data.TempPath, "cache", no)
	for i := 1; i <= 100; i++ {
		err := utils.DownloadWithProxy(url, filePath)
		if err != nil {
			seelog.Warnf("index.htm，编号%d，尝试第%d次失败，错误原因：%s", no, i, err.Error())
		} else {
			seelog.Debugf("index.htm，编号%d，尝试第%d次，下载成功", no, i)
			return
		}
	}
	os.Remove(filePath)
	seelog.Errorf("%s", url)
	seelog.Warnf("index.htm，编号%d，下载失败", no)
	fail = true
}

func parseIndex(no int) {
	seelog.Debugf("开始解析index.htm，编号%d", no)

	book := &data.Book{BookNo: no}

	indexFilePath := fmt.Sprintf("%s/%s/%d.htm", data.TempPath, "cache", no)
	indexFile, err := os.Open(indexFilePath)
	if err != nil {
		seelog.Warnf("解析index.htm失败，错误原因：%s", err.Error())
		fail = true
	}
	defer indexFile.Close()
	raw, err2 := ioutil.ReadAll(indexFile)
	if err2 != nil {
		seelog.Warnf("解析index.htm失败，错误原因：%s", err2.Error())
		fail = true
	}
	utf8data, err3 := utils.GBKToUTF8(raw)
	if err3 != nil {
		seelog.Warnf("解析index.htm失败，错误原因：%s", err3.Error())
		fail = true
	}
	doc, err4 := goquery.NewDocumentFromReader(bytes.NewReader(utf8data))
	if err4 != nil {
		seelog.Warnf("解析index.htm失败，错误原因：%s", err4.Error())
		fail = true
	}
	book.BookName = doc.Find("#title").Text()
	var volume *data.Volume
	doc.Find("td").Each(
		func(index int, node *goquery.Selection) {
			if node.HasClass("vcss") {
				volume = &data.Volume{}
				volume.VolumeName = node.Text()
				book.Volumes = append(book.Volumes, volume)
			}
			if node.HasClass("ccss") {
				chapter := &data.Chapter{}
				chapterNode := node.Find("a")
				chapterPath, exist := chapterNode.Attr("href")
				if !exist {
					return
				}
				chapterStr := strings.Split(chapterPath, ".")
				chapterNo, err := strconv.Atoi(chapterStr[0])
				if err != nil {
					return
				}
				chapter.ChapterNo = chapterNo
				chapter.ChapterName = chapterNode.Text()
				volume.Chapters = append(volume.Chapters, chapter)
			}
		})
	mutex.Lock()
	defer mutex.Unlock()
	data.NewData.Books = append(data.NewData.Books, book)
	seelog.Debugf("解析index.htm成功，编号%d", no)
}

func getIndex(no int) {
	defer pool.Done()
	downloadIndex(no)
	parseIndex(no)
}

func GetAndParseIndex() error {
	fail = false
	pool = utils.NewPool(10)
	data.NewData = &data.WKData{}
	data.NewData.Books = make([]*data.Book, 0)
	err := utils.CheckFolderAndMake(data.TempPath + "/cache")
	if err != nil {
		return err
	}
	for _, no := range data.BookList {
		pool.Add(1)
		go getIndex(no)
	}
	pool.Wait()
	if fail {
		return errors.New("下载index.htm失败")
	}
	return nil
}
