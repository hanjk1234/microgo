package common

import (
	"github.com/seefan/to"
	"io/ioutil"
	"os"
)

func SavePid(file string) error {
	return ioutil.WriteFile(file, to.Bytes(os.Getpid()), 0764)
}
func GetPid(file string) (string, error) {
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
