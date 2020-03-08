package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	// "runtime"
	"strings"
	"sync"

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/douban-girls/qiniu-migrate/config"
	"github.com/douban-girls/qiniu-migrate/service"
)

func doMigrate() {
	imgsChannel := make(chan *config.Cell)
	imgsWillDelete := make(chan *config.Cell)
	var wg sync.WaitGroup

	count := service.GetImageLen(true)
	length := service.GetImageLen(false)
	wg.Add(count + length)
	bar := pb.StartNew(length + count)

	qiniuService := service.NewQiniuService()

	go service.GetImages(imgsChannel, count, true)
	go service.GetImages(imgsWillDelete, length, false)

	for i := 0; i < 10; i++ {
		go func(index int) {
			for {
				select {
				case item := <-imgsChannel:
					if item == nil || strings.HasPrefix(item.Src, "qn://") {
						continue
					}

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
						log.Println(err)
						// 这里有时候会报错，暂时不 panic 了
						// panic(err)
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

}

func testFaceDetch() {
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

func getAction() string {
	action := flag.String("action", "", "migrate weibo images to qiniu")
	flag.Parse()

	if *action == "" {
		panic("action should be `migrate`, `instrgam` or `printNot1`")
	}
	return *action
}

func main() {
	action := getAction()
	db := service.DbConnect()
	defer db.Close()

	if action == "migrate" {
		// doMigrate()
		return
	}

	if action == "tag:migrate" {
		service.MigrateTagsFromCategories()
		return
	}

	panic("not support this type yet: " + action)

	// doMigrate()
	// insertIgData()
	// printAllNot1Image()
}

func insertIgData() {
	err := service.IGMain()
	if err != nil {
		log.Println(err)
	}
}

const baseFromQuery = "from cells where createdat < '2019-07-30 00:00:00' AND img like '%qn://%' and premission = 2"

// 临时代码，打印出符合 qiniu shell 规范得文件
func printAllNot1Image() {
	// TODO: 获取所有 qiniu 图片，并将类型置为 1

	db := service.DbConnect()
	f, err := os.OpenFile("local-2.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)

	if err != nil {
		log.Println(err)
		return
	}

	var cursor = 0
	var count int

	db.Get(&count, "SELECT count(*) "+baseFromQuery)

	log.Println("count", count)

	for cursor < count {
		var keys []string
		if err := db.Select(&keys, "SELECT img "+baseFromQuery+" LIMIT 1000 OFFSET $1", cursor); err != nil {
			log.Println(err)
			return
		}

		for _, k := range keys {
			newK := strings.Replace(k, "qn://", "", -1)
			if _, e := f.WriteString(newK + "\t1\n"); e != nil {
				log.Println("write file error", e)
			}
		}

		cursor += 1000
	}

	if err := f.Close(); err != nil {
		log.Println("close", err)
	}

	return
}
