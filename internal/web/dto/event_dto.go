package dto

type EventRequestCreate struct {
	EventName    string `json:"event_name" binding:"required" example:"Встреча с командой"`
	Date         string `json:"date" binding:"required" example:"2026-03-01"`
	EventText    string `json:"event" example:"Обсуждение планов на квартал"`
	ReminderTime string `json:"reminder_time" example:"2026-03-01T09:00"`
}
type EventRequestUpdate struct {
	EventID      string `json:"event_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	EventName    string `json:"event_name" binding:"required" example:"Встреча с командой"`
	Date         string `json:"date" binding:"required" example:"2026-03-01"`
	EventText    string `json:"event" example:"Обсуждение планов на квартал"`
	ReminderTime string `json:"reminder_time" example:"2026-03-01T09:00"`
}

type EventRequestDelete struct {
	EventID string `json:"event_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
