package download

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"

	"github.com/cihub/seelog"
)

var (
	pool *utils.SimplePool
	fail bool
)

func downloadBook(no int, ext string) {
	defer pool.Done()
	seelog.Debugf("开始下载%d.%s", no, ext)
	a := no / 1000
	ftype := "txtgbk"
	umdMark := ""
	if ext == "umd" {
		ftype = "umd"
		umdMark = strconv.Itoa(no) + "/"
	}
	url := fmt.Sprintf("http://dl.wkcdn.com/%s/%d/%s%d.%s", ftype, a, umdMark, no, ext)
	folder := data.DataPath + "/txt"
	if ext == "umd" {
		folder = data.DataPath + "/umd"
	}
	filePath := fmt.Sprintf("%s/%d.%s", folder, no, ext)
	bookDir, err := ioutil.ReadDir(data.DataPath + "/txt")
	if err != nil {
		fail = true
		return
	}
	for _, book := range bookDir {
		if strings.HasPrefix(book.Name(), strconv.Itoa(no)+"_") && strings.HasSuffix(book.Name(), ext) {
			if book.Size() == utils.GetFileSize(url) {
				err := os.Rename(data.DataPath+"/txt/"+book.Name(), filePath)
				if err != nil {
					fail = true
					return
				}
				seelog.Debugf("跳过未更新文本文件：%d", no)
				return
			}
			break
		}
	}
	for i := 1; i <= 100; i++ {
		err := utils.Download(url, filePath)
		if err != nil {
			seelog.Warnf("%d.%s，尝试第%d次失败，错误原因：%s", no, ext, i, err.Error())
		} else {
			seelog.Debugf("%d.%s，尝试第%d次，下载成功，url：%s", no, ext, i, url)
			return
		}
	}
	os.Remove(filePath)
	seelog.Errorf("%s", url)
	seelog.Warnf("%d.%s，下载失败", no, ext)
	fail = true
}

func downloadImg(bookNo int, chapterNo int, img *data.Img) {
	defer pool.Done()
	seelog.Debugf("开始下载%d/%d/%d.%s", bookNo, chapterNo, img.ImgNo, img.Extend)
	a := bookNo / 1000
	url := fmt.Sprintf("http://pic.wkcdn.com/pictures/%d/%d/%d/%d.%s", a, bookNo, chapterNo, img.ImgNo, img.Extend)
	filePath := fmt.Sprintf("%s/%d/%d/%d.%s", data.TempPath+"/pic", bookNo, chapterNo, img.ImgNo, img.Extend)
	for i := 1; i <= 100; i++ {
		err := utils.Download(url, filePath)
		if err != nil {
			seelog.Warnf("%s，尝试第%d次失败，错误原因：%s", filePath, i, err.Error())
		} else {
			seelog.Debugf("%s，尝试第%d次，下载成功", filePath, i)
			return
		}
	}
	os.Remove(filePath)
	seelog.Errorf("%s", url)
	seelog.Warnf("%s，下载失败", filePath)
	fail = true
}

func DownloadTXT() error {
	err := utils.CheckFolderAndMake(data.DataPath + "/txt")
	if err != nil {
		return err
	}
	pool = utils.NewPool(10)
	fail = false
	for _, no := range data.BookList {
		pool.Add(1)
		go downloadBook(no, "txt")
	}
	pool.Wait()
	if fail {
		return errors.New("下载TXT失败")
	}
	return nil
}

func DownloadPic() error {
	err := utils.CheckFolderAndMake(data.TempPath + "/pic")
	if err != nil {
		return err
	}
	pool = utils.NewPool(10)
	fail = false
	if data.OldData == nil {
		for _, pic := range data.NewData.Pics {
			for _, img := range pic.PicNo {
				pool.Add(1)
				go downloadImg(pic.BookNo, pic.ChapterNo, img)
			}
		}
	} else {
		for _, pic := range data.NewData.Pics {
			old, has := data.OldData.Pics[pic.ChapterNo]
			if has {
				for _, img := range pic.PicNo {
					exist := false
					for _, oldImg := range old.PicNo {
						if img.ImgNo == oldImg.ImgNo && img.Extend == oldImg.Extend {
							exist = true
							break
						}
					}
					if exist {
						continue
					}
					pool.Add(1)
					go downloadImg(pic.BookNo, pic.ChapterNo, img)
				}
			} else {
				for _, img := range pic.PicNo {
					pool.Add(1)
					go downloadImg(pic.BookNo, pic.ChapterNo, img)
				}
			}
		}
	}
	pool.Wait()
	if fail {
		return errors.New("下载图片失败")
	}
	return nil
}
