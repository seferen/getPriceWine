package main

import (
	"log"
	"runtime"
	"sync"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/seferen/getPriceWine/auchan"
	"github.com/seferen/getPriceWine/bristol"
	"github.com/seferen/getPriceWine/common"
	"github.com/seferen/getPriceWine/lenta"
	"github.com/seferen/getPriceWine/metro"
	"github.com/seferen/getPriceWine/winelab"
)

var elements []common.ListGetter

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Application was started")
	elements = []common.ListGetter{
		&bristol.Bristol{}, //fine
		&metro.Metro{},     // fine
		&lenta.Lenta{},     //fine
		&auchan.Auchan{},   //fine
		&winelab.Winelab{}, //fine
	}
}

var wg = sync.WaitGroup{}

var rg = sync.WaitGroup{}

func main() {

	defer func() {
		log.Println("**********was writen***********")
		for _, elem := range elements {
			log.Println(elem.GetWriten())
		}
		log.Println("*******************************")
		log.Println("Application was finished")
	}()

	// create channel for writing to xlsx file information
	fWr := make(chan common.XlsxWriter, 10)
	qu := make(chan int)

	// // create file for writing information

	defer func(file *excelize.File) {
		if err := file.SaveAs("result.xlsx"); err != nil {
			log.Println(err)
		}
	}(common.Xlsx)

	go func(xlsxWriter chan common.XlsxWriter, quite chan int) {

		for _, elem := range elements {
			elem.NewItem()
			wg.Add(2)
			go check(elem.WriteHeader(xlsxWriter, &wg))
			go check(elem.GetList(xlsxWriter, &wg))
		}

		wg.Wait()
		rg.Wait()

		quite <- 0

	}(fWr, qu)

	// processing
	for {
		select {
		case obj := <-fWr:
			rg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				err := obj.XlsxWrite()
				if err != nil {
					log.Println(err)
				}
			}(&rg)

		case <-qu:
			log.Println("stop working job")
			return
		default:
		}
	}

}

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}
