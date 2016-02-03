package ts

import (
	"math"
	"math/rand"
)

func main() {

}

func FixedInterval(seed int64, n uint64, c chan float64) {
	interval := float64(1.00) / float64(n)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	for idx = 0; idx < n; idx++ {
		val := float64((float64(idx) * interval))
		c <- val
	}
	close(c)
}

func SpreadInterval(seed int64, n uint64, c chan float64) {
	interval := float64(1.00) / float64(n)
	src := rand.NewSource(seed)
	// To force type (type mismatch in for if assumed)
	var idx uint64
	for idx = 0; idx < n; idx++ {
		reach := float64(src.Int63()) / float64(math.MaxInt64)
		val := float64((float64(idx) * interval) + (interval * reach))
		c <- val
	}
	close(c)
}
