package algo

////////////////////////////////////////
// offsetRange

type offsetRange struct {
	begin, end int // [begin, end) offsets
}

func (b *offsetRange) Size() int {
	return b.end - b.begin
}

func (b *offsetRange) Empty() bool {
	return b.begin == b.end
}

////////////////////////////////////////
// sumWeight

type sumWeight struct {
	weightedSum float64
	totalWeight float64
}

func (sw *sumWeight) Add(value, weight float64) {
	sw.weightedSum += value * weight
	sw.totalWeight += weight
}

func (sw *sumWeight) Subtract(value, weight float64) {
	sw.weightedSum -= value * weight
	sw.totalWeight -= weight
}
