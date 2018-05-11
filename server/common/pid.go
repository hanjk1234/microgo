package common

import (
	"github.com/seefan/to"
	"io/ioutil"
	"os"
	"strconv"
)

func SavePid(file string) error {
	return ioutil.WriteFile(file, to.Bytes(os.Getpid()), 0764)
}
func GetPid(file string) (int, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(bs))
}
