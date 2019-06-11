package service

import (
	"context"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

func SetupQiniu() string {
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
	err := uploader.Put(context.Background(), &ret, token, filename, content, length, &storage.PutExtra{})
	if err != nil {
		if err == io.EOF {
			log.Panic("error when post the image to qiniu server")
		}
		if err.Error() == "file exists" {
			log.Println(err)
			return filename, true
		}
		log.Println("retry to save the images", err)
		return UploadToQiniu(uploader, img, token)
	}
	return filename, true
}

func DeleteFromQiniu(bucketManager *storage.BucketManager, filename string) {
	bucket := config.Bucket
	err := bucketManager.Delete(bucket, filename)
	if err != nil {
		log.Println(err)
		return
	}
}

func GetBucketManager() *storage.BucketManager {
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	cfg := storage.Config{UseHTTPS: false}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	return bucketManager
}

func downloadImg(cell *config.Cell) (io.ReadCloser, int64, bool) {
	src := cell.Src
	// fmt.Println(src)
	if !strings.HasPrefix(cell.Src, "http") {
		// 微博图片，需要转 url
		src = "http://ww2.sinaimg.cn/large/" + src
	}
	res, e := http.Get(src)
	if e != nil {
		log.Println(cell.ID, src)
		errorChecker(e)
	}
	lenStr := res.Header.Get("content-length")
	length, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		log.Println(cell.Src, cell.ID, err)
		errorChecker(err)
	}
	return res.Body, length, true
}
