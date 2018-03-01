package config

import (
	"fmt"
	"net/url"
	"strings"
)

func ErrorHandle(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

// GenFilename will generater a filebase base on origin last filename not path
func GenFilename(origin string) (filename string) {
	u, _ := url.Parse(origin)
	strs := strings.Split(u.Path, "/")
	realName := strs[len(strs)-1]
	filename = "athena/zhihu/" + realName
	return
}

func RevertFilename(filename string) (origin string) {
	origin = strings.Replace(filename, "qn://", "", 0)
	return
}
