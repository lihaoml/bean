package bean

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

type TimeSeries []TimePoint

type TimePoint struct {
	Time  time.Time
	Value float64
}

func (ts TimeSeries) ToCSV(filename string) {
	csvFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()
	var data [][]string

	head := []string{
		"Time",
		"Value",
	}
	data = append(data, head)
	for _, v := range ts {
		s := []string{
			v.Time.Format(time.RFC3339),
			fmt.Sprint(v.Value),
		}
		data = append(data, s)
	}
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.WriteAll(data)
	csvWriter.Flush()
}
