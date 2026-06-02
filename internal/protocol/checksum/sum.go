package checksum

func Sum8(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum
}

// Sum is kept as a compatibility alias for the Step 1 baseline API.
func Sum(data []byte) byte {
	return Sum8(data)
}
