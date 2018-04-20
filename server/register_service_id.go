package server

import "github.com/seefan/microgo/global"

func RegisterServiceId(serviceName, id string) {
	global.ServiceId[serviceName] = id
}
