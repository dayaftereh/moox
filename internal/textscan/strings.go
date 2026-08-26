package textscan

// String is one printable ASCII run found inside a binary payload.
type String struct {
	Offset int    `json:"offset"`
	Value  string `json:"value"`
}

// ASCII returns printable ASCII runs of at least minLen bytes.
func ASCII(data []byte, minLen int) []String {
	if minLen < 1 {
		minLen = 1
	}
	var out []String
	start := -1
	flush := func(end int) {
		if start >= 0 && end-start >= minLen {
			out = append(out, String{Offset: start, Value: string(data[start:end])})
		}
		start = -1
	}
	for i, b := range data {
		if b >= 0x20 && b <= 0x7e {
			if start < 0 {
				start = i
			}
			continue
		}
		flush(i)
	}
	flush(len(data))
	return out
}
