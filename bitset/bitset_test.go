package bitset

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBitSet_All(t *testing.T) {
	Convey("BitSet 测试", t, func() {
		bitset := NewBitSet()
		So(bitset.Empty(), ShouldBeTrue)

		Convey("添加一个数", func() {
			bitset.Add(10)
			for i := uint64(0); i < uint64(64); i++ {
				if i == 10 {
					So(bitset.Has(i), ShouldBeTrue)
				} else {
					So(bitset.Has(i), ShouldBeFalse)
				}
			}
			So(bitset.Empty(), ShouldBeFalse)
		})

		Convey("添加多个数", func() {
			bitset.Add(20)
			bitset.Add(30)
			bitset.Add(40)
			for i := uint64(0); i < uint64(64); i++ {
				if i == 20 || i == 30 || i == 40 {
					So(bitset.Has(i), ShouldBeTrue)
				} else {
					So(bitset.Has(i), ShouldBeFalse)
				}
			}
			So(bitset.Empty(), ShouldBeFalse)
		})

		Convey("删除多个数", func() {
			bitset.Add(20)
			bitset.Add(30)
			bitset.Add(40)
			bitset.Del(30)
			for i := uint64(0); i < uint64(64); i++ {
				if i == 20 || i == 40 {
					So(bitset.Has(i), ShouldBeTrue)
				} else {
					So(bitset.Has(i), ShouldBeFalse)
				}
			}
			So(bitset.Empty(), ShouldBeFalse)
			bitset.Del(20)
			for i := uint64(0); i < uint64(64); i++ {
				if i == 40 {
					So(bitset.Has(i), ShouldBeTrue)
				} else {
					So(bitset.Has(i), ShouldBeFalse)
				}
			}
			So(bitset.Empty(), ShouldBeFalse)
			bitset.Del(40)
			for i := uint64(0); i < uint64(64); i++ {
				So(bitset.Has(i), ShouldBeFalse)
			}
			So(bitset.Empty(), ShouldBeTrue)
		})
	})
}

func TestBitSet_MarshalJSON(t *testing.T) {
	Convey("bitset json 序列化", t, func() {
		bitset := NewBitSet()
		bitset.Add(1)
		bitset.Add(2)
		bitset.Add(4)
		bitset.Add(10)
		buf, err := json.Marshal(bitset)
		So(err, ShouldBeNil)
		So(string(buf), ShouldEqual, `"10000010110"`)
	})

	Convey("bitset json 反序列化", t, func() {
		bitset := NewBitSet()
		err := json.Unmarshal([]byte(`"10000010110"`), &bitset)
		So(err, ShouldBeNil)
		So(bitset.Has(1), ShouldBeTrue)
		So(bitset.Has(2), ShouldBeTrue)
		So(bitset.Has(4), ShouldBeTrue)
		So(bitset.Has(10), ShouldBeTrue)
	})
}

func BenchmarkBitSet_MarshalJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bitset := NewBitSet()
		bitset.Add(1)
		bitset.Add(2)
		bitset.Add(4)
		bitset.Add(10)
		_, _ = json.Marshal(bitset)
	}
}

func BenchmarkSetContains(b *testing.B) {
	bitset := NewBitSet()
	hashset := map[uint64]struct{}{}
	for _, i := range []uint64{1, 2, 4, 10} {
		bitset.Add(i)
		hashset[i] = struct{}{}
	}

	b.Run("bitset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for i := uint64(0); i < uint64(10); i++ {
				_ = bitset.Has(i)
			}
		}
	})

	b.Run("hashset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for i := uint64(0); i < uint64(10); i++ {
				_, _ = hashset[i]
			}
		}
	})
}
