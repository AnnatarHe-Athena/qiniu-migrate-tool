package qn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

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
) (filename string) {
	content, length := downloadImg(img)
	defer content.Close()
	filename = config.GenFilename(img.Src)
	ret := storage.PutExtra{}
	err := uploader.Put(context.Background(), &ret, token, filename, content, length, nil)
	if err != nil {
		fmt.Println("retry to save the images")
		return UploadToQiniu(uploader, img, token)
	}
	return
}

func downloadImg(cell *config.Cell) (io.ReadCloser, int64) {
	fmt.Println(cell.Src)
	res, e := http.Get(cell.Src)
	config.ErrorHandle(e)
	lenStr := res.Header.Get("content-length")
	length, err := strconv.ParseInt(lenStr, 10, 64)
	config.ErrorHandle(err)
	return res.Body, length
}
