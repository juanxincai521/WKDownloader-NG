package copyright

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/cihub/seelog"
)

var (
	history map[int]*data.Pic
	pool    *utils.SimplePool
	count   int
	mutex   sync.Mutex
	fail    bool
)

func loadCopyrightPicHistory() error {
	history = make(map[int]*data.Pic)
	historyFilePath := data.ConfPath + "/history.json"
	historyFile, err := os.Open(historyFilePath)
	if err != nil {
		return err
	}
	defer historyFile.Close()
	content, err2 := ioutil.ReadAll(historyFile)
	if err2 != nil {
		seelog.Warnf("历史文件读取失败，错误原因：%s", err2.Error())
		return err2
	}
	err3 := json.Unmarshal(content, &history)
	if err3 != nil {
		seelog.Warnf("历史文件解析失败，错误原因：%s", err3.Error())
		return err3
	}
	appendPic(history)
	return nil
}

func appendPic(list map[int]*data.Pic) {
	for chapterNo, pic := range list {
		_, isCopyright := data.NewData.CopyrightChapters[chapterNo]
		if !isCopyright {
			continue
		}
		existPic, ok := data.NewData.Pics[chapterNo]
		if !ok {
			data.NewData.Pics[chapterNo] = pic
		} else {
			for _, img := range pic.PicNo {
				flag := false
				for _, existImg := range existPic.PicNo {
					if existImg.ImgNo == img.ImgNo && existImg.Extend == img.Extend {
						flag = true
						break
					}
				}
				if flag {
					continue
				}
				existPic.PicNo = append(existPic.PicNo, img)
			}
		}
	}
}

func checkPicLink(page *data.Page) {
	defer pool.Done()
	minNearestChapter := 0
	delta := 999999
	minNo := 999999
	for chapterNo := range data.NewData.Pics {
		tmp := page.ChapterNo - chapterNo
		if chapterNo < minNo {
			minNo = chapterNo
		}
		if tmp < 0 {
			continue
		}
		if tmp < delta {
			delta = tmp
			minNearestChapter = chapterNo
		}
	}
	if minNearestChapter == 0 {
		minNearestChapter = minNo
	}
	seelog.Debugf("最接近的最小章节%d", minNearestChapter)
	maxNearestChapter := 0
	delta = 999999
	maxNo := 0
	for chapterNo := range data.NewData.Pics {
		tmp := page.ChapterNo - chapterNo
		if chapterNo > maxNo {
			maxNo = chapterNo
		}
		if tmp <= 0 {
			tmp = -tmp
		} else {
			continue
		}
		if tmp < delta {
			delta = tmp
			maxNearestChapter = chapterNo
		}
	}
	if maxNearestChapter == 0 {
		maxNearestChapter = maxNo
	}
	seelog.Debugf("最接近的最大章节%d", maxNearestChapter)
	minimalPicNo := 999999
	for _, picNo := range data.NewData.Pics[minNearestChapter].PicNo {
		if picNo.ImgNo < minimalPicNo {
			minimalPicNo = picNo.ImgNo
		}
	}
	if minimalPicNo > 50 {
		minimalPicNo -= 100
	}
	maximumPicNo := 0
	for _, picNo := range data.NewData.Pics[maxNearestChapter].PicNo {
		if picNo.ImgNo > minimalPicNo {
			maximumPicNo = picNo.ImgNo
		}
	}
	maximumPicNo += 100
	a := page.BookNo / 1000
	url := fmt.Sprintf("http://pic.wenku8.com/pictures/%d/%d/%d/", a, page.BookNo, page.ChapterNo)
	seelog.Infof("%d-测试链接区间：min-%d,max-%d", page.ChapterNo, minimalPicNo, maximumPicNo)
	for i := minimalPicNo; i < maximumPicNo; i++ {
		if testurl(fmt.Sprintf("%s%d.%s", url, i, "jpg")) {
			addPic(page.BookNo, page.ChapterNo, i, "jpg")
			continue
		}
		if testurl(fmt.Sprintf("%s%d.%s", url, i, "jpeg")) {
			addPic(page.BookNo, page.ChapterNo, i, "jpeg")
			continue
		}
		if testurl(fmt.Sprintf("%s%d.%s", url, i, "png")) {
			addPic(page.BookNo, page.ChapterNo, i, "png")
			continue
		}
	}
}

func addPic(bookNo int, chapterNo int, picNo int, extend string) {
	mutex.Lock()
	defer mutex.Unlock()
	_, ok := data.NewData.Pics[chapterNo]
	if !ok {
		data.NewData.Pics[chapterNo] = &data.Pic{BookNo: bookNo, ChapterNo: chapterNo}
		data.NewData.Pics[chapterNo].PicNo = make([]*data.Img, 0)
	}
	data.NewData.Pics[chapterNo].PicNo = append(data.NewData.Pics[chapterNo].PicNo, &data.Img{ImgNo: picNo, Extend: extend})
	count++
}

func testurl(url string) bool {
	for j := 1; j <= 100; j++ {
		res, err := utils.GetTestHTTPClient().Client.Get(url)
		if err == nil {
			defer res.Body.Close()
			if res.StatusCode == 404 {
				seelog.Debugf("测试链接：%s，第%d次，确认不存在", url, j)
				return false
			} else if res.StatusCode >= 200 && res.StatusCode <= 400 {
				seelog.Infof("测试链接：%s，第%d次，确认存在", url, j)
				return true
			} else {
				seelog.Warnf("测试链接：%s，第%d次失败", url, j)
			}
		}
	}
	seelog.Warnf("测试链接：%s，失败", url)
	fail = true
	return false
}

func FetchCopyrightPic() error {
	err := loadCopyrightPicHistory()
	if err != nil {
		return err
	}
	if data.OldData != nil {
		appendPic(data.OldData.Pics)
	}
	pool = utils.NewPool(runtime.NumCPU())
	count = 0
	fail = false
	for chapterNo, chapterPage := range data.NewData.CopyrightChapters {
		if strings.Contains(chapterPage.ChapterName, "插图") {
			_, ok := data.NewData.Pics[chapterNo]
			if ok {
				continue
			}
			pool.Add(1)
			go checkPicLink(chapterPage)
		}
	}
	pool.Wait()
	seelog.Infof("共获得%d张版权插图", count)
	if fail {
		return errors.New("测试插图链接失败")
	}
	return nil
}
