package main

import (
	"fmt"
	"sync"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/qn"
)

const goroutineCount = 16

func main() {
	var wg sync.WaitGroup
	imgsChannel := make(chan *config.Cell, goroutineCount)

	db := qn.DbConnect()
	defer db.Close()

	count := qn.GetImageLen(db)
	wg.Add(count)

	fmt.Println("count: ", count)

	qn.GetImages(db, imgsChannel, count)

	token := qn.SetupQiniu()
	uploader := qn.UploaderGet()

	for i := 0; i < goroutineCount; i++ {
		go func(ch chan *config.Cell) {
			item := <-ch
			fmt.Println(item)
			fmt.Println(item.ID, item.Src)
			filename := qn.UploadToQiniu(uploader, item, token)
			item.Src = "qn://" + filename
			if qn.UpdateImage(db, item) {
				fmt.Println("--- SUCCESS SAVED A FILE ---")
			}
			wg.Done()
		}(imgsChannel)
	}
	wg.Wait()
	println("job done")
}
