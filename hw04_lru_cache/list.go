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

// структура, которая реализует интерфейс List.
type list struct {
	head   *ListItem
	tail   *ListItem
	length int
}

func NewList() List {
	return new(list)
}

// длина списка.
func (l *list) Len() int {
	return l.length
}

// первый элемент списка.
func (l *list) Front() *ListItem {
	return l.head
}

// последний элемент списка.
func (l *list) Back() *ListItem {
	return l.tail
}

// добавить значение в начало.
func (l *list) PushFront(v interface{}) *ListItem {
	// адрес нового элемента.
	var newItem *ListItem

	if l.length == 0 {
		newItem = &ListItem{
			Value: v,
			Next:  nil,
			Prev:  nil,
		}
		l.tail = newItem
	} else {
		newItem = &ListItem{
			Value: v,
			Next:  l.head,
			Prev:  nil,
		}
		l.head.Prev = newItem
	}
	l.head = newItem
	l.length++

	return newItem
}

// добавить значение в конец.
func (l *list) PushBack(v interface{}) *ListItem {
	// адрес нового узла.
	var newItem *ListItem

	if l.length == 0 {
		newItem = &ListItem{
			Value: v,
			Next:  nil,
			Prev:  nil,
		}
		l.head = newItem
	} else {
		newItem = &ListItem{
			Value: v,
			Next:  nil,
			Prev:  l.tail,
		}
		l.tail.Next = newItem
	}
	l.tail = newItem
	l.length++

	return newItem
}

// удалить элемент.
func (l *list) Remove(i *ListItem) {
	switch {
	case l.length == 1:
		l.head = nil
		l.tail = nil
	case l.head == i:
		l.head = i.Next
		l.head.Prev = nil
	case l.tail == i:
		l.tail = i.Prev
		l.tail.Next = nil
	case l.length > 2:
		i.Prev.Next = i.Next
		i.Next.Prev = i.Prev
	}
	// удаляем старые ссылки, больше не нужны.
	i.Prev = nil
	i.Next = nil

	l.length--
}

// переместить элемент в начало.
func (l *list) MoveToFront(i *ListItem) {
	// первый элемент - ничего не делаем.
	if l.head == i {
		return
	}

	if l.tail == i {
		l.tail = i.Prev
		l.tail.Next = nil
	} else {
		// элемент "посередине" перепривязали соседей друг к другу.
		i.Next.Prev = i.Prev
		i.Prev.Next = i.Next
	}

	// перемещение в начало списка.
	i.Prev = nil
	i.Next = l.head
	l.head.Prev = i
	l.head = i
}
