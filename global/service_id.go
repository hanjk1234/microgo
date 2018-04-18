package global

import "fmt"

var (
	//service id=>name
	ServiceId = map[string]string{}
)

func RegisterServiceId(id, serviceName string) {
	ServiceId[id] = serviceName
}

func GetServiceName(id string) (string, error) {
	if n, ok := ServiceId[id]; ok {
		return n, nil
	} else {
		return "", fmt.Errorf("empty service name")
	}
}
