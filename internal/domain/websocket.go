package domain

// ZoneOccupancyUpdate represents a real-time event for zone occupancy changes.
type ZoneOccupancyUpdate struct {
	ZoneID         int   `json:"zone_id"`
	OccupancyCount int64 `json:"occupancy_count"`
}

// WSHub defines the interface for the WebSocket hub to broadcast messages.
type WSHub interface {
	BroadcastZoneUpdate(update ZoneOccupancyUpdate)
}
