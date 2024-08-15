package words

import (
	"fmt"
	"sync"
)

// Vec is a generic type with type parameters T and N.
type Vec[T any, N int] struct {
	elements [N]T
}

// NewVec initializes a Vec with default values.
func NewVec[T any, N int]() Vec[T, N] {
	return Vec[T, N]{}
}

// ToArray returns the internal fixed-size array of the vector.
func (v *Vec[T, N]) ToArray() [N]T {
	return v.elements
}

type ArrayPool[T any, N int] struct {
	pools    map[int]*sync.Pool
	elements [N]T
}

func (p *ArrayPool[T, N]) Get() [N]T {
	var pool *sync.Pool
	if a, ok := p.pools[N]; ok {
		pool = a
	} else {
		pool = &sync.Pool{
			New: func() interface{} {
				return make([]T, 0, N)
			},
		}
		p.pools[size] = pool
	}

	return pool.Get().([]T)[:0]
}

func (p *ArrayPool[T, N]) Put(array []T) {
	size := cap(array)
	if pool, ok := p.pools[size]; ok {
		array = array[:0]
		pool.Put(array)
		return
	}
	panic("pool not found for " + fmt.Sprint(size))
}
