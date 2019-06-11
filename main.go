package main

import (
	"log"
	"runtime"
	"strings"
	"sync"

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/service"
)

func main() {
	imgsChannel := make(chan *config.Cell)
	imgsWillDelete := make(chan *config.Cell)
	var wg sync.WaitGroup

	db := service.DbConnect()

	count := service.GetImageLen(true)

	length := service.GetImageLen(false)
	wg.Add(count + length)
	bar := pb.StartNew(length + count)

	token := service.SetupQiniu()
	uploader := service.UploaderGet()
	bm := service.GetBucketManager()

	go service.GetImages(imgsChannel, count, true)
	go service.GetImages(imgsWillDelete, length, false)

	goroutineCount := runtime.NumCPU()
	for i := 0; i < goroutineCount; i++ {
		go func(index int) {
			for {
				select {
				case item := <-imgsChannel:
					if item != nil && !strings.HasPrefix(item.Src, "qn://") {
						filename, ok := service.UploadToQiniu(uploader, item, token)
						if ok {
							item.Src = "qn://" + filename
							if service.UpdateImage(item) {
								// log.Println("--- SUCCESS SAVED A FILE ---")
							} else {
								//  已存在，不用删除文件，但是要删掉数据库的文件
								// log.Println("--- Already have the file ---")
								service.DeleteRecord(item)
							}
						} else {
							// 不存在，但是 图片没了，还是要删掉数据库文件
							// log.Println("--- image has gone ---")
							service.DeleteRecord(item)
						}
						bar.Increment()
						wg.Done()
					}

				case item := <-imgsWillDelete:
					if item != nil && strings.HasPrefix(item.Src, "qn://") {
						filename := config.RevertFilename(item.Src)
						log.Println(filename)
						if err := bm.Delete(config.GetConfig().Bucket, filename); err != nil && err.Error() != "no such file or directory" {
							log.Println("bm delete error, and the error is: ", err)
						} else {
							service.DeleteRecordSoft(item)
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
