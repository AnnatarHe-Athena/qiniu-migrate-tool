package service

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

type qiniuService struct {
	Bucket        string
	AccessKey     string
	SecretKey     string
	token         string
	uploader      *storage.FormUploader
	bucketManager *storage.BucketManager
}

type QiniuService interface {
	Upload(content io.ReadCloser, length int64, originFileName string) (filename string, ok bool)
	UploadByFetch(src, originFileName string) (filename string, err error)
	Delete(filename string) error
}

func initQiniuToken() string {
	putPolicy := storage.PutPolicy{
		Scope: config.Bucket,
	}
	putPolicy.Expires = 3600 * 10
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}
func initBucketManager() *storage.BucketManager {
	mac := qbox.NewMac(config.AccessKey, config.SecretKey)
	cfg := storage.Config{UseHTTPS: false}
	bucketManager := storage.NewBucketManager(mac, &cfg)
	return bucketManager
}

func NewQiniuService() QiniuService {
	return qiniuService{
		Bucket:        config.Bucket,
		AccessKey:     config.AccessKey,
		SecretKey:     config.SecretKey,
		token:         initQiniuToken(),
		uploader:      initUploader(),
		bucketManager: initBucketManager(),
	}
}

func initUploader() *storage.FormUploader {
	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuadong
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	return formUploader
}

func (s qiniuService) Upload(content io.ReadCloser, length int64, originFileName string) (filename string, ok bool) {
	filename = config.GenFilename(originFileName)
	ret := storage.PutRet{}
	extra := storage.PutExtra{}

	err := s.uploader.Put(context.Background(), &ret, s.token, filename, content, length, &extra)

	if err != nil {
		if err == io.EOF {
			log.Panic("error when post the image to qiniu server")
		}
		if err.Error() == "file exists" {
			log.Println(err)
			return filename, true
		}
		log.Println("retry to save the images", err)
		return s.Upload(content, length, originFileName)
	}
	return filename, true
}

func (s qiniuService) UploadByFetch(src, originFileName string) (filename string, err error) {
	filename = config.GenFilename(originFileName)
	requestUrl := src
	if !strings.HasPrefix(requestUrl, "http") {
		// 微博图片，需要转 url
		requestUrl = "http://ww2.sinaimg.cn/large/" + src
	}

	response, err := s.bucketManager.Fetch(requestUrl, s.Bucket, filename)
	if err != nil {
		panic(err)
	}
	return response.Key, err
}

func (s qiniuService) Delete(filename string) error {
	return s.bucketManager.Delete(s.Bucket, filename)
}

func DownloadImage(cell *config.Cell) (io.ReadCloser, bool) {
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
	return res.Body, true
}
