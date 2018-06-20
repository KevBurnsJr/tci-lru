package lru

import (
	"fmt"
	"math/rand"
	"testing"
)


func noTags(i int) []string {
	return []string{}
}

func oneTag(i int) []string {
	return []string{"tag-1"}
}

func largeTagSpace(i int) []string {
	return []string{
		fmt.Sprintf("tag-%d", i),
	}
}

func smallTagSpace10(n int) (tags []string) {
	for i := 0; i < 10; i++ {
		tags = append(tags, fmt.Sprintf("tag-%d", int(n)%100+i))
	}
	return tags
}

func largeTagSpace10(n int) (tags []string) {
	for i := 0; i < 10; i++ {
		tags = append(tags, fmt.Sprintf("tag-%d", int(n)+i))
	}
	return tags
}

func smallTagSpace100(n int) (tags []string) {
	for i := 0; i < 100; i++ {
		tags = append(tags, fmt.Sprintf("tag-%d", int(n)%1000+i))
	}
	return tags
}

func largeTagSpace100(n int) (tags []string) {
	for i := 0; i < 100; i++ {
		tags = append(tags, fmt.Sprintf("tag-%d", int(n)+i))
	}
	return tags
}

// Mixed add/get benchmarks (for comparison with golang-lru)
func BenchmarkTCI_Rand_NoTags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, noTags)
}

func BenchmarkTCI_Freq_NoTags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, noTags)
}

func BenchmarkTCI_Rand_OneTag(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, oneTag)
}

func BenchmarkTCI_Freq_OneTag(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, oneTag)
}

func BenchmarkTCI_Rand_LargeTagSpace(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace)
}

func BenchmarkTCI_Freq_LargeTagSpace(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace)
}

func BenchmarkTCI_Rand_SmallTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, smallTagSpace10)
}

func BenchmarkTCI_Freq_SmallTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, smallTagSpace10)
}

func BenchmarkTCI_Rand_LargeTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace10)
}

func BenchmarkTCI_Freq_LargeTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace10)
}

func BenchmarkTCI_Rand_SmallTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, smallTagSpace100)
}

func BenchmarkTCI_Freq_SmallTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, smallTagSpace100)
}

func BenchmarkTCI_Rand_LargeTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace100)
}

func BenchmarkTCI_Freq_LargeTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_TagFunc(b, largeTagSpace100)
}

// Get benchmark
func BenchmarkTCI_Rand_Get(b *testing.B) {
	benchmarkTCI_Rand_Get_TagFunc(b, noTags)
}

// Add benchmarks
func BenchmarkTCI_Rand_Add_NoTags(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, noTags)
}

func BenchmarkTCI_Rand_Add_OneTag(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, oneTag)
}

func BenchmarkTCI_Rand_Add_LargeTagSpace(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, largeTagSpace)
}

func BenchmarkTCI_Rand_Add_SmallTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, smallTagSpace10)
}

func BenchmarkTCI_Rand_Add_LargeTagSpace_10Tags(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, largeTagSpace10)
}

func BenchmarkTCI_Rand_Add_SmallTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, smallTagSpace100)
}

func BenchmarkTCI_Rand_Add_LargeTagSpace_100Tags(b *testing.B) {
	benchmarkTCI_Rand_Add_TagFunc(b, largeTagSpace100)
}

func benchmarkTCI_Rand_Add_TagFunc(b *testing.B, makeTags func(int) []string) {
	l, err := NewLRU(8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N)
	for i := 0; i < b.N; i++ {
		trace[i] = rand.Int63() % 32768
	}

	tags := make(map[int64][]string)
	for _, n := range trace {
		tags[n] = makeTags(int(n))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i], tags[trace[i]]...)
	}
}

func benchmarkTCI_Rand_Get_TagFunc(b *testing.B, makeTags func(int) []string) {
	l, err := NewLRU(8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N)
	for i := 0; i < b.N; i++ {
		trace[i] = rand.Int63() % 32768
	}

	tags := make(map[int64][]string)
	for _, n := range trace {
		tags[n] = makeTags(int(n))
	}

	var hit, miss int
	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i], tags[trace[i]]...)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}

	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func benchmarkTCI_Rand_TagFunc(b *testing.B, makeTags func(int) []string) {
	l, err := NewLRU(8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	tags := make(map[int64][]string)
	for _, n := range trace {
		tags[n] = makeTags(int(n))
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Add(trace[i], trace[i], tags[trace[i]]...)
		} else {
			_, ok := l.Get(trace[i])
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func benchmarkTCI_Freq_TagFunc(b *testing.B, makeTags func(int) []string) {
	l, err := NewLRU(8192, nil)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	tags := make(map[int64][]string)
	for _, n := range trace {
		tags[n] = makeTags(int(n))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i], tags[trace[i]]...)
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}
