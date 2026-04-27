package handlers

import (
	"encoding/json"
	"net/http"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
)

type roomStageTileLayout struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
	Z float64 `json:"z"`
}

type roomStageLayoutPayload struct {
	StageLayouts  map[string]roomStageTileLayout `json:"stage_layouts"`
	PinnedTileIDs []string                       `json:"pinned_tile_ids"`
}

type roomStageLayoutState struct {
	StageLayouts  map[string]roomStageTileLayout `json:"stage_layouts"`
	PinnedTileIDs []string                       `json:"pinned_tile_ids"`
	SharedLocked  bool                           `json:"shared_locked"`
	CanEditShared bool                           `json:"can_edit_shared"`
}

func normalizeRoomStageLayoutPayload(payload roomStageLayoutPayload) roomStageLayoutPayload {
	if payload.StageLayouts == nil {
		payload.StageLayouts = map[string]roomStageTileLayout{}
	}
	if payload.PinnedTileIDs == nil {
		payload.PinnedTileIDs = []string{}
	}
	return payload
}

func parseRoomSharedStageLayout(raw string) roomStageLayoutPayload {
	if raw == "" {
		return normalizeRoomStageLayoutPayload(roomStageLayoutPayload{})
	}

	var payload roomStageLayoutPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return normalizeRoomStageLayoutPayload(roomStageLayoutPayload{})
	}

	return normalizeRoomStageLayoutPayload(payload)
}

func canEditSharedStageLayout(member *db.RoomMember, room *db.Room) bool {
	if member == nil {
		return false
	}

	if !room.SharedStageLayoutLocked {
		return true
	}

	return member.Role == "creator"
}

func UpdateRoomSharedStageLayout(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	member, err := ensureRoomMember(room, currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a room member"})
		return
	}
	if !canEditSharedStageLayout(member, room) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Shared stage layout is locked by the room owner"})
		return
	}

	var req roomStageLayoutPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	req = normalizeRoomStageLayoutPayload(req)

	serialized, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode shared stage layout"})
		return
	}

	room.SharedStageLayout = string(serialized)
	if err := db.DB.Save(room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shared stage layout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated shared stage layout",
		"stage_layout": roomStageLayoutState{
			StageLayouts:  req.StageLayouts,
			PinnedTileIDs: req.PinnedTileIDs,
			SharedLocked:  room.SharedStageLayoutLocked,
			CanEditShared: canEditSharedStageLayout(member, room),
		},
	})
}

func UpdateRoomSharedStageLayoutLock(c *gin.Context) {
	currentUser, ok := getAuthenticatedDBUser(c)
	if !ok {
		return
	}

	room, err := resolveRoomByParam(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	member, err := ensureRoomMember(room, currentUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify room membership"})
		return
	}
	if member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a room member"})
		return
	}
	if member.Role != "creator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the room owner can change shared stage lock"})
		return
	}

	var req struct {
		SharedLocked bool `json:"shared_locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	room.SharedStageLayoutLocked = req.SharedLocked
	if err := db.DB.Save(room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shared stage lock"})
		return
	}

	payload := parseRoomSharedStageLayout(room.SharedStageLayout)
	c.JSON(http.StatusOK, gin.H{
		"message": "Updated shared stage lock",
		"stage_layout": roomStageLayoutState{
			StageLayouts:  payload.StageLayouts,
			PinnedTileIDs: payload.PinnedTileIDs,
			SharedLocked:  room.SharedStageLayoutLocked,
			CanEditShared: canEditSharedStageLayout(member, room),
		},
	})
}
