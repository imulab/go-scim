package shared

type Queue interface {
	Offer(item interface{})
	Poll() interface{}
	Peek() interface{}
	Size() int
	Clone() Queue
	Capacity() int
}

type queue struct {
	data []interface{}
}

func NewQueue(cap int) *queue {
	return &queue{data: make([]interface{}, 0, cap)}
}

func NewQueueWithoutLimit() *queue {
	return &queue{data: make([]interface{}, 0)}
}

func (q *queue) Offer(item interface{}) {
	q.data = append(q.data, item)
}

func (q *queue) Clone() Queue {
	q0 := NewQueueWithoutLimit()
	for _, elem := range q.data {
		q0.data = append(q0.data, elem)
	}
	return q0
}

func (q *queue) Poll() interface{} {
	if q.Size() == 0 {
		return nil
	}
	item := q.data[0]
	q.data = q.data[1:]
	return item
}

func (q *queue) Peek() interface{} {
	if q.Size() == 0 {
		return nil
	}
	return q.data[0]
}

func (q *queue) Size() int {
	return len(q.data)
}

func (q *queue) Capacity() int {
	return cap(q.data)
}
