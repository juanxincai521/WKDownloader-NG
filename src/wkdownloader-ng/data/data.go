package data

const (
	// ConfPath 数据文件的位置
	ConfPath = "../conf"
	// DataPath 数据文件的位置
	DataPath = "../data"
	// TempPath 临时文件的位置
	TempPath = "../temp"
)

var (
	BookList []int
	OldData  *WKData
	NewData  *WKData
)

type WKData struct {
	Books             []*Book       `json:"books"`
	Pics              map[int]*Pic  `json:"pics"`
	CopyrightChapters map[int]*Page `json:"copyright"`
}

type Book struct {
	BookNo   int       `json:"book_no"`
	BookName string    `json:"book_name"`
	Volumes  []*Volume `json:"volumes"`
}

type Volume struct {
	VolumeName string     `json:"volume_name"`
	Chapters   []*Chapter `json:"chapters"`
}

type Chapter struct {
	ChapterNo   int    `json:"chapter_no"`
	ChapterName string `json:"chapter_name"`
}

type Pic struct {
	BookNo    int    `json:"book_no"`
	ChapterNo int    `json:"chapter_no"`
	PicNo     []*Img `json:"pic_no"`
}

type Img struct {
	ImgNo  int    `json:"img_no"`
	Extend string `json:"extend"`
}

type Page struct {
	BookNo      int
	BookName    string
	ChapterNo   int
	ChapterName string
}

func BVCToPage(books []*Book) map[int]*Page {
	pages := make(map[int]*Page)
	for _, book := range books {
		for _, volume := range book.Volumes {
			for _, chapter := range volume.Chapters {
				page := &Page{}
				page.BookNo = book.BookNo
				page.BookName = book.BookName
				page.ChapterNo = chapter.ChapterNo
				page.ChapterName = chapter.ChapterName
				pages[chapter.ChapterNo] = page
			}
		}
	}
	return pages
}
