package global

var (
	//service id=>name
	serviceId = map[string]string{}
)

func RegisterServiceId(id, serviceName string) {
	serviceId[id] = serviceName
}

func GetServiceId() (r map[string]string) {
	r = make(map[string]string)
	for k, v := range serviceId {
		r[k] = v
	}
	return
}
func MergeServiceId(r map[string]string) (map[string]string) {
	for k, v := range serviceId {
		r[k] = v
	}
	return r
}

//func GetServiceName(id string) (string, error) {
//	if n, ok := ServiceId[id]; ok {
//		return n, nil
//	} else {
//		return "", fmt.Errorf("empty service name")
//	}
//}
