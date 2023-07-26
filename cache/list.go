package cache

/*
Функция deleter для доп. логики при удалении
из листа (например удаление с диска).
*/
var customDeleter func(value interface{})

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
	len   int
	first *ListItem
	last  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.first
}

func (l *list) Back() *ListItem {
	return l.last
}

func (l *list) PushFront(v interface{}) *ListItem {
	if l.first != nil {
		l.first.Prev = &ListItem{
			Next: l.first,
		}
		l.first = l.first.Prev
	} else {
		l.first = &ListItem{}
		l.last = l.first
	}

	l.first.Value = v
	l.len++

	return l.first
}

func (l *list) PushBack(v interface{}) *ListItem {
	if l.last != nil {
		l.last.Next = &ListItem{
			Prev: l.last,
		}
		l.last = l.last.Next
	} else {
		l.last = &ListItem{}
		l.first = l.last
	}

	l.last.Value = v
	l.len++

	return l.last
}

func (l *list) Remove(i *ListItem) {
	if i != nil {
		if customDeleter != nil {
			customDeleter(i.Value.(listValue).value)
		}
		l.extractElement(i)
		l.len--
	}
}

func (l *list) MoveToFront(i *ListItem) {
	if i != l.first && i != nil {
		l.extractElement(i)
		l.first.Prev = i
		i.Next = l.first
		i.Prev = nil
		l.first = i
	}
}

func (l *list) extractElement(i *ListItem) {
	switch {
	case l.last == l.first:
		l.last = nil
		l.first = nil
	case i == l.last:
		l.last = i.Prev
		l.last.Next = nil
	case i == l.first:
		l.first = i.Next
		l.first.Prev = nil
	default:
		if i.Prev != nil {
			i.Prev.Next = i.Next
		}
		if i.Next != nil {
			i.Next.Prev = i.Prev
		}
	}
}

func NewList() List {
	return new(list)
}
