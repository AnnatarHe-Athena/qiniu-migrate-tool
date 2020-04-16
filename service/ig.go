package service

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/qiniu/api.v7/storage"
	"github.com/sirupsen/logrus"
)

type igImage struct {
	Typename         string `json:"__typename"`
	CommentsDisabled bool   `json:"comments_disabled"`
	Dimensions       struct {
		Height int `json:"height"`
		Width  int `json:"width"`
	} `json:"dimensions"`
	DisplayURL           string `json:"display_url"`
	EdgeMediaPreviewLike struct {
		Count int `json:"count"`
	} `json:"edge_media_preview_like"`
	EdgeMediaToCaption struct {
		Edges []struct {
			Node struct {
				Text string `json:"text"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"edge_media_to_caption"`
	EdgeMediaToComment struct {
		Count int `json:"count"`
	} `json:"edge_media_to_comment"`
	GatingInfo   interface{} `json:"gating_info"`
	ID           string      `json:"id"`
	IsVideo      bool        `json:"is_video"`
	MediaPreview string      `json:"media_preview"`
	Owner        struct {
		ID string `json:"id"`
	} `json:"owner"`
	Shortcode          string   `json:"shortcode"`
	Tags               []string `json:"tags"`
	TakenAtTimestamp   int      `json:"taken_at_timestamp"`
	ThumbnailResources []struct {
		ConfigHeight int    `json:"config_height"`
		ConfigWidth  int    `json:"config_width"`
		Src          string `json:"src"`
	} `json:"thumbnail_resources"`
	ThumbnailSrc string   `json:"thumbnail_src"`
	Urls         []string `json:"urls"`
	Username     string   `json:"username"`
}

type igProfile struct {
	GraphImages      []igImage `json:"GraphImages"`
	GraphProfileInfo struct {
		CreatedTime int `json:"created_time"`
		Info        struct {
			Biography         string `json:"biography"`
			FollowersCount    int    `json:"followers_count"`
			FollowingCount    int    `json:"following_count"`
			FullName          string `json:"full_name"`
			ID                string `json:"id"`
			IsBusinessAccount bool   `json:"is_business_account"`
			IsJoinedRecently  bool   `json:"is_joined_recently"`
			IsPrivate         bool   `json:"is_private"`
			PostsCount        int    `json:"posts_count"`
			ProfilePicURL     string `json:"profile_pic_url"`
		} `json:"info"`
		Username string `json:"username"`
	} `json:"GraphProfileInfo"`
}

const igHost = "https://www.instagram.com/"

func IGMain(dir string) error {
	log.Println("gogogo", dir)
	dirs, err := ioutil.ReadDir(dir)

	DbConnect()
	qiniuService := NewQiniuService()

	if err != nil {
		return err
	}

	log.Println("all: ", len(dirs))

	for idx, v := range dirs {
		if !v.IsDir() {
			continue
		}

		log.Println("current: ", idx, v.Name())

		cells := fetchCellsFromDir(v.Name(), dir)

		for _, c := range cells {
			if err := uploadIgFileToQiniu(qiniuService, c); err != nil {
				logrus.Errorln(err)
				return err
			}

			c.imageKeyInQiniu = "qn://" + c.imageKeyInQiniu
			if err := c.Save(); err != nil {
				if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					logrus.Errorln(err)
				}
				continue
				// return err
			}
		}
	}

	return nil
}

func uploadIgFileToQiniu(service QiniuService, item CellItem) error {
	f, err := os.OpenFile(item.Img, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	info, err := f.Stat()

	if err != nil {
		return err
	}

	ret := storage.PutRet{}

	err = service.UploadByReader(&ret, item.imageKeyInQiniu, f, info.Size())

	if err != nil {
		if err == io.EOF {
			logrus.Panic("error when post the image to qiniu server")
		}
		if err.Error() == "file exists" {
			logrus.Println(err)
			return nil
		}
		logrus.Println("retry to save the images", err)
		return uploadIgFileToQiniu(service, item)
	}
	return nil
}

func fetchCellsFromDir(dirname string, baseDir string) (results []CellItem) {
	metaData, _ := ioutil.ReadFile(baseDir + "/" + dirname + "/" + dirname + ".json")

	var meta igProfile

	err := json.Unmarshal(metaData, &meta)

	if err != nil {
		panic(err)
	}

	for _, i := range meta.GraphImages {
		if i.IsVideo {
			continue
		}
		for _, u := range i.Urls {

			uurl, _ := url.Parse(u)
			paths := strings.Split(uurl.Path, "/")
			img := paths[len(paths)-1]
			if !strings.HasSuffix(img, ".jpg") {
				continue
			}
			_, err := os.Stat(baseDir + "/" + dirname + "/" + img)
			if os.IsNotExist(err) {
				logrus.Errorln(baseDir+"/"+dirname+"/"+img, err)
				continue
			}
			imageEdges := i.EdgeMediaToCaption.Edges

			imageText := i.Username

			if len(imageEdges) > 0 {
				imageText = imageEdges[0].Node.Text
			}

			item := CellItem{
				FromID:          igHost + i.Username,
				FromURL:         igHost + "p/" + i.Shortcode,
				Text:            imageText,
				Img:             baseDir + "/" + dirname + "/" + img,
				imageKeyInQiniu: "athena/instagram/" + img,
				Cate:            247,
			}

			results = append(results, item)
		}
	}

	return results
}
