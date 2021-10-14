package Util

type ArrayList struct {
	elements []interface{}
	size     int
}

func New(values ...interface{}) *ArrayList {
	list := &ArrayList{}
	list.elements = make([]interface{}, 10)
	if len(values) > 0 {
		list.Add(values...)
	}
	return list
}

func (that ArrayList) Add(values ...interface{}) {
	if that.size+len(values) >= len(that.elements)-1 {
		newElements := make([]interface{}, that.size+len(values)+1)
		copy(newElements, that.elements)
		that.elements = newElements
	}

	for _, value := range values {
		that.elements[that.size] = value
		that.size++
	}

}

func (that ArrayList) Remove(index int) interface{} {
	if index < 0 || index >= that.size {
		return nil
	}

	curEle := that.elements[index]
	that.elements[index] = nil
	copy(that.elements[index:], that.elements[index+1:that.size])
	that.size--
	return curEle
}

func (that ArrayList) Get(index int) interface{} {
	if index < 0 || index >= that.size {
		return nil
	}
	return that.elements[index]
}

func (that ArrayList) IsEmpty() bool {
	return that.size == 0
}

func (that ArrayList) Size() int {
	return that.size
}

func (that ArrayList) Contains(value interface{}) bool {
	for _, curValue := range that.elements {
		if curValue == value {
			return true
		}
	}

	return false
}
