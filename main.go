package main

import (
	"io/ioutil"
	"log"

	// "runtime"
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

	qiniuService := service.NewQiniuService()

	go service.GetImages(imgsChannel, count, true)
	go service.GetImages(imgsWillDelete, length, false)

	// goroutineCount := runtime.NumCPU()
	for i := 0; i < 10; i++ {
		go func(index int) {
			for {
				select {
				case item := <-imgsChannel:
					if item == nil || strings.HasPrefix(item.Src, "qn://") {
						continue
					}

					// imageReader, ok := service.DownloadImage(item)
					// if !ok {
					// 	log.Println("download image error: ", item.Src)
					// }

					// imageByte, _ := ioutil.ReadAll(imageReader)

					// faceDetectionService := service.TencentFaceDetectionService{
					// 	AppID:  config.TencentAIAppID,
					// 	AppKey: config.TencentAIAppKey,
					// 	Image:  imageByte,
					// }

					// response, err := faceDetectionService.Request()
					// if err != nil {
					// 	log.Println("face", err, item.Src)
					// }

					// if len(response.FaceList) == 0 {
					// 	// TODO: update db data
					// 	continue
					// }

					// if !faceDetectionService.IsValid(response.FaceList[0]) {
					// 	// TODO: update db data
					// 	continue
					// }

					// filename, ok := qiniuService.Upload(imageReader, length, item.Src)
					filename, err := qiniuService.UploadByFetch(item.Src, item.Src)
					if err == nil {
						item.Src = "qn://" + filename
						if service.UpdateImage(item) {
							// log.Println("--- SUCCESS SAVED A FILE ---")
						} else {
							//  已存在，不用删除文件，但是要删掉数据库的文件
							// log.Println("--- Already have the file ---")
							service.DeleteRecord(item)
						}
					}

					if err != nil && !strings.Contains(err.Error(), "EOF") {
						log.Panic(err, "unexpected error")
					}
					// 不存在，但是 图片没了，还是要删掉数据库文件
					// log.Println("--- image has gone ---")
					// service.DeleteRecord(item)

					bar.Increment()
					// imageReader.Close()
					wg.Done()
				case item := <-imgsWillDelete:
					if item != nil && strings.HasPrefix(item.Src, "qn://") {
						filename := config.RevertFilename(item.Src)
						log.Println(filename)
						if err := qiniuService.Delete(filename); err != nil && err.Error() != "no such file or directory" {
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

func main1() {
	item := &config.Cell{
		Src: "https://wx3.sinaimg.cn/large/8112eefdly1fonbw696egj20lc0sgx0h.jpg",
	}
	imageReader, ok := service.DownloadImage(item)
	if !ok {
		log.Println("download image error: ", item.Src)
	}

	imageByte, _ := ioutil.ReadAll(imageReader)

	faceDetectionService := service.TencentFaceDetectionService{
		AppID:  config.TencentAIAppID,
		AppKey: config.TencentAIAppKey,
		Image:  imageByte,
	}

	// TODO: is vaild
	response, err := faceDetectionService.Request()
	firstFace := response.FaceList[0]

	log.Println(firstFace.Beauty, firstFace.Gender)

	log.Println("face detection service: ", response, err)
}
