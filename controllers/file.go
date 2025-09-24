package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

var (
	imgexp = regexp.MustCompile("^image")
)

func FileUpload(w http.ResponseWriter, r *http.Request) {
	// 从参数读取 url
	imgURL := r.URL.Query().Get("url")
	if imgURL == "" {
		http.Error(w, `{"error":"missing url"}`, http.StatusBadRequest)
		return
	}

	// 下载远程图片
	resp, err := http.Get(imgURL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf(`{"error":"failed to download image: %s"}`, resp.Status), http.StatusBadRequest)
		return
	}

	// 创建临时文件
	tempfile, err := ioutil.TempFile("", "ocrserver-")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}
	defer func() {
		tempfile.Close()
		os.Remove(tempfile.Name())
	}()

	// 保存下载的内容到本地文件
	if _, err = io.Copy(tempfile, resp.Body); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	// OCR
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(tempfile.Name())
	client.Languages = []string{"eng"}
	if langs := r.URL.Query().Get("languages"); langs != "" {
		client.Languages = strings.Split(langs, ",")
	}
	if whitelist := r.URL.Query().Get("whitelist"); whitelist != "" {
		client.SetWhitelist(whitelist)
	}

	var out string
	switch r.URL.Query().Get("format") {
	case "hocr":
		out, err = client.HOCRText()
	default:
		out, err = client.Text()
	}
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
		return
	}

	// 返回 JSON
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"result": strings.Trim(out, r.URL.Query().Get("trim")),
	})
}
