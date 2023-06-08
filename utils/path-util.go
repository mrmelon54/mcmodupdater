package utils

import "path"

func PathJoin(elem ...string) string {
	if len(elem) >= 1 {
		if elem[0] == "" {
			return ""
		}
		return path.Join(elem...)
	}
	return ""
}
