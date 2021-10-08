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
	array.head = 0
	array.tail = 0
	return array
}

func (c *CircleQueue) mod(i int) int {
	if i < c.max {
		return i
	}

	return i - c.max
}

func (c *CircleQueue) Size() int {
	return c.mod(c.max + c.tail - c.head)
}

func (c *CircleQueue) IsEmpty() bool {
	return c.tail == c.head
}

func (c *CircleQueue) IsFull() bool {
	return c.mod(c.tail+1) == c.head
}

func (c *CircleQueue) Push(val interface{}, cover bool) bool {
	if c.IsFull() {
		if cover {
			c.head = c.mod(c.head + 1)
			return true

		} else {
			return false
		}
	}

	c.array[c.tail] = val
	c.tail = c.mod(c.tail + 1)
	return true
}

func (c *CircleQueue) Pop() (interface{}, bool) {
	if c.IsEmpty() {
		return nil, false
	}

	head := c.head
	v := c.array[head]
	c.head = c.mod(head + 1)
	c.array[head] = nil
	return v, true
}

func (c *CircleQueue) Get(idx int) (interface{}, bool) {
	if idx < 0 || idx >= c.Size() {
		return nil, false
	}

	idx = c.mod(c.head + idx)
	return c.array[idx], true
}

func (c *CircleQueue) Set(idx int, val interface{}) bool {
	if idx < 0 || idx >= c.Size() {
		return false
	}

	idx = c.mod(c.head + idx)
	c.array[idx] = val
	return true
}

func (c *CircleQueue) Clear() {
	c.head = 0
	c.tail = 0
	for i := c.max - 1; i >= 0; i-- {
		c.array[i] = nil
	}
}

func (c *CircleQueue) Remove(val interface{}) bool {
	if c.IsEmpty() {
		return false
	}

	head := c.head
	v := c.array[head]
	if v == val {
		c.head = c.mod(head + 1)
		c.array[head] = nil
		return true
	}

	return false
}
