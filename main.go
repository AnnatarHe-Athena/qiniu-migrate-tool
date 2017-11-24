package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/qn"
)

const goroutineCount = 16

func main() {
	imgsChannel := make(chan *config.Cell)
	var wg sync.WaitGroup
	// done := make(chan bool)
	// var counter uint32 = 0

	db := qn.DbConnect()
	defer db.Close()

	count := qn.GetImageLen(db)
	println(count)
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
