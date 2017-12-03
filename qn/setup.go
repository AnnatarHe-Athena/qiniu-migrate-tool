package qn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

func SetupQiniu() string {
	config := config.GetConfig()
	putPolicy := storage.PutPolicy{
		Scope: config.Bucket,
	}
	putPolicy.Expires = 3600 * 10
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}

func UploaderGet() *storage.FormUploader {
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuadong
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	return formUploader
}

func UploadToQiniu(
	uploader *storage.FormUploader,
	img *config.Cell,
	token string,
) (filename string, ok bool) {
	content, length, ok := downloadImg(img)
	defer content.Close()
	if !ok {
		return "", ok
	}
	filename = config.GenFilename(img.Src)
	ret := storage.PutExtra{}
	err := uploader.Put(context.Background(), &ret, token, filename, content, length, nil)
	if err != nil {
		fmt.Println("retry to save the images")
		return UploadToQiniu(uploader, img, token)
	}
	return filename, true
}

func DeleteFromQiniu(bucketManager *storage.BucketManager, filename string) {
	bucket := config.GetConfig().Bucket
	err := bucketManager.Delete(bucket, filename)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func GetBucketManager() *storage.BucketManager {
	c := config.GetConfig()
	mac := qbox.NewMac(c.AccessKey, c.SecretKey)
	cfg := storage.Config{UseHTTPS: false}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	return bucketManager
}

func downloadImg(cell *config.Cell) (io.ReadCloser, int64, bool) {
	src := cell.Src
	fmt.Println(src)
	if !strings.HasPrefix(cell.Src, "http") {
		// 微博图片，需要转 url
		src = "http://wx2.sinaimg.cn/large/" + src
	}
	res, e := http.Get(src)
	config.ErrorHandle(e)
	lenStr := res.Header.Get("content-length")
	length, err := strconv.ParseInt(lenStr, 10, 64)
	config.ErrorHandle(err)
	if res.StatusCode == 301 {
		return res.Body, length, false
	}
	return res.Body, length, true
}
