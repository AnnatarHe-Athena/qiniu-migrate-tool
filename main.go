package main

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

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

	go qn.GetImages(db, imgsChannel, count)

	token := qn.SetupQiniu()
	uploader := qn.UploaderGet()

	for i := 0; i < goroutineCount; i++ {
		go func() {
			for {
				select {
				case item := <-imgsChannel:
					if item != nil && !strings.HasPrefix(item.Src, "qn://") {
						fmt.Println("go item: ", item.Src)
						filename := qn.UploadToQiniu(uploader, item, token)
						item.Src = "qn://" + filename
						if qn.UpdateImage(db, item) {
							fmt.Println("--- SUCCESS SAVED A FILE ---")
						} else {
							//  已存在，不用删除文件，但是要删掉数据库的文件
							fmt.Println("--- Already have the file ---")
							qn.DeleteRecord(db, item)
						}
						wg.Done()
					}
					// done <- true

				}
			}
		}()
	}
	wg.Wait()
	println("job done")
}
