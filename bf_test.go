package bf_test

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/patrickmcnamara/bf"
)

func TestBloomFilterForPositives(t *testing.T) {
	bl := bf.Basic(10)
	poss := bytes.Split([]byte("Hey there, buddy! My name is Patrick McNamara."), []byte{' '})
	for _, pos := range poss {
		bl.Insert(pos)
	}
	for _, pos := range poss {
		if !bl.Search(pos) {
			t.Errorf("%q tested negative but shouldn't have", string(pos))
		}
	}
}

func TestBloomFilterForNegatives(t *testing.T) {
	bl := bf.Basic(10)
	negs := bytes.Split([]byte("My credit card details are ..."), []byte{' '})
	for _, neg := range negs {
		if bl.Search(neg) {
			t.Errorf("%q tested positive but shouldn't have", string(neg))
		}
	}
}

func TestBloomFilterForPositivesAndNegatives(t *testing.T) {
	// new bloom filter
	bl := bf.Basic(100000)

	// generate positives and negatives
	poss := make([][]byte, 500)
	for i := 0; i < len(poss); i++ {
		poss[i] = make([]byte, 64)
		rand.Read(poss[i])
	}
	negs := make([][]byte, 500)
	for i := 0; i < len(negs); i++ {
		negs[i] = make([]byte, 64)
		rand.Read(negs[i])
	}

	// add positives to bloom filter
	for _, pos := range poss {
		bl.Insert(pos)
	}

	// pretty print random data
	pretty := func(data []byte) string {
		return base64.RawURLEncoding.EncodeToString(data)[:32] + "..."
	}

	// test positives for positivity
	for _, pos := range poss {
		if !bl.Search(pos) {
			t.Errorf("%q tested negative but shouldn't have", pretty(pos))
		}
	}

	// test negatives for negativity
	negC := 0
	for _, neg := range negs {
		if bl.Search(neg) {
			negC++
		}
	}
	if negC > 50 {
		t.Errorf("%d, more than 50, tested negative but shouldn't have", negC)
	}
}

func TestBloomFilterString(t *testing.T) {
	bl := bf.Basic(100)
	bl.Insert([]byte("Hello, world!")) // [14 74 76 92]
	bls := bl.String()
	for i, c := range bls {
		if i == 14 || i == 74 || i == 76 || i == 92 {
			if c != '1' {
				t.Errorf("%d bit should be 1", i)
			}
		} else if c != '0' {
			t.Errorf("%d bit should be 0", i)
		}
	}
	if t.Failed() {
		var bitsset []int
		for i, c := range bls {
			if c == '1' {
				bitsset = append(bitsset, i)
			}
		}
		t.Logf("%v are set", bitsset)
	}
}

func TestBloomFilterBinary(t *testing.T) {
	// new bloom filter
	bl := bf.Basic(10000)

	// add datas to bloom filter
	for i := 0; i < 500; i++ {
		data := make([]byte, 64)
		rand.Read(data)
		bl.Insert(data)
	}

	// encode to binary and unencode
	bef := bl.String()
	data, _ := bl.MarshalBinary()
	bl.UnmarshalBinary(data)
	aft := bl.String()

	if bef != aft {
		t.Error("bloom filter is different before and after encoding")
	}
}
