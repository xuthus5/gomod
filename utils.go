package gomod

import "strings"

func ElemIn(elems []string, target string) bool {
	for _, elem := range elems {
		if strings.Contains(target, elem) {
			return true
		}
	}
	return false
}
