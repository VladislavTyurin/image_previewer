package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("myTest", func(t *testing.T) {
		l := NewList()
		l.MoveToFront(l.Back())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
		l.Remove(l.Front())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		l.PushFront(50)
		require.Equal(t, l.Front(), l.Back())
		require.NotNil(t, l.Back())
		require.NotNil(t, l.Front())

		l.Remove(l.Front())
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())

		l.PushBack(60)
		require.Equal(t, l.Back(), l.Front())

		l.PushBack(40)
		l.Remove(l.Front())
		require.NotNil(t, l.Front())
		require.Equal(t, 40, l.Front().Value)
		require.Equal(t, l.Front(), l.Back())

		l.Remove(l.Back())
		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("moveFromBackToFront", func(t *testing.T) {
		l := NewList()
		l.PushFront(1) // [1]
		l.PushFront(2) // [2 1]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{2, 1}, elems)

		l.MoveToFront(l.Back()) // [1 2]

		elems = make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{1, 2}, elems)
		require.Equal(t, 2, l.Back().Value)
	})

	t.Run("removeLast", func(t *testing.T) { //nolint:dupl
		l := NewList()
		l.PushFront(1) // [1]
		l.PushFront(2) // [2 1]
		l.PushFront(3) // [3 2 1]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{3, 2, 1}, elems)

		l.Remove(l.Back()) // [3 2]

		elems = make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{3, 2}, elems)
		require.Equal(t, 2, l.Back().Value)
		require.Equal(t, 3, l.Front().Value)
	})

	t.Run("removeFirst", func(t *testing.T) { //nolint:dupl
		l := NewList()
		l.PushFront(1) // [1]
		l.PushFront(2) // [2 1]
		l.PushFront(3) // [3 2 1]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{3, 2, 1}, elems)

		l.Remove(l.Front()) // [2 1]

		elems = make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{2, 1}, elems)
		require.Equal(t, 2, l.Front().Value)
		require.Equal(t, 1, l.Back().Value)
	})
}
