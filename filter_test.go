package bloom

import (
	"encoding/binary"
	"hash/fnv"
	"testing"
)

func doubleFNV(b []byte) (uint64, uint64) {
	hx := fnv.New64()
	hx.Write(b)
	x := hx.Sum64()
	hy := fnv.New64a()
	hy.Write(b)
	y := hy.Sum64()
	return x, y
}

func TestFalsePositive(t *testing.T) {
	const n = 1e6
	bf := New(n, 1e-4, doubleFNV)

	buf := make([]byte, 8)
	for i := 0; i < n; i += 2 {
		binary.PutVarint(buf, int64(i))
		bf.Add(buf)
	}

	fp := 0 // false postive count
	for i := 1; i < n; i += 2 {
		binary.PutVarint(buf, int64(i))
		if bf.Test(buf) {
			fp++
		}
	}
	fpr := float64(fp) / n
	t.Logf("FP = %d, FPR = %.4f%%", fp, fpr*100)
}

func BenchmarkAdd(b *testing.B) {
	b.StopTimer()
	bf := New(1e6, 1e-4, doubleFNV)
	buf := make([]byte, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		binary.PutUvarint(buf, uint64(i))
		bf.Add(buf)
	}
}

func BenchmarkTest(b *testing.B) {
	b.StopTimer()
	bf := New(1e6, 1e-4, doubleFNV)
	buf := make([]byte, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		binary.PutUvarint(buf, uint64(i))
		bf.Test(buf)
	}
}
