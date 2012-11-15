package gobloom

import (
	"encoding/binary"
	"testing"
)

func ExampleFPR(t *testing.T) {
	for b := 8; b < 20; b++ {
		n := 1e6
		m := float64(b) * n
		for k := 1; k < 9; k++ {
			kOpt := bestK(m, n)
			fprOpt := fpr(m, n, kOpt)
			fprK := fpr(m, n, float64(k))
			t.Logf("m/n = %d, k* = %.1f, fpr(k*) = %6.2f%%, fpr(k=%d) = %6.2f%%",
				b, kOpt, fprOpt*100, k, fprK*100)
		}
		t.Logf("________")
	}
}

func TestFalseRate(t *testing.T) {
	bf := New(10000000, 5)
	for i := 0; i < 1000000; i += 2 {
		buf := make([]byte, 8)
		n := binary.PutVarint(buf, int64(i))
		if n == 0 {
			t.Fatalf("Encoding failed")
		}
		bf.Add(buf)
	}

	// count false negative: should be none
	for i := 0; i < 1000000; i += 2 {
		buf := make([]byte, 8)
		binary.PutVarint(buf, int64(i))
		if !bf.Test(buf) {
			t.Fatalf("False negative occurred")
		}
	}

	// count false positive
	fp := 0
	for i := 1; i < 1000000; i += 2 {
		buf := make([]byte, 8)
		binary.PutVarint(buf, int64(i))
		if bf.Test(buf) {
			fp++
		}
	}
	fpr := float64(fp) / 1000000
	estFpr := bf.FPR(1000000)

	t.Logf("FP = %d, FPR = %.4f%%, FPR est. = %.4f%%", fp, fpr*100, estFpr*100)

	if fpr > estFpr {
		t.Fatalf("false positive rate too high: %.4f%% > %.4f%%", fpr*100, estFpr*100)
	}
}

func BenchmarkBloomFilterAdd(b *testing.B) {
	b.StopTimer()
	bf := New(8*1e6, 4)
	buf := make([]byte, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		binary.PutUvarint(buf, uint64(i))
		bf.Add(buf)
	}
}

func BenchmarkBloomFilterTest(b *testing.B) {
	b.StopTimer()
	bf := New(8*1e6, 4)
	buf := make([]byte, 20)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		binary.PutUvarint(buf, uint64(i))
		bf.Test(buf)
	}
}
