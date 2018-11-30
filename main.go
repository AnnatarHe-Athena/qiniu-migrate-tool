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
	imgsChannel := make(chan *config.Cell)
	imgsWillDelete := make(chan *config.Cell)
	var wg sync.WaitGroup

	db := qn.DbConnect()

	count := qn.GetImageLen(true)

	length := qn.GetImageLen(false)
	wg.Add(count + length)
	bar := pb.StartNew(length + count)

	token := qn.SetupQiniu()
	uploader := qn.UploaderGet()
	bm := qn.GetBucketManager()

	go qn.GetImages(imgsChannel, count, true)
	go qn.GetImages(imgsWillDelete, length, false)

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
							if qn.UpdateImage(item) {
								// log.Println("--- SUCCESS SAVED A FILE ---")
							} else {
								//  已存在，不用删除文件，但是要删掉数据库的文件
								// log.Println("--- Already have the file ---")
								qn.DeleteRecord(item)
							}
						} else {
							// 不存在，但是 图片没了，还是要删掉数据库文件
							// log.Println("--- image has gone ---")
							qn.DeleteRecord(item)
						}
						bar.Increment()
						wg.Done()
					}

				case item := <-imgsWillDelete:
					if item != nil && strings.HasPrefix(item.Src, "qn://") {
						filename := config.RevertFilename(item.Src)
						log.Println(filename)
						if err := bm.Delete(config.GetConfig().Bucket, filename); err != nil {
							log.Println(err)
							qn.DeleteRecordSoft(item)
						}
						// 暂不删除，先测测再说
						bar.Increment()
						wg.Done()
					}
				}
			}
		}(i)
	}
	wg.Wait()
	log.Println("job done")
	bar.Finish()

	defer db.Close()
}
