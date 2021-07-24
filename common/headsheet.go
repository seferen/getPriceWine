package common

type HeadSheet struct {
	SheetName string
	HeadNames []string
}

func (h *HeadSheet) XlsxWrite() error {
	// create a new sheat
	Xlsx.NewSheet(h.SheetName)

	// write headers of sheat
	for i, v := range h.HeadNames {
		Xlsx.SetCellValue(h.SheetName, string(abs[i])+"1", v)

	}
	return nil
}
