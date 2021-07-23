package common

import "sync"

type XlsxWriter interface {
	XlsxWrite() error
}

type ListGetter interface {
	NewItem()
	WriteHeader(xlsxWriter chan XlsxWriter, wg *sync.WaitGroup) error
	GetList(xlsxWriter chan XlsxWriter, wg *sync.WaitGroup) error
	GetWriten() string
}
