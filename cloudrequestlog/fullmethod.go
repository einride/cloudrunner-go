package cloudrequestlog

import "strings"

func splitFullMethod(fullMethod string) (service, method string, ok bool) {
	serviceAndMethod := strings.SplitN(strings.TrimPrefix(fullMethod, "/"), "/", 2)
	if len(serviceAndMethod) != 2 {
		return "", "", false
	}
	return serviceAndMethod[0], serviceAndMethod[1], true
}
