package dsa

type Queue[T any] struct {
	items []T
}

// NewQueue 创建一个新的队列
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{}
}

// Enqueue 向队列中添加元素
func (q *Queue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

// Dequeue 从队列中移除并返回一个元素
func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.items) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	item := q.items[0]
	q.items = q.items[1:]
	return item, true
}

// Front 返回队列的第一个元素
func (q *Queue[T]) Front() (T, bool) {
	if len(q.items) == 0 {
		var zeroValue T
		return zeroValue, false
	}
	return q.items[0], true
}

// Size 返回队列的大小
func (q *Queue[T]) Size() int {
	return len(q.items)
}

// IsEmpty 检查队列是否为空
func (q *Queue[T]) IsEmpty() bool {
	return len(q.items) == 0
}
