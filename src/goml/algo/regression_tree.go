// Package algo contains various ML algorithms.
package algo

import "log"
import "math"
import "sort"

// TODO:
//  - Support sparse inputs. I.e. use Feature instead of []float32.
//  - Support "equal" splits.

// NOTE: Until we switch to Feature objects, ids can be interpreted as offsets.

// TODO: Support different split types (GT, LT, or EQ). We should probably just
// define an "interval" struct to represent an interval of float values.
type split struct {
	// If inputIdx is -1, there's no split, and thresh is an adjustment.
	inputIdx int     // index of input to split on
	thresh   float32 // split threshold value
}

// TODO: Maybe compress splits/adjs at the end, in case some splits didn't
// happen. In this case, each split would store a newRegion int.
type Tree struct {
	// Invariant: len(splits) + 1 == len(adjs)
	// Split i has children 2*i+1 and 2*i+2.
	// Leaf split i points to adjs 2*i+1-n and 2*i+2-n where n is len(splits).
	splits []split
	adjs   []float32
}

func (t *Tree) Eval(inputs [][]float32) []float32 {
	// Map inputs to adjustments.
	// TODO: Optimize.
	output := make([]float32, len(inputs[0]))
	for i := 0; i < len(inputs[0]); i++ {
		region := 0
		for {
			if region >= len(t.splits) {
				output[i] = t.adjs[region-len(t.splits)]
				break
			}
			sp := t.splits[region]
			if sp.inputIdx == -1 {
				output[i] = sp.thresh
				break
			} else {
				region = 2*region + 1
				if inputs[sp.inputIdx][i] <= sp.thresh {
					region += 1 // not in split
				}
			}
		}
	}
	return output
}

// TODO: Change minItems to minWeight?
type treeBuilder struct {
	output   []float32   // values to predict
	weight   []float32   // per-output weights
	inputs   [][]float32 // input features
	maxDepth int         // maximum depth of tree
	minItems int         // minimum number of items needed to create a branch

	// Ids sorted first by region, then by id. Guaranteed to be full.
	byRegionIds []int

	// Per-input ids sorted first by region, then by value.
	// TODO: Should we allow this to be sparse, i.e. allow the []int's to have
	// different lengths? To support that, we'd need to maintain region boundaries
	// for each sparse input. Maybe better to include ids for missing values, and
	// intelligently skip those when looking for splits?
	byRegionInputs [][]int

	t *Tree // the final result
}

// BuildTree builds a single regression tree.
func BuildTree(output []float32, weight []float32, inputs [][]float32,
	maxDepth, minItems int) *Tree {
	tb := &treeBuilder{
		output:   output,
		weight:   weight,
		inputs:   inputs,
		maxDepth: maxDepth,
		minItems: minItems,
	}

	// Compute tb internals.
	maxRegions := 1 << (uint)(maxDepth)
	tb.t = &Tree{
		splits: make([]split, maxRegions-1),
		adjs:   make([]float32, maxRegions),
	}

	// TODO: Make [range(size)] slice construction a one-liner.
	tb.byRegionIds = make([]int, len(output))
	for i := 0; i < len(output); i++ {
		tb.byRegionIds[i] = i
	}

	tb.byRegionInputs = make([][]int, len(inputs))
	for i := 0; i < len(inputs); i++ {
		tb.byRegionInputs[i] = SortOffsetsByValue(inputs[i])
	}

	done := make(chan bool)
	go tb.ProcessRegion(offsetRange{0, len(output)}, 0, 0, done)
	for i := 0; i < maxRegions; i++ {
		<-done
	}
	return tb.t
}

func (tb *treeBuilder) ProcessRegion(region offsetRange, regionIdx, depth int, done chan<- bool) {
	// Try to split region.
	spEnd := tb.TryToSplitRegion(region, regionIdx, depth)

	if spEnd != 0 {
		// Process child regions.
		spReg, exReg := offsetRange{0, spEnd}, offsetRange{spEnd, region.end}
		go tb.ProcessRegion(spReg, regionIdx*2+1, depth+1, done)
		go tb.ProcessRegion(exReg, regionIdx*2+2, depth+1, done)
	} else {
		// Failed to split, so compute adjustment for region. Adjustment is simply
		// the weighted mean value of output in this region.
		// TODO: Maybe reuse sumWeight computed in TryToSplitRegion.
		sw := tb.ComputeSumWeight(region)
		adj := float32(sw.weightedSum / sw.totalWeight)
		// We stopped splitting before hitting maxDepth.
		if regionIdx < len(tb.t.splits) {
			tb.t.splits[regionIdx].inputIdx = -1
			tb.t.splits[regionIdx].thresh = adj
		} else {
			adjIdx := regionIdx - len(tb.t.splits)
			tb.t.adjs[adjIdx] = adj
		}
		for i := 0; i < 1<<(uint)(tb.maxDepth-depth); i++ {
			done <- true
		}
	}
}

func (tb *treeBuilder) ComputeSumWeight(region offsetRange) sumWeight {
	res := sumWeight{}
	for i := region.begin; i < region.end; i++ {
		offset := tb.byRegionIds[i] // note, id == offset
		res.Add(float64(tb.output[offset]), float64(tb.weight[offset]))
	}
	return res
}

type splitInfo struct {
	inputIdx int
	region   offsetRange
	thresh   float32
	gain     float64
}

