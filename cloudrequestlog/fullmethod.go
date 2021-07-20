package cloudrequestlog

import "strings"

func splitFullMethod(fullMethod string) (service, method string) {
	serviceAndMethod := strings.SplitN(strings.TrimPrefix(fullMethod, "/"), "/", 2)
	return serviceAndMethod[0], serviceAndMethod[1]
}
