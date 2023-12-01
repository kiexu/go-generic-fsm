package fsm

type visited[T, S comparable, U, V any] struct {
	vCounter []int8
	eCounter map[*Edge[T, S, U, V]]int8
}

func newVisited[T, S comparable, U, V any](len int) *visited[T, S, U, V] {
	return &visited[T, S, U, V]{
		vCounter: make([]int8, len),
		eCounter: make(map[*Edge[T, S, U, V]]int8, len),
	}
}

func (v *visited[T, S, U, V]) vIncr(idx int) {
	v.vCounter[idx] += 1
}

func (v *visited[T, S, U, V]) vDecr(idx int) {
	v.vCounter[idx] -= 1
}

func (v *visited[T, S, U, V]) vCnt(idx int) int8 {
	return v.vCounter[idx]
}

func (v *visited[T, S, U, V]) eIncr(e *Edge[T, S, U, V]) {
	if e == nil {
		return
	}
	v.eCounter[e] += 1
}

func (v *visited[T, S, U, V]) eDecr(e *Edge[T, S, U, V]) {
	if e == nil {
		return
	}
	v.eCounter[e] -= 1
}

func (v *visited[T, S, U, V]) eCnt(e *Edge[T, S, U, V]) int8 {
	if e == nil {
		return 0
	}
	return v.eCounter[e]
}