func (tb *treeBuilder) FindBestSplitForInput(region offsetRange, inputIdx int, c chan<- splitInfo) {
	// Note: Minimizing squared error is equivalent to maximizing
	// gain(spSumWeight) + gain(exSumWeight), with gain defined as below.
	gain := func(sw sumWeight) float64 {
		return sw.weightedSum * (sw.weightedSum / sw.totalWeight)
	}
	// Start with all items included in split.
	spSumWeight, exSumWeight := tb.ComputeSumWeight(region), sumWeight{}
	// Try each split for current input.
	// TODO: Use an iterator of some sort.
	// TODO: Handle missing values.
	// TODO: Search for equal splits.
	input := tb.inputs[inputIdx]
	byValue := tb.byRegionInputs[inputIdx][region.begin:region.end]
	prevValue := float32(math.Inf(-1))
	bestSplit := splitInfo{gain: -1}
	for i := 0; i < len(byValue); i++ {
		id := byValue[i]
		value := input[id]
		output, weight := float64(tb.output[id]), float64(tb.weight[id])

		curGain := gain(spSumWeight) + gain(exSumWeight)
		spSumWeight.Subtract(output, weight)
		exSumWeight.Add(output, weight)

		// TODO: Smarter float comparison.
		// http://www.cygnus-software.com/papers/comparingfloats/comparingfloats.htm
		if value < prevValue+1e-6 {
			prevValue = value
			continue
		}
		prevMid := (prevValue + value) / 2
		prevValue = value

		// TODO: Move this out of the inner loop.
		if i < tb.minItems || len(byValue)-i < tb.minItems {
			continue
		}
		// TODO: Profile moving the 1e-6 into bestGain.
		if curGain > bestSplit.gain+1e-6 {
			// Includes ids i through end.
			bestSplit = splitInfo{inputIdx, offsetRange{i, len(byValue)}, prevMid, curGain}
		}
	}
	c <- bestSplit
}

// Returns end offset of in-split range, or 0 if no split was found.
// Updates byRegionIds and byRegionInputs as needed.
func (tb *treeBuilder) TryToSplitRegion(region offsetRange, splitIdx, depth int) int {
	// If split would violate max depth or min items, give up.
	// TODO: Make the size check handle missing values.
	if depth+1 > tb.maxDepth || region.Size() < 2*tb.minItems {
		return 0
	}

	// Algorithm:
	// - For each input, try each split; pick the one that minimizes squared error.
	// - Pick the best split across inputs.
	// - If a split was found, update internal state.

	c := make(chan splitInfo)
	for inputIdx := 0; inputIdx < len(tb.inputs); inputIdx++ {
		go tb.FindBestSplitForInput(region, inputIdx, c)
	}

	var bestSplit *splitInfo
	for inputIdx := 0; inputIdx < len(tb.inputs); inputIdx++ {
		si := <-c
		log.Print(si)
		// Break ties using inputIdx (for determinism).
		if bestSplit == nil || si.gain > bestSplit.gain ||
			(si.gain == bestSplit.gain && si.inputIdx < bestSplit.inputIdx) {
			bestSplit = &si
		}
	}
	log.Print(bestSplit)
	if bestSplit.region.Empty() {
		return 0
	}

	idsInSplit := tb.byRegionInputs[bestSplit.inputIdx][bestSplit.region.begin:bestSplit.region.end]
	tb.t.splits[splitIdx] = split{bestSplit.inputIdx, bestSplit.thresh}

	// Update internal state:
	// - byRegionIds: resort subslice by (region, id).
	// - byRegionInputs: resort each subslice by (region, value for id).

	// Re-sort the relevant subslice of byRegionIds so that in-split ids come
	// before not-in-split ids. Running time is O(n + m) where n is num ids total
	// and m is num ids in region. Note, this works even for equal splits.
	idIsInSplit := make([]bool, len(tb.output))
	for i := 0; i < len(idsInSplit); i++ {
		idIsInSplit[idsInSplit[i]] = true
	}

	byRegionIdsSlice := tb.byRegionIds[region.begin:region.end]
	exIdx := len(byRegionIdsSlice) - 1
	for i := exIdx; i >= 0; i-- {
		id := byRegionIdsSlice[i]
		if !idIsInSplit[id] {
			byRegionIdsSlice[exIdx] = id
			exIdx--
		}
	}
	spIdx := 0
	for i := 0; i < len(idIsInSplit); i++ {
		if idIsInSplit[i] {
			byRegionIdsSlice[spIdx] = i
			spIdx++
		}
	}

	// Update byRegionInputs. Note, this uses idIsInSplit from above.
	// TODO: Maybe special-case the input we split on.
	for i := 0; i < len(tb.byRegionInputs); i++ {
		byValueIds := tb.byRegionInputs[i][region.begin:region.end]
		sort.Sort(&sortByRegionInputsStruct{
			idsInRegion: byValueIds,
			idIsInSplit: idIsInSplit,
		})
	}

	return bestSplit.region.Size()
}

////////////////////////////////////////
// sortByRegionInputsStruct

type sortByRegionInputsStruct struct {
	idsInRegion []int
	idIsInSplit []bool
}

func (s *sortByRegionInputsStruct) Len() int {
	return len(s.idsInRegion)
}

func (s *sortByRegionInputsStruct) Less(i, j int) bool {
	iId, jId := s.idsInRegion[i], s.idsInRegion[j]
	if s.idIsInSplit[iId] {
		if !s.idIsInSplit[jId] {
			return true
		} else {
			return i < j // ids are already sorted by value
		}
	} else if s.idIsInSplit[jId] {
		return false
	}
	return i < j // neither are in split
}

func (s *sortByRegionInputsStruct) Swap(i, j int) {
	s.idsInRegion[i], s.idsInRegion[j] = s.idsInRegion[j], s.idsInRegion[i]
}
