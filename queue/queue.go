package queue

import "errors"

var (
	ErrQueueIsFull  = errors.New("queue is full")
	ErrQueueIsEmpty = errors.New("queue is empty")
)

type Queue struct {
	capacity int
	data     []int
}

func NewQueue(capacity int) *Queue {
	return &Queue{
		capacity: capacity,
		data:     make([]int, 0, capacity),
	}
}

func (q *Queue) Enqueue(val int) error {
	if len(q.data) == q.capacity {
		return ErrQueueIsFull
	}
	q.data = append(q.data, val)
	return nil
}

func (q *Queue) Dequeue() (int, error) {
	if len(q.data) == 0 {
		return 0, ErrQueueIsEmpty
	}
	val := q.data[0]
	q.data = q.data[1:]
	return val, nil
}
