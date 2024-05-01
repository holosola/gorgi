package ex

import "log"

func Handle(err error) {
	log.Fatalf("app got error: %s", err.Error())
	panic(err)
}
