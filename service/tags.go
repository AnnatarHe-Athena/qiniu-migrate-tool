package service

import (
	"log"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// tags 服务
type Category struct {
	ID        int    `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Src       int    `json:"src" db:"src"`
	Count     int    `json:"count" db:"count"`
	CreatedAt int64  `json:"createdAt" db:"createdat"`
	UpdatedAt int64  `json:"updatedAt" db:"updatedat"`
}

type Tag struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"createdat"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updatedat"`
}

type TagGirlsMapping struct {
	ID     int `json:"id" db:"id"`
	TagID  int `json:"tagId" db:"tag_id"`
	CellID int `json:"cellId" db:"cell_id"`
}

func fetchCategories() (categories []Category, err error) {
	if err = db.Select(&categories, "select id, name, src from categories"); err != nil {
		return
	}
	return categories, nil
}

func fetchTags() (tags []Tag, err error) {
	if err = db.Select(&tags, "SELECT * FROM tags"); err != nil {
		return
	}

	return tags, nil
}

func insertTagsFromCategories() {
	categories, err := fetchCategories()

	if err != nil {
		panic(err)
	}

	for _, category := range categories {
		tag := &Tag{
			Name:        category.Name,
			Description: category.Name,
		}
		_, err := db.NamedExec("insert into tags(name, description) values(:name, :description)", tag)
		if err != nil {
			panic(err)
		}
	}

	log.Println("Done")
}

func MigrateTagsFromCategories() {
	// insertTagsFromCategories()
	err := tagCells()

	if err != nil {
		panic(err)
	}
}

type cellLite struct {
	ID      int    `db:"id"`
	Content string `db:"text"`
	Cate    int    `db:"cate"`
}

func (c cellLite) FindCateNameBy(categories []Category) string {
	for _, category := range categories {
		if category.ID == c.Cate {
			return category.Name
		}
	}

	return ""
}

func (c cellLite) FindTagIdByCateName(cateName string, tags []Tag) int {
	for _, tag := range tags {
		if tag.Name == cateName {
			return tag.ID
		}
	}

	return -1
}

func tagCells() error {
	lastID := 0
	categories, err := fetchCategories()
	if err != nil {
		panic(err)
	}
	tags, err := fetchTags()

	if err != nil {
		panic(err)
	}

	keywordForTagIdDict := map[string]int{
		// hard code
		"腿":    3,
		"胸":    6,
		"臀":    5,
		"屁股":   5,
		"🐻":    6,
		"腰":    5,
		"自拍":   2,
		"plmm": 2,
	}

	var cells []cellLite

	// logrus.Println(categories, tags)

	for lastID != -1 {
		if err := db.Select(&cells, "SELECT id, text, cate from cells where id > $1 ORDER BY id ASC LIMIT 1000", lastID); err != nil {
			return err
		}

		if len(cells) == 0 {
			lastID = -1
		}

		for i, cell := range cells {
			tagID := cell.FindTagIdByCateName(
				cell.FindCateNameBy(categories),
				tags,
			)
			logrus.Println("tagid: ", tagID)

			doTag(cell, tagID, keywordForTagIdDict)

			if i == len(cells)-1 {
				lastID = cells[len(cells)-1].ID
			}
		}

		logrus.Println("next: ", lastID)
	}

	return nil
}

func execInsert(t TagGirlsMapping) error {
	_, err := db.NamedExec("INSERT INTO tags_girls(tag_id, cell_id) VALUES(:tag_id, :cell_id)", &t)

	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "conflict") {
		return nil
	}

	return err
}

func doTag(cell cellLite, belongsToTagID int, dict map[string]int) error {
	// 本身从属的 category 要塞入

	if belongsToTagID != -1 {
		execInsert(TagGirlsMapping{
			TagID:  belongsToTagID,
			CellID: cell.ID,
		})
	}

	for keyword, tagID := range dict {
		if strings.Contains(cell.Content, keyword) {
			execInsert(TagGirlsMapping{
				TagID:  tagID,
				CellID: cell.ID,
			})
		}
	}

	return nil
}
