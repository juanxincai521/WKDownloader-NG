package rename

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/cihub/seelog"

	"wkdownloader-ng/data"
	"wkdownloader-ng/utils"
)

func RenameTXT() error {
	files, err := ioutil.ReadDir(data.DataPath + "/txt")
	if err != nil {
		return err
	}
	for _, file := range files {
		if strings.Contains(file.Name(), "_") {
			seelog.Debugf("清除残留文件：" + file.Name())
			err = os.Remove(data.DataPath + "/txt/" + file.Name())
			if err != nil {
				return err
			}
			continue
		}
		bookNo, ok := strconv.Atoi(strings.Replace(file.Name(), ".txt", "", 1))
		if ok != nil {
			seelog.Debugf("清除异常文件：" + file.Name())
			err = os.Remove(data.DataPath + "/txt/" + file.Name())
			if err != nil {
				return err
			}
			continue
		}
		remove := true
		for _, no := range data.BookList {
			if bookNo == no {
				remove = false
				break
			}
		}
		if remove {
			seelog.Debugf("清除错误文件：" + file.Name())
			err = os.Remove(data.DataPath + "/txt/" + file.Name())
			if err != nil {
				return err
			}
		}
	}
	files, err = ioutil.ReadDir(data.DataPath + "/txt")
	if err != nil {
		return err
	}
	for _, file := range files {
		seelog.Debugf("开始重命名文件：" + file.Name())
		bookNo, _ := strconv.Atoi(strings.Replace(file.Name(), ".txt", "", 1))
		for _, book := range data.NewData.Books {
			if bookNo == book.BookNo {
				unquoteFileName, err := strconv.Unquote(`"` + book.BookName + `"`)
				if err != nil {
					return err
				}
				newBookName := strconv.Itoa(bookNo) + "_" + unquoteFileName + ".txt"
				newBookName = utils.CleanFileName(newBookName)
				seelog.Debugf("重命名%s至%s", file.Name(), newBookName)
				err2 := os.Rename(data.DataPath+"/txt/"+file.Name(), data.DataPath+"/txt/"+newBookName)
				if err2 != nil {
					return err2
				}
			}
		}
	}
	return nil
}

func RenameAndPackagePic() error {
	err := utils.CheckFolderAndMake(data.DataPath + "/pic")
	if err != nil {
		return err
	}
	bookDirs, err := ioutil.ReadDir(data.DataPath + "/pic")
	if err != nil {
		return err
	}
	for _, bookDir := range bookDirs {
		d := strings.Split(bookDir.Name(), "_")
		if d == nil || d[0] == "" {
			continue
		}
		seelog.Debugf("重置文件夹名" + data.DataPath + "/pic/" + d[0])
		err := os.Rename(data.DataPath+"/pic/"+bookDir.Name(), data.DataPath+"/pic/"+d[0])
		if err != nil {
			return err
		}
	}
	_, err = os.Stat(data.TempPath + "/pic")
	if err != nil {
		if os.IsNotExist(err) {
			seelog.Info("无新图片")
		} else {
			return err
		}
	} else {
		doErr := doPackAndRename()
		if doErr != nil {
			return doErr
		}
	}
	bookDirs2, err := ioutil.ReadDir(data.DataPath + "/pic")
	if err != nil {
		return err
	}
	for _, bookDir := range bookDirs2 {
		for _, book := range data.NewData.Books {
			if strconv.Itoa(book.BookNo) == bookDir.Name() {
				bookName, err := strconv.Unquote(`"` + book.BookName + `"`)
				if err != nil {
					return err
				}
				newDirName := data.DataPath + "/pic/" + bookDir.Name() + "_" + utils.CleanFileName(bookName)
				seelog.Debugf("重命名文件夹" + newDirName)
				err2 := os.Rename(data.DataPath+"/pic/"+bookDir.Name(), newDirName)
				if err2 != nil {
					return err2
				}
				seelog.Debugf("重命名文件夹成功")
				break
			}
		}
	}
	return nil
}

