package defaulttypes

import (
	"log"
	"regexp"
)

type HeadSheet struct {
	SheetName string
	HeadNames []string
}

var ReDigit = regexp.MustCompile(`\D`)

type XlsxWriter interface {
	XlsxWrite() error
}

func (h *HeadSheet) XlsxWrite() error {
	log.Println("I am   HEADSHEET writer")
	return nil
}
