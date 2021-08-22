package lib

import "github.com/sanity-io/litter"

func PrintType(t interface{}) {
	litter.Dump(t)
}
