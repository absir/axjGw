package Util

type CircleQueue struct {
	array []interface{}
	max   int
	head  int
	tail  int
}

func NewCircleQueue(max int) *CircleQueue {
	array := new(CircleQueue)
	array.array = make([]interface{}, max)
	array.max = max
	//array.head = 0
	//array.tail = 0
	return array
}

func (that CircleQueue) mod(i int) int {
	if i < that.max {
		return i
	}

	return i - that.max
}

func (that CircleQueue) Size() int {
	return that.mod(that.max + that.tail - that.head)
}

func (that CircleQueue) IsEmpty() bool {
	return that.tail == that.head
}

func (that CircleQueue) IsFull() bool {
	return that.mod(that.tail+1) == that.head
}

func (that CircleQueue) Push(val interface{}, cover bool) bool {
	if that.IsFull() {
		if cover {
			that.head = that.mod(that.head + 1)
			return true

		} else {
			return false
		}
	}

	that.array[that.tail] = val
	that.tail = that.mod(that.tail + 1)
	return true
}

func (that CircleQueue) Pop() (interface{}, bool) {
	if that.IsEmpty() {
		return nil, false
	}

	head := that.head
	v := that.array[head]
	that.head = that.mod(head + 1)
	that.array[head] = nil
	return v, true
}

func (that CircleQueue) Get(idx int) (interface{}, bool) {
	if idx < 0 || idx >= that.Size() {
		return nil, false
	}

	idx = that.mod(that.head + idx)
	return that.array[idx], true
}

func (that CircleQueue) Set(idx int, val interface{}) bool {
	if idx < 0 || idx >= that.Size() {
		return false
	}

	idx = that.mod(that.head + idx)
	that.array[idx] = val
	return true
}

func (that CircleQueue) Clear() {
	that.head = 0
	that.tail = 0
	for i := that.max - 1; i >= 0; i-- {
		that.array[i] = nil
	}
}

func (that CircleQueue) Remove(val interface{}) bool {
	if that.IsEmpty() {
		return false
	}

	head := that.head
	v := that.array[head]
	if v == val {
		that.head = that.mod(head + 1)
		that.array[head] = nil
		return true
	}

	return false
}
