package common

import (
	"os"
	"strings"
)

func Path(path ...string) string {
	return strings.Join(path, string(os.PathSeparator))
}

/**
 * 数组去重 去空
 */
func RemoveDuplicatesAndEmpty(a []string) (ret []string) {
	a_len := len(a)
	for i := 0; i < a_len; i++ {
		if (i > 0 && a[i-1] == a[i]) || len(a[i]) == 0 {
			continue
		}
		ret = append(ret, a[i])
	}
	return
}

func NotExist(path string) bool {
	_, err := os.Stat(path)
	return err != nil && os.IsNotExist(err)
}