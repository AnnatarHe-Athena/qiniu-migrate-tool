package qn

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math"

	"github.com/douban-girls/qiniu-migrate/config"
	_ "github.com/lib/pq"
)

const baseWhereCond = " FROM cells WHERE img not like '%qn://%' "

func DbConnect() *sql.DB {
	cfg := config.GetConfig()
	dbPath := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Username, cfg.Pwd, cfg.Dbname)
	db, err := sql.Open("postgres", dbPath)
	config.ErrorHandle(err)
	return db
}

func GetImageLen(db *sql.DB) (count int) {
	result, err := db.Query("SELECT count(id)" + baseWhereCond)
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
	stmt, err := db.Prepare("SELECT id, img" + baseWhereCond + "ORDER BY id ASC LIMIT 1000 OFFSET $1")
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
				Md5: md5Hash(src),
			}
			result <- cell
		}
	}
	defer stmt.Close()
	close(result)
}

func md5Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func UpdateImage(db *sql.DB, cell *config.Cell) (rep bool) {
	rows, err := db.Query("UPDATE cells SET img=$1, md5=$2 WHERE id=$3", cell.Src, cell.Md5, cell.ID)
	if err != nil {
		log.Println("update error: ", err, cell.Src)
		log.Println("md5: ", cell.Md5)
		rep = false
		return
	}
	rep = true
	// 手动关闭
	rows.Close()
	return
}

func DeleteRecord(db *sql.DB, cell *config.Cell) bool {
	rows, err := db.Query("DELETE FROM cells WHERE id=$1", cell.ID)
	config.ErrorHandle(err)
	defer rows.Close()
	return true
}
