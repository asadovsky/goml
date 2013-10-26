// Tests and benchmarks for Feature.
//
// To run tests:
//   go test goml/core
//
// To run benchmarks:
//   go test goml/core -bench=.
//   go test goml/core -bench=Sum

package core

import "testing"

// NOTE: We don't test for out-of-bounds offset, since there's no built-in way
// to test for panics.
func TestNewFeature(t *testing.T) {
	values := []float32{1, 2, 3}
	f := NewFeature(values)
	for i, v := range values {
		f_v := f.Value(i)
		if f_v != v {
			t.Error(f_v, v)
		}
	}
}

const BIG = 1e6

func makeBigArray() [BIG]float32 {
	var res [BIG]float32
	for i := 0; i < BIG; i++ {
		res[i] = float32(i)
	}
	return res
}

func makeBigSlice() []float32 {
	res := make([]float32, BIG)
	for i := 0; i < BIG; i++ {
		res[i] = float32(i)
	}
	return res
}

func BenchmarkSumArrayIndex(b *testing.B) {
	values := makeBigArray()
	for i := 0; i < b.N; i++ {
		var sum float32 = 0
		for j := 0; j < len(values); j++ {
			sum += values[j]
		}
	}
}

// 35% slower than BenchmarkSumArrayIndex!
func BenchmarkSumArrayRange(b *testing.B) {
	values := makeBigArray()
	for i := 0; i < b.N; i++ {
		var sum float32 = 0
		for _, v := range values {
			sum += v
		}
	}
}

// NOTE: Index is faster than Range, so the remaining benchmarks use index.

// Surprisingly, 3% faster than BenchmarkSumArrayIndex.
func BenchmarkSumSlice(b *testing.B) {
	values := makeBigSlice()
	for i := 0; i < b.N; i++ {
		var sum float32 = 0
		for j := 0; j < len(values); j++ {
			sum += values[j]
		}
	}
}

// 15% slower than BenchmarkSumSlice. Same speed if to extract values slice.
func BenchmarkSumFeature(b *testing.B) {
	f := NewFeature(makeBigSlice())
	for i := 0; i < b.N; i++ {
		var sum float32 = 0
		//values := f.Values
		for j := 0; j < f.Size(); j++ {
			sum += f.Value(j)
			//sum += values[j]
		}
	}
}
