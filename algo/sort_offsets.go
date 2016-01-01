package algo

import "sort"

type offsetsAndValues struct {
	offsets []int
	values  []float32
}

func (s *offsetsAndValues) Len() int {
	return len(s.values)
}

func (s *offsetsAndValues) Less(i, j int) bool {
	return s.values[i] < s.values[j]
}

func (s *offsetsAndValues) Swap(i, j int) {
	s.offsets[i], s.offsets[j] = s.offsets[j], s.offsets[i]
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func SortOffsetsByValue(arr []float32) []int {
	s := &offsetsAndValues{
		offsets: make([]int, len(arr), len(arr)),
		values:  make([]float32, len(arr), len(arr)),
	}

	for i := 0; i < len(arr); i++ {
		s.offsets[i] = i
	}
	copy(s.values, arr)

	sort.Sort(s)
	return s.offsets
}
