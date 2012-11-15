package gobloom

import (
	"hash/fnv"
	"math"
	"sync"
)

const (
	logWordSize = 6
	wordSize    = 1 << logWordSize
)

type word uint64

type BloomFilter interface {
	Add(b []byte)
	Test(b []byte) bool
	Clear()
	FPR(n uint64) float64
}

type standardBloomFilter struct {
	sync.RWMutex
	m   uint64
	set []word
	k   uint64
}

// Create a standard Bloom filter with m bits and k hashes
func New(m, k uint64) BloomFilter {
	// NOTE: m cannot be greater than 2^38 (256GB) due to Go 1.0's limition of
	// int being only 32-bit even on 64-bit machines. Go 1.1 is supposed to allow
	// 64-bit in on 64-bit machine, which will make this problem disappear. 
	if m > (1<<38 - 1) {
		panic("m overflows")
	}

	l := (m + (wordSize - 1)) >> logWordSize
	if l == 0 {
		l = 1
	}
	return &standardBloomFilter{
		m:   m,
		set: make([]word, l),
		k:   k,
	}
}

// Calculate the k slots in m to set/check
func (self *standardBloomFilter) index(b []byte) []uint64 {
	// Use enhanced double hashing technique based on this paper from
	// http://www.ccs.neu.edu/home/pete/research/bloom-filters-verification.html
	hx := fnv.New64a()
	hy := fnv.New64()
	hx.Write(b)
	hy.Write(b)
	x := hx.Sum64()
	y := hy.Sum64()

	idx := make([]uint64, self.k)
	for i := uint64(0); i < uint64(self.k); i++ {
		idx[i] = x % self.m
		x += y
		y += i
	}
	return idx
}

// Add an element to the Bloom filter
func (self *standardBloomFilter) Add(b []byte) {
	self.Lock()
	defer self.Unlock()
	for _, off := range self.index(b) {
		self.set[off>>logWordSize] |= 1 << (off & (wordSize - 1))
	}
}

// Test if an element is in the Bloom filter (might be false positive)
func (self *standardBloomFilter) Test(b []byte) bool {
	self.RLock()
	defer self.RUnlock()
	for _, off := range self.index(b) {
		if 0 == self.set[off>>logWordSize]&(1<<(off&(wordSize-1))) {
			return false
		}
	}
	return true
}

func (self *standardBloomFilter) Clear() {
	self.Lock()
	defer self.Unlock()
	self.set = make([]word, len(self.set))
}

func (self *standardBloomFilter) FPR(n uint64) float64 {
	return FPR(float64(self.m), float64(n), float64(self.k))
}

// Optimal number of hashes
func BestK(m, n float64) float64 {
	return math.Ln2 * m / n
}

// Calculate the approximation of false positive rate
func FPR(m, n, k float64) float64 {
	return math.Pow((1 - math.Pow(math.E, -k*n/m)), k)
}
