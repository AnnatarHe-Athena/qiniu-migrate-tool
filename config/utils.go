package config

import (
	"fmt"
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
	strs := strings.Split(origin, "/")
	realName := strs[len(strs)-1]
	filename = "athena/zhihu/" + realName
	return
}

func RevertFilename(filename string) (origin string) {
	origin = strings.Replace(filename, "qn://", "", 0)
	return
}
