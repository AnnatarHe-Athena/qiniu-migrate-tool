package main

import (
	"fmt"
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
	wg.Add(count)

	go qn.GetImages(db, imgsChannel, count)

	token := qn.SetupQiniu()
	uploader := qn.UploaderGet()

	for i := 0; i < goroutineCount; i++ {
		go func() {
			item := <-imgsChannel
			if item == nil {
				return
			}
			filename := qn.UploadToQiniu(uploader, item, token)
			item.Src = "qn://" + filename
			if qn.UpdateImage(db, item) {
				fmt.Println("--- SUCCESS SAVED A FILE ---")
			}
			wg.Done()
			// done <- true
		}()
	}
	wg.Wait()
	println("job done")
}
