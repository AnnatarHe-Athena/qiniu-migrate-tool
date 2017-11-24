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

func GenFilename(origin string) (filename string) {
	strs := strings.Split(origin, "/")
	realName := strs[len(strs)-1]
	filename = "athena/zhihu/" + realName
	return
}
