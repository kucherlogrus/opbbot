package utils

import "github.com/sanity-io/litter"

func PrintType(t interface{}) {
	litter.Dump(t)
}
