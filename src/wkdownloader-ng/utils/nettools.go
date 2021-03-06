package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

var (
	defaultHTTPClient *HTTPClient
	proxyHTTPClient   *HTTPClient
	testHTTPClient    *HTTPClient
)

func GetFileSize(url string) int64 {
	res, err := GetDefaultHTTPClient().Client.Get(url)
	if err != nil {
		return 0
	}
	sizeStr := res.Header.Get("Content-Length")
	size, err2 := strconv.ParseInt(sizeStr, 0, 64)
	if err2 != nil {
		return 0
	}
	return size
}

func Download(url string, filePath string) error {
	tmpFilePath := fmt.Sprintf("%s.tmp", filePath)
	downloadErr := downloadFile(url, tmpFilePath, false)
	if downloadErr != nil {
		return downloadErr
	}
	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		err2 := os.Remove(filePath)
		if err2 != nil {
			return err2
		}
	}
	err3 := os.Rename(tmpFilePath, filePath)
	if err3 != nil {
		return err3
	}
	return nil
}

func DownloadWithProxy(url string, filePath string) error {
        tmpFilePath := fmt.Sprintf("%s.tmp", filePath)
        downloadErr := downloadFile(url, tmpFilePath, true)
        if downloadErr != nil {
                return downloadErr
        }
        _, err := os.Stat(filePath)
        if !os.IsNotExist(err) {
                err2 := os.Remove(filePath)
                if err2 != nil {
                        return err2
                }
        }
        err3 := os.Rename(tmpFilePath, filePath)
        if err3 != nil {
                return err3
        }
        return nil
}

func downloadFile(url string, filePath string, withProxy bool) error {
	err := CheckFolderAndMake(path.Dir(filePath))
	if err != nil {
		return err
	}

	file, err2 := os.Create(filePath)
	if err2 != nil {
		return err2
	}

	var res *http.Response
	var err3 error
	if withProxy {
		res, err3 = GetProxyHTTPClient().Client.Get(url)
	} else {
		res, err3 = GetDefaultHTTPClient().Client.Get(url)
	}
	if err3 != nil {
		return err3
	}
	defer res.Body.Close()

	_, err4 := io.Copy(file, res.Body)
	if err4 != nil {
		return err4
	}
	file.Close()

	check, err5 := os.Open(filePath)
	defer check.Close()
	if err5 != nil {
		return err5
	}

	stat, err6 := check.Stat()
	if err6 != nil || stat.Size() == 0 {
		return errors.New("文件为空")
	}
	return nil
}

type HTTPClient struct {
	Client http.Client
}

func newDefaultHTTPClient() {
	defaultHTTPClient = &HTTPClient{}
	defaultHTTPClient.Client = http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(600 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 5,
		},
	}
}

func GetDefaultHTTPClient() *HTTPClient {
	if defaultHTTPClient == nil {
		newDefaultHTTPClient()
	}
	return defaultHTTPClient
}

func newProxyHTTPClient() {
	urlParser := url.URL{}
	proxyUrl, _ := urlParser.Parse("http://127.0.0.1:8118")
	proxyHTTPClient = &HTTPClient{}
	proxyHTTPClient.Client = http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(600 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 5,
		},
	}
}

func GetProxyHTTPClient() *HTTPClient {
	if proxyHTTPClient == nil {
		newProxyHTTPClient()
	}
	return proxyHTTPClient
}

func newTestHTTPClient() {
	testHTTPClient = &HTTPClient{}
	testHTTPClient.Client = http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(10 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second * 5,
			DisableKeepAlives:     true,
		},
	}
}

func GetTestHTTPClient() *HTTPClient {
	if testHTTPClient == nil {
		newTestHTTPClient()
	}
	return testHTTPClient
}
