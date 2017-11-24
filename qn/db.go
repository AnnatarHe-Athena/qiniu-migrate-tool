package qn

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/douban-girls/qiniu-migrate/config"
	_ "github.com/lib/pq"
)

func DbConnect() *sql.DB {
	cfg := config.GetConfig()
	dbPath := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Username, cfg.Pwd, cfg.Dbname)
	db, err := sql.Open("postgres", dbPath)
	config.ErrorHandle(err)
	return db
}

func GetImageLen(db *sql.DB) (count int) {
	result, err := db.Query("SELECT count(id) FROM cells WHERE cate=176 OR cate=179")
	defer result.Close()
	config.ErrorHandle(err)
	for result.Next() {
		if err := result.Scan(&count); err != nil {
			config.ErrorHandle(err)
		}
	}
	return
}

func GetImages(db *sql.DB, result chan *config.Cell, count int) {
	times := int(math.Ceil(float64(count) / 1000))
	stmt, err := db.Prepare("SELECT id, img FROM cells WHERE cate=176 OR cate=179 ORDER BY id ASC LIMIT 1000 OFFSET $1")
	config.ErrorHandle(err)
	for i := 0; i < times; i++ {
		rows, err := stmt.Query(1000 * i)
		config.ErrorHandle(err)
		defer rows.Close()

		if err := rows.Err(); err != nil {
			config.ErrorHandle(err)
		}
		for rows.Next() {
			var id int
			var src string
			if err := rows.Scan(&id, &src); err != nil {
				config.ErrorHandle(err)
			}
			cell := &config.Cell{
				ID:  id,
				Src: src,
			}
			result <- cell
		}
	}
	close(result)
}

func UpdateImage(db *sql.DB, cell *config.Cell) bool {
	_, err := db.Query("UPDATE cells SET img=$1 WHERE id=$2", cell.Src, cell.ID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
