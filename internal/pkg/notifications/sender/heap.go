package sender

import (
	"calendar/internal/domain/event"
)

type reminderHeap []*event.Event

func (h *reminderHeap) Len() int { return len(*h) }
func (h *reminderHeap) Less(i, j int) bool {
	return (*h)[i].Reminder.Before((*h)[j].Reminder)
}
func (h *reminderHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *reminderHeap) Push(x any) {
	*h = append(*h, x.(*event.Event))
}

func (h *reminderHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak — let GC collect the event
	*h = old[:n-1]
	return item
}
