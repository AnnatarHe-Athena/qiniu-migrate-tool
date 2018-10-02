package main

import (
	"database/sql"
	"log"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/qn"
	"github.com/qiniu/api.v7/storage"
)

func main() {
	imgsChannel := make(chan *config.Cell)
	var wg sync.WaitGroup

	db := qn.DbConnect()
	defer db.Close()

	count := qn.GetImageLen(db, true)
	wg.Add(count)
	bar := pb.StartNew(count)

	// go qn.GetImages(db, imgsChannel, count, true)

	token := qn.SetupQiniu()
	log.Println("qiniu token: ", token)
	// uploader := qn.UploaderGet()

	// 先迁移
	// migrateToQiniu(imgsChannel, uploader, db, token, func() interface{} {
	// 	bar.Increment()
	// 	wg.Done()
	// 	return nil
	// })
	close(imgsChannel)
	bar.Finish()

	// wg.Wait()
	log.Println("job done")

	// 后删库
	deleteImages(db)
}

// deleteImages will delete qiniu image and delete record in database
func deleteImages(db *sql.DB) {
	bm := qn.GetBucketManager()
	length := qn.GetImageLen(db, false)
	bar := pb.StartNew(length)
	var wg sync.WaitGroup
	wg.Add(length)
	imgsWillDelete := make(chan *config.Cell)
	go qn.GetImages(db, imgsWillDelete, length, false)

	go func() {
		for {
			select {
			case item := <-imgsWillDelete:
				if item != nil && !strings.HasPrefix(item.Src, "qn://") {
					filename := config.RevertFilename(item.Src)
					log.Println(filename)
					if err := bm.Delete(config.GetConfig().Bucket, filename); err != nil {
						log.Println(err)
					}
				}
				bar.Increment()
				wg.Done()
			}
		}
	}()
	wg.Wait()
}

func migrateToQiniu(
	imgsChannel chan *config.Cell,
	uploader *storage.FormUploader,
	db *sql.DB,
	token string,
	onOneEnd func() interface{},
) {
	goroutineCount := runtime.NumCPU()
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			for {
				select {
				case item := <-imgsChannel:
					if item != nil && !strings.HasPrefix(item.Src, "qn://") {
						filename, ok := qn.UploadToQiniu(uploader, item, token)
						if ok {
							item.Src = "qn://" + filename
							if qn.UpdateImage(db, item) {
								// log.Println("--- SUCCESS SAVED A FILE ---")
							} else {
								//  已存在，不用删除文件，但是要删掉数据库的文件
								// log.Println("--- Already have the file ---")
								qn.DeleteRecord(db, item)
							}
						} else {
							// 不存在，但是 图片没了，还是要删掉数据库文件
							// log.Println("--- image has gone ---")
							qn.DeleteRecord(db, item)
						}
					}
					onOneEnd()
				}
			}
		}(i)
	}
}
