package words

import (
	"sync"
	"testing"
)

var arrayPool = &ArrayPool[string, 5]{
	pools: make(map[int]*sync.Pool),
}

func TestPool(t *testing.T) {
	sizes := []int{8, 12, 8}

	for _, size := range sizes {
		t.Logf("Getting %d", size)
		a := arrayPool.Get(size)
		a[1] = "hello"

		t.Logf(" cap(a) %d", cap(a))
		t.Logf(" len(a) %d", len(a))
		t.Logf(" a %v", a)
		t.Logf(" len(arrayPool.pools) %d", len(arrayPool.pools))

		t.Logf("Putting %d", size)
		arrayPool.Put(a)
	}
}
