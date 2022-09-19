package test

import (
	"simpleGinIm/helper"
	"testing"
)

func TestHttpPost(t *testing.T) {
	err := helper.SendPost("http://www.baidu.com", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
}
