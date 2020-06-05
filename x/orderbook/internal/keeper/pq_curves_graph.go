package keeper

import (
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SDGraphValue struct {
	SupplyInited bool
	DemandInited bool
}

type SDGraphValues []SDGraphValue

func (c *SDCurves) Graph() string {
	const pointStrLen = 5

	blankStr, dashStr := "", ""
	for i := 0; i < pointStrLen-1; i++ {
		blankStr += " "
		dashStr += "--"
	}

	// build axis
	xAxis := make([]sdk.Uint, 0)
	yAxis := make([]sdk.Uint, 0)
	for _, item := range *c {
		xAxis = append(xAxis, item.Price)

		ySupplyTickFound, yDemandTickFound := false, false
		for _, yTick := range yAxis {
			if yTick.Equal(item.Supply) {
				ySupplyTickFound = true
			}
			if yTick.Equal(item.Demand) {
				yDemandTickFound = true
			}
		}
		if !ySupplyTickFound {
			yAxis = append(yAxis, item.Supply)
		}
		if !yDemandTickFound {
			yAxis = append(yAxis, item.Demand)
		}
	}
	sort.Slice(yAxis, func(i, j int) bool {
		return yAxis[i].GT(yAxis[j])
	})

	// build values matrix
	matrix := make([]SDGraphValues, 0, len(xAxis))
	for _, item := range *c {
		yItems := make(SDGraphValues, len(yAxis), len(yAxis))

		for i := 0; i < len(yAxis); i++ {
			yTick := yAxis[i]
			yItem := &yItems[i]
			if item.Supply.Equal(yTick) {
				yItem.SupplyInited = true
			}
			if item.Demand.Equal(yTick) {
				yItem.DemandInited = true
			}
		}

		matrix = append(matrix, yItems)
	}

	// draw
	trimUint := func(v sdk.Uint) string {
		s := v.String()
		if len(s) <= pointStrLen {
			return blankStr[:pointStrLen-len(s)] + s
		}
		s = s[:pointStrLen-2]

		return s + ".."
	}
	blankPoint := blankStr + " "
	bidPoint := blankStr + "b"
	askPoint := blankStr + "a"
	crossPoint := blankStr + "x"

	graph := strings.Builder{}
	// Y axis + Separator + Rows
	for yRowIdx, yTick := range yAxis {
		graph.WriteString(fmt.Sprintf("%s | ", trimUint(yTick)))

		for xColIdx := range xAxis {
			matrixItem := matrix[xColIdx][yRowIdx]
			for {
				if matrixItem.SupplyInited && matrixItem.DemandInited {
					graph.WriteString(crossPoint)
					break
				}

				if matrixItem.SupplyInited {
					graph.WriteString(askPoint)
					break
				}

				if matrixItem.DemandInited {
					graph.WriteString(bidPoint)
					break
				}

				graph.WriteString(blankPoint)
				break
			}
			graph.WriteString(" ")
		}

		graph.WriteString("\n")
	}

	// Separator + X axis
	for i := 0; i < len(xAxis) + 1; i++ {
		graph.WriteString(dashStr)
	}
	graph.WriteString("\n")

	graph.WriteString(blankPoint + " | ")
	for _, xTick := range xAxis {
		graph.WriteString(fmt.Sprintf("%s ", trimUint(xTick)))
	}
	graph.WriteString("\n")

	return graph.String()
}
