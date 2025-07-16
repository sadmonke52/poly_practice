package stress_tester

import (
	"fmt"
	"math/rand"
	"strconv"
)

// какая то реальная логика в продакшене
func SimulateHeavyGCPollution() {
	size := 1<<20 + rand.Intn(3<<20)
	_ = make([]byte, size)

	for i := 0; i < 1000; i++ {
		_ = fmt.Sprintf("%d-%d-%d", rand.Int63(), rand.Int63(), rand.Int63())
	}

	m := make(map[string][]byte, 1000)
	for i := 0; i < 10000; i++ {
		m[strconv.Itoa(i)] = make([]byte, 8<<10)
	}
}
