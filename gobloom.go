package gobloom

import (
	"hash"
	"hash/fnv"
	"math"
)

const (
	logWordSize = 6
	wordSize    = 1 << logWordSize
)

type word uint64

type BloomFilter struct {
	m   uint64
	set []word
	k   uint64
	hx  hash.Hash64
	hy  hash.Hash64
	idx []uint64
}

// Create a Bloom filter with m bits and k hashes
func New(m, k uint64) *BloomFilter {
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
	return &BloomFilter{
		m:   m,
		set: make([]word, l),
		k:   k,
		hx:  fnv.New64a(),
		hy:  fnv.New64(),
		idx: make([]uint64, k),
	}
}

// Calculate the k slots in m to set/check
func (self *BloomFilter) index(b []byte) []uint64 {
	// Use enhanced double hashing technique based on this paper from
	// http://www.ccs.neu.edu/home/pete/research/bloom-filters-verification.html
	self.hx.Reset()
	self.hy.Reset()
	self.hx.Write(b)
	self.hy.Write(b)
	x := self.hx.Sum64()
	y := self.hy.Sum64()

	for i := uint64(0); i < uint64(self.k); i++ {
		self.idx[i] = x % self.m
		x += y
		y += i
	}
	return self.idx
}

// Add an element to the Bloom filter
func (self *BloomFilter) Add(b []byte) {
	for _, off := range self.index(b) {
		self.set[off>>logWordSize] |= 1 << (off & (wordSize - 1))
	}
}

// Test if an element is in the Bloom filter (might be false positive)
func (self *BloomFilter) Test(b []byte) bool {
	for _, off := range self.index(b) {
		if 0 == self.set[off>>logWordSize]&(1<<(off&(wordSize-1))) {
			return false
		}
	}
	return true
}

func (self *BloomFilter) Clear() {
	self.set = make([]word, len(self.set))
}

func (self *BloomFilter) FPR(n uint64) float64 {
	return fpr(float64(self.m), float64(n), float64(self.k))
}

// Optimal number of hashes
func bestK(m, n float64) float64 {
	return math.Ln2 * m / n
}

// Calculate the approximation of false positive rate
func fpr(m, n, k float64) float64 {
	return math.Pow((1 - math.Pow(math.E, -k*n/m)), k)
}
