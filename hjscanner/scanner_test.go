package hjscanner

import (
	"testing"

	"github.com/myafeier/log"
)

func init() {
	log.SetLogLevel(log.DEBUG)
	log.SetPrefix("HJ")

}

func TestDefaultScanner(t *testing.T) {

	chipType := new(ChipType)
	chipType.Id = 1
	chipType.Col = 2
	chipType.Row = 8
	chipType.CellHeight = 5
	chipType.CellWidth = 5
	chipType.CellSpacingX = 3
	chipType.CellSpacingY = 3
	chipType.PaddingLeft = 10
	chipType.PaddingTop = 10

	DefaultScaner.SetChip(1, chipType)
	t.Logf("chip:%+v\n", DefaultScaner.Chips[0])
	for _, v := range DefaultScaner.Chips[0].Cells {
		t.Logf("cell:%+v\n", v)
	}

}
