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
	h   hash.Hash64
}

// Create a Bloom filter with m bits and k hashes
func New(m, k uint64) *BloomFilter {
	l := (m + (wordSize - 1)) >> logWordSize
	if l == 0 {
		l = 1
	}
	return &BloomFilter{
		m:   m,
		set: make([]word, l),
		k:   k,
		h:   fnv.New64a(),
	}
}

// Add an element to the Bloom filter
func (self *BloomFilter) Add(b []byte) {
	h := self.hash(b)
	h0, h1 := h&0xFFFFFFFF, h>>32
	for i := uint64(0); i < self.k; i++ {
		off := (h0 + i*h1) % self.m
		self.set[off>>logWordSize] |= 1 << (off & (wordSize - 1))
	}
}

// Test if an element is in the Bloom filter (might be false positive)
func (self *BloomFilter) Test(b []byte) bool {
	h := self.hash(b)
	h0, h1 := h&0xFFFFFFFF, h>>32
	for i := uint64(0); i < self.k; i++ {
		off := (h0 + i*h1) % self.m
		if 0 == self.set[off>>logWordSize]&(1<<(off&(wordSize-1))) {
			return false
		}
	}
	return true
}

func (self *BloomFilter) Clear() {
	self.set = make([]word, len(self.set))
}

func (self *BloomFilter) hash(b []byte) uint64 {
	self.h.Write(b)
	r := self.h.Sum64()
	self.h.Reset()
	return r
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
