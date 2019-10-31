package util

// RmZeros removes negatives as unrelated (i.e., 0)
func RmZeros(vals []float64) []float64 {
	for i, v := range vals {
		if v < 0 {
			vals[i] = 0
		}
	}
	return vals
}

// RmHeader removes the header row of a table
// TODO: Make checks for header content before removal
func RmHeader(records [][]string) [][]string {
	return records[1:]
}