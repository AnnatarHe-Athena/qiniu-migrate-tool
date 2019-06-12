package config

import (
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ErrorHandle(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GenFilename will generater a filebase base on origin last filename not path
func GenFilename(origin string) (filename string) {
	if strings.Contains(origin, "qq.com") {
		return "athena/qq/" + randStringRunes(35) + "-" + strconv.Itoa(int(time.Now().UnixNano())) + ".jpg"
	}
	u, err := url.Parse(origin)
	if err != nil {
		panic(err)
	}
	strs := strings.Split(u.Path, "/")
	realName := strs[len(strs)-1]

	if len(realName) <= 1 {
		return "athena/misc/" + randStringRunes(35) + "-" + strconv.Itoa(int(time.Now().UnixNano())) + ".jpg"
	}

	if strings.Contains(origin, "ruguoapp.com") {
		return "athena/jike/" + realName
	}
	filename = "athena/zhihu/" + realName
	return
}

func RevertFilename(filename string) (origin string) {
	origin = strings.Replace(filename, "qn://", "", -1)
	return
}