func doPackAndRename() error {
	dirs, err := ioutil.ReadDir(data.TempPath + "/pic")
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		bookDataPath := data.DataPath + "/pic/" + dir.Name()
		seelog.Debugf("开始处理图片：" + bookDataPath)
		if !dir.IsDir() {
			seelog.Debugf("清除异常文件：" + dir.Name())
			err = os.Remove(bookDataPath)
			if err != nil {
				return err
			}
			continue
		}
		_, err := strconv.Atoi(dir.Name())
		if err != nil {
			seelog.Debugf("清除异常文件夹：" + dir.Name())
			err = os.RemoveAll(bookDataPath)
			if err != nil {
				return err
			}
			continue
		}
		picTempPath := data.TempPath + "/pic/" + dir.Name()
		subdirs, err := ioutil.ReadDir(picTempPath)
		if err != nil {
			return err
		}
		for _, subdir := range subdirs {
			chapterTempPath := picTempPath + "/" + subdir.Name()
			seelog.Debugf("进入图片文件夹：" + chapterTempPath)
			chapterNo, err := strconv.Atoi(subdir.Name())
			if err != nil {
				return err
			}
			//picChapter := data.NewData.Pics[chapterNo]
			_, dirExist := os.Stat(bookDataPath)
			if dirExist == nil {
				picPacks, err := ioutil.ReadDir(bookDataPath)
				if err != nil {
					return err
				}
				for _, picPack := range picPacks {
					if strings.HasPrefix(picPack.Name(), strconv.Itoa(chapterNo)+"_") {
						seelog.Debug("已存在图片包：" + picPack.Name())
						reader, err := zip.OpenReader(bookDataPath + "/" + picPack.Name())
						if err != nil {
							return err
						}
						for _, picFile := range reader.File {
							_, picExist := os.Stat(chapterTempPath + "/" + picFile.FileHeader.Name)
							if picExist != nil {
								f, err := os.OpenFile(chapterTempPath+"/"+picFile.FileHeader.Name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, picFile.Mode())
								defer f.Close()
								if err != nil {
									return err
								}
								fileReader, err := picFile.Open()
								if err != nil {
									return err
								}
								defer fileReader.Close()
								_, err = io.Copy(f, fileReader)
								if err != nil {
									return err
								}
								seelog.Debug("解压图片：" + picFile.FileHeader.Name)
								continue
							}
							seelog.Debug("跳过解压图片：" + picFile.FileHeader.Name)
						}
						reader.Close()
						err = os.Remove(bookDataPath + "/" + picPack.Name())
						if err != nil {
							return err
						}
						seelog.Debug("删除旧压缩包：" + bookDataPath + "/" + picPack.Name())
						break
					}
				}
			}
			volumeName := ""
			chapterName := ""
			find := false
			for _, book := range data.NewData.Books {
				for _, volume := range book.Volumes {
					volumeName, err = strconv.Unquote(`"` + volume.VolumeName + `"`)
					if err != nil {
						return err
					}
					for _, chapter := range volume.Chapters {
						if chapter.ChapterNo == chapterNo {
							var err error
							chapterName, err = strconv.Unquote(`"` + chapter.ChapterName + `"`)
							if err != nil {
								return err
							}
							find = true
							break
						}
					}
					if find {
						break
					}
				}
				if find {
					break
				}
			}
			err2 := utils.CheckFolderAndMake(bookDataPath)
			if err2 != nil {
				return err2
			}
			err = packCBZ(chapterNo, volumeName, chapterName, bookDataPath, chapterTempPath)
			if err != nil {
				return err
			}
			err = os.RemoveAll(chapterTempPath)
			if err != nil {
				return err
			}
			seelog.Debug("删除图片临时文件：" + chapterTempPath)
		}
	}
	return nil
}

func packCBZ(chapterNo int, volumeName, chapterName, bookDataPath, chapterTempPath string) error {
	cbzName := strconv.Itoa(chapterNo) + "_" + volumeName + "_" + chapterName + ".cbz"
	cbzName = utils.CleanFileName(cbzName)
	cbzFile, err := os.Create(bookDataPath + "/" + cbzName)
	defer cbzFile.Close()
	seelog.Debug("建立压缩包：" + bookDataPath + "/" + cbzName)
	writer := zip.NewWriter(cbzFile)
	defer writer.Close()
	picDir, err := ioutil.ReadDir(chapterTempPath)
	if err != nil {
		return err
	}
	for _, picFile := range picDir {
		seelog.Debug("归档图片：" + chapterTempPath + "/" + picFile.Name())
		fw, err := writer.Create(picFile.Name())
		if err != nil {
			return err
		}
		in, err := ioutil.ReadFile(chapterTempPath + "/" + picFile.Name())
		if err != nil {
			return err
		}
		fw.Write(in)
		seelog.Info("归档图片成功："+cbzName)
	}
	return nil
}
