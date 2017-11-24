package main

import (
	"fmt"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/qn"
)

func main() {
	// db := qn.DbConnect()
	// defer db.Close()

	// imgs := qn.GetImages(db)
	imgs := []*config.Cell{&config.Cell{
		ID:  11,
		Src: "https://pic3.zhimg.com/50/v2-47409bff0e05d0e614d898eeec7d310a_hd.jpg",
	}}
	token := qn.SetupQiniu()
	uploader := qn.UploaderGet()

	for _, img := range imgs {
		filename := qn.UploadToQiniu(uploader, img, token)
		img.Src = "qn://" + filename
		fmt.Println(img.Src)
		// if qn.UpdateImage(db, img) {
		// 	fmt.Println("--- SUCCESS SAVED A FILE ---")
		// }
	}
}
