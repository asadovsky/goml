// Tests and benchmarks for BuildTree.
//
// To run tests:
//   go test goml/algo
//   go test goml/algo -run=Affects
//
// To run benchmarks:
//   go test goml/algo -bench=.

package algo

import "testing"

func (t *Tree) Equals(other *Tree) bool {
	if len(t.splits) != len(other.splits) {
		return false
	}
	for i := 0; i < len(t.splits); i++ {
		if t.splits[i] != other.splits[i] {
			return false
		}
	}
	if len(t.adjs) != len(other.adjs) {
		return false
	}
	for i := 0; i < len(t.adjs); i++ {
		if t.adjs[i] != other.adjs[i] {
			return false
		}
	}
	return true
}

func SliceEqual(a, b []float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// maxDepth is 0, so we should return a Tree with no splits, where the
// adjustment is the average of output values.
func TestConstantTree(t *testing.T) {
	output := []float32{1, 2, 3, 4}
	weight := []float32{1, 1, 1, 1}
	inputs := [][]float32{
		[]float32{0, 0, 1, 1},
		[]float32{0, 1, 1, 1},
	}
	actual := BuildTree(output, weight, inputs, 0, 0)

	// There should be zero splits and one adjustment.
	// Adjustment should be mean([1, 2, 3, 4]) = 2.5.
	expected := &Tree{adjs: []float32{2.5}}
	if !expected.Equals(actual) {
		t.Error(actual)
	}

	// Check eval output.
	x := actual.Eval(inputs)
	if !SliceEqual(x, []float32{2.5, 2.5, 2.5, 2.5}) {
		t.Error(x)
	}
}

// minItems is 3, so we can't make any splits, so we should return a Tree with
// no splits.
func TestMinItemsNoSplits(t *testing.T) {
	output := []float32{1, 2, 3, 4}
	weight := []float32{1, 1, 1, 1}
	inputs := [][]float32{
		[]float32{0, 0, 1, 1},
		[]float32{0, 1, 1, 1},
	}
	actual := BuildTree(output, weight, inputs, 2, 3)
	// Note, len(splits) should be 2^2 - 1 = 3, but the first split should have
	// inputIdx -1 to indicate that no split was made.
	if len(actual.splits) != 3 {
		t.Error(actual.splits)
	}
	if actual.splits[0].inputIdx != -1 {
		t.Error(actual.splits)
	}
	if actual.splits[0].thresh != 2.5 {
		t.Error(actual.splits)
	}

	// Check eval output.
	x := actual.Eval(inputs)
	if !SliceEqual(x, []float32{2.5, 2.5, 2.5, 2.5}) {
		t.Error(x)
	}
}

// Simple depth-1 tree.
func TestDepth1Tree(t *testing.T) {
	output := []float32{1, 2, 3, 4, 5}
	weight := []float32{1, 1, 1, 1, 1}
	inputs := [][]float32{
		[]float32{0, 0, 1, 1, 1},
		[]float32{0, 1, 1, 1, 1},
	}
	actual := BuildTree(output, weight, inputs, 1, 0)

	// There should be one split and two adjustments.
	// Split should be on inputs[0], split thresh should be mean([0, 1]) = 0.5.
	// First adjustment should be 4, second should be 1.5.
	expected := &Tree{
		splits: []split{split{inputIdx: 0, thresh: 0.5}},
		adjs:   []float32{4, 1.5},
	}
	if !expected.Equals(actual) {
		t.Error(actual)
	}

	// Check eval output.
	x := actual.Eval(inputs)
	if !SliceEqual(x, []float32{1.5, 1.5, 4, 4, 4}) {
		t.Error(x)
	}
}

// Same as above, but with scrambled output, weight, inputs, and order of inputs.
// NOTE: For this test, we intentionally made it so the best split creates
// regions of different sizes.
func TestDepth1TreeScrambled(t *testing.T) {
	output := []float32{3, 1, 5, 4, 2}
	weight := []float32{1, 1, 1, 1, 1}
	inputs := [][]float32{
		[]float32{1, 0, 1, 1, 1},
		[]float32{1, 0, 1, 1, 0},
	}
	actual := BuildTree(output, weight, inputs, 1, 0)
	expected := &Tree{
		splits: []split{split{inputIdx: 1, thresh: 0.5}},
		adjs:   []float32{4, 1.5},
	}
	if !expected.Equals(actual) {
		t.Error(actual)
	}

	// Check eval output.
	x := actual.Eval(inputs)
	if !SliceEqual(x, []float32{4, 1.5, 4, 4, 1.5}) {
		t.Error(x)
	}
}

func TestMinItemsAffectsSplit(t *testing.T) {
	output := []float32{1, 5, 8, 8}
	weight := []float32{1, 1, 1, 1}
	inputs := [][]float32{
		[]float32{0, 0, 1, 1},
		[]float32{0, 1, 1, 1},
	}
	tree0 := BuildTree(output, weight, inputs, 1, 0)
	tree1 := BuildTree(output, weight, inputs, 1, 1)
	tree2 := BuildTree(output, weight, inputs, 1, 2)

	// minItems=0 and minItems=1 should give the same tree.
	if !tree0.Equals(tree1) {
		t.Error(tree0, tree1)
	}
	// With minItems=2, we should get a different tree; instead of splitting on
	// input 1, we split on input 0.
	if tree0.Equals(tree2) {
		t.Error(tree0, tree2)
	}

	// Check eval output.
	x := tree0.Eval(inputs)
	if !SliceEqual(x, []float32{1, 7, 7, 7}) {
		t.Error(x)
	}
	x = tree2.Eval(inputs)
	if !SliceEqual(x, []float32{3, 3, 8, 8}) {
		t.Error(x)
	}
}

func TestWeight(t *testing.T) {
	output := []float32{0, 2, 5}
	inputs := [][]float32{
		[]float32{0, 1, 2},
	}

	weight0 := []float32{1, 1, 1}
	weight1 := []float32{1, 1, 0.2}

	actual0 := BuildTree(output, weight0, inputs, 1, 0)
	actual1 := BuildTree(output, weight1, inputs, 1, 0)

	if actual0.Equals(actual1) {
		t.Error(actual0, actual1)
	}

	expected0 := &Tree{
		splits: []split{split{inputIdx: 0, thresh: 1.5}},
		adjs:   []float32{5, 1},
	}
	// Adjustment should be weighted average of 2 and 5.
	// (1*2 + 0.2*5) / (1 + 0.2) = 3 / 1.2 = 2.5.
	expected1 := &Tree{
		splits: []split{split{inputIdx: 0, thresh: 0.5}},
		adjs:   []float32{2.5, 0},
	}

	if !expected0.Equals(actual0) {
		t.Error(actual0)
	}
	if !expected1.Equals(actual1) {
		t.Error(actual1)
	}
}
