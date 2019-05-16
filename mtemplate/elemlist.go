package mtemplate

type elemlist struct {
	elements []interface{}
}

// NewElemlist creates and returns a new elemlist instance
func NewElemlist() *elemlist {
	newElemList := new(elemlist)
	newElemList.elements = make([]interface{}, 0)
	return newElemList
}

// Push adds a new item to an elemlist
func (list *elemlist) Push(item interface{}) {
	list.elements = append(list.elements, item)
}

// At returns at tiem stored in the elemlist at the given position
func (list elemlist) At(index int) interface{} {
	var foundItem interface{}

	if index < len(list.elements) {
		foundItem = list.elements[index]
	}

	return foundItem
}

// Len returns the length of the elemlist
func (list elemlist) Len() int {
	return len(list.elements)
}
