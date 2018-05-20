package main

import (
	"log"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/qn"
)

func main() {
	goroutineCount := runtime.NumCPU()
	imgsChannel := make(chan *config.Cell)
	var wg sync.WaitGroup

	db := qn.DbConnect()
	defer db.Close()

	count := qn.GetImageLen(db)
	wg.Add(count)
	bar := pb.StartNew(count)

	go qn.GetImages(db, imgsChannel, count)

	token := qn.SetupQiniu()
	log.Println("qiniu token: ", token)
	uploader := qn.UploaderGet()

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
						bar.Increment()
						wg.Done()
					}

				}
			}
		}(i)
	}
	wg.Wait()
	log.Println("job done")
}
