package bf

import (
	"hash"
	"hash/crc64"
	"hash/fnv"
	"strings"
	"sync"
)

// BloomFilter is a bloom filter.
type BloomFilter struct {
	bitset  []bool
	hashfns []hash.Hash64
	mu      sync.Mutex
}

// Insert puts data into the bloom filter.
func (bl *BloomFilter) Insert(data []byte) {
	// generate sums
	sums := bl.sums(data)
	// lock + unlock
	bl.mu.Lock()
	defer bl.mu.Unlock()
	// add sums to bitset
	for _, sum := range sums {
		i := bl.index(sum)
		bl.bitset[i] = true
	}
}

// Search queries the bloom filter with data.
func (bl *BloomFilter) Search(data []byte) (pos bool) {
	// generate sums
	sums := bl.sums(data)
	// lock + unlock
	bl.mu.Lock()
	defer bl.mu.Unlock()
	// test sums
	for _, sum := range sums {
		i := bl.index(sum)
		if bl.bitset[i] {
			pos = true
			return
		}
	}
	return
}

func (bl *BloomFilter) sums(data []byte) (sums []uint64) {
	sums = make([]uint64, len(bl.hashfns))
	for i, hashfn := range bl.hashfns {
		hashfn.Write(data)
		sum := hashfn.Sum64()
		sums[i] = sum
		hashfn.Reset()
	}
	return sums
}

func (bl *BloomFilter) index(sum uint64) (i int) {
	return int(sum % uint64(len(bl.bitset)))
}

// String returns a string representation of the bloom filter's bitset as a
// series of ones and zeros.
func (bl *BloomFilter) String() string {
	// lock + unlock
	bl.mu.Lock()
	defer bl.mu.Unlock()
	// build string
	var sb strings.Builder
	sb.Grow(len(bl.bitset))
	for _, bit := range bl.bitset {
		c := '0'
		if bit {
			c = '1'
		}
		sb.WriteRune(c)
	}
	return sb.String()
}

// MarshalBinary returns the bloom filter's bitset in binary. It never returns an
// error.
func (bl *BloomFilter) MarshalBinary() (data []byte, err error) {
	// lock + unlock
	bl.mu.Lock()
	defer bl.mu.Unlock()
	// build binary
	data = make([]byte, (len(bl.bitset)+7)/8)
	for i, bit := range bl.bitset {
		if bit {
			data[i/8] |= 0x80 >> (i % 8)
		}
	}
	return
}

// UnmarshalBinary reads from data in binary and writes to bloom filter's bitset.
// It never returns an error.
func (bl *BloomFilter) UnmarshalBinary(data []byte) (err error) {
	// lock + unlock
	bl.mu.Lock()
	defer bl.mu.Unlock()
	// read binary to bitset
	bl.bitset = make([]bool, len(data)*8)
	for i, bit := range data {
		for j := 0; j < 8; j++ {
			if (bit<<j)&0x80 == 0x80 {
				bl.bitset[(i*8)+j] = true
			}
		}
	}
	return
}

// Basic returns a basic bloom filter with four has functions given a size. The
// size is rounded to the nearest 8.
func Basic(size int) *BloomFilter {
	return Custom(
		size,
		fnv.New64(),
		fnv.New64a(),
		crc64.New(crc64.MakeTable(crc64.ISO)),
		crc64.New(crc64.MakeTable(crc64.ECMA)),
	)
}

// Custom returns a bloom filter given a size and hash functions. The size is
// rounded to the nearest 8.
func Custom(size int, hashfns ...hash.Hash64) *BloomFilter {
	size += size % 8
	return &BloomFilter{
		bitset:  make([]bool, size),
		hashfns: hashfns,
	}
}
