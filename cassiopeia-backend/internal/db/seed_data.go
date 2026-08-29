package db

import (
	"embed"
	"encoding/csv"
	"io"
)

//go:embed seed/*.csv
var seedFS embed.FS

// readSeedCSV returns the data rows of an embedded seed CSV, dropping the
// header row.
func readSeedCSV(name string) ([][]string, error) {
	f, err := seedFS.Open("seed/" + name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	var rows [][]string
	if _, err := r.Read(); err != nil { // header
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}
	for {
		row, err := r.Read()
		if err == io.EOF {
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
}
