package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	front *ListItem
	back  *ListItem
	len   int
}

func NewList() List {
	return &list{}
}

func (lst *list) Len() int {
	return lst.len
}

func (lst *list) Front() *ListItem {
	return lst.front
}

func (lst *list) Back() *ListItem {
	return lst.back
}

func (lst *list) PushFront(v interface{}) (item *ListItem) {
	item = &ListItem{Value: v}
	defer func() {
		lst.len++
	}()
	if lst.len == 0 {
		lst.front = item
		lst.back = item
		return
	}
	item.Next = lst.front
	lst.front.Prev = item
	lst.front = item
	return
}

func (lst *list) PushBack(val interface{}) (item *ListItem) {
	item = &ListItem{Value: val}
	defer func() {
		lst.len++
	}()
	if lst.len == 0 {
		lst.front = item
		lst.back = item
		return
	}
	item.Prev = lst.back
	lst.back.Next = item
	lst.back = item
	return
}

func (lst *list) Remove(item *ListItem) {
	if item.Prev != nil {
		item.Prev.Next = item.Next
	} else {
		lst.front = item.Next
	}

	if item.Next != nil {
		item.Next.Prev = item.Prev
	} else {
		lst.back = item.Prev
	}

	lst.len--
}

func (lst *list) MoveToFront(item *ListItem) {
	if item == lst.front {
		return
	}

	if item == lst.back {
		lst.back = item.Prev
	}

	item.Prev.Next = item.Next
	if item.Next != nil {
		item.Next.Prev = item.Prev
	}

	item.Prev = nil
	item.Next = lst.front
	lst.front.Prev = item
	lst.front = item
}
