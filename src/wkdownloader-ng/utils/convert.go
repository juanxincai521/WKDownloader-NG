package utils

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GBKToUTF8(content []byte) ([]byte, error) {
	result, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(content), simplifiedchinese.GBK.NewDecoder()))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func UTF8ToGBK(content []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(content), simplifiedchinese.GBK.NewEncoder())
	result, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return result, nil
}
