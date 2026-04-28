package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"GoCall_api/db"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

type roomStageNoteState struct {
	ID         uint    `json:"id"`
	UserID     string  `json:"user_id"`
	Username   string  `json:"username"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	W          float64 `json:"w"`
	H          float64 `json:"h"`
	Z          float64 `json:"z"`
	Title      string  `json:"title"`
	Body       string  `json:"body"`
	BodyBold   bool    `json:"body_bold"`
	BodyStrike bool    `json:"body_strike"`
	BodySize   string  `json:"body_size"`
	Color      string  `json:"color"`
	IsPinned   bool    `json:"is_pinned"`
	IsLocked   bool    `json:"is_locked"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type roomStageImageState struct {
	ID        uint    `json:"id"`
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	W         float64 `json:"w"`
	H         float64 `json:"h"`
	Z         float64 `json:"z"`
	Src       string  `json:"src"`
	FileName  string  `json:"file_name"`
	Caption   string  `json:"caption"`
	BgMode    string  `json:"bg_mode"`
	IsPinned  bool    `json:"is_pinned"`
	IsLocked  bool    `json:"is_locked"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
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

func buildRoomStageNoteState(note db.RoomStageNote) roomStageNoteState {
	username := ""
	var user db.User
	if err := db.DB.Where("user_id = ?", note.UserID).First(&user).Error; err == nil {
		username = user.Username
	}

	return roomStageNoteState{
		ID:         note.ID,
		UserID:     note.UserID,
		Username:   username,
		X:          note.X,
		Y:          note.Y,
		W:          note.W,
		H:          note.H,
		Z:          note.Z,
		Title:      note.Title,
		Body:       note.Body,
		BodyBold:   note.BodyBold,
		BodyStrike: note.BodyStrike,
		BodySize:   note.BodySize,
		Color:      note.Color,
		IsPinned:   note.IsPinned,
		IsLocked:   note.IsLocked,
		CreatedAt:  note.CreatedAt.Format(http.TimeFormat),
		UpdatedAt:  note.UpdatedAt.Format(http.TimeFormat),
	}
}

func listRoomStageNotes(roomID string) ([]roomStageNoteState, error) {
	var notes []db.RoomStageNote
	if err := db.DB.Where("room_id = ?", roomID).Order("z ASC, created_at ASC").Find(&notes).Error; err != nil {
		return nil, err
	}

	noteStates := make([]roomStageNoteState, 0, len(notes))
	for _, note := range notes {
		noteStates = append(noteStates, buildRoomStageNoteState(note))
	}

	return noteStates, nil
}

func buildRoomStageImageState(image db.RoomStageImage) roomStageImageState {
	username := ""
	var user db.User
	if err := db.DB.Where("user_id = ?", image.UserID).First(&user).Error; err == nil {
		username = user.Username
	}

	return roomStageImageState{
		ID:        image.ID,
		UserID:    image.UserID,
		Username:  username,
		X:         image.X,
		Y:         image.Y,
		W:         image.W,
		H:         image.H,
		Z:         image.Z,
		Src:       image.Src,
		FileName:  image.FileName,
		Caption:   image.Caption,
		BgMode:    image.BgMode,
		IsPinned:  image.IsPinned,
		IsLocked:  image.IsLocked,
		CreatedAt: image.CreatedAt.Format(http.TimeFormat),
		UpdatedAt: image.UpdatedAt.Format(http.TimeFormat),
	}
}

func listRoomStageImages(roomID string) ([]roomStageImageState, error) {
	var images []db.RoomStageImage
	if err := db.DB.Where("room_id = ?", roomID).Order("z ASC, created_at ASC").Find(&images).Error; err != nil {
		return nil, err
	}

	imageStates := make([]roomStageImageState, 0, len(images))
	for _, image := range images {
		imageStates = append(imageStates, buildRoomStageImageState(image))
	}

	return imageStates, nil
}

func resolveRoomStageNote(roomID, rawNoteID string) (*db.RoomStageNote, error) {
	noteID, err := strconv.ParseUint(rawNoteID, 10, 64)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var note db.RoomStageNote
	if err := db.DB.Where("room_id = ?", roomID).First(&note, uint(noteID)).Error; err != nil {
		return nil, err
	}

	return &note, nil
}

func resolveRoomStageImage(roomID, rawImageID string) (*db.RoomStageImage, error) {
	imageID, err := strconv.ParseUint(rawImageID, 10, 64)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}

	var image db.RoomStageImage
	if err := db.DB.Where("room_id = ?", roomID).First(&image, uint(imageID)).Error; err != nil {
		return nil, err
	}

	return &image, nil
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

func CreateRoomStageNote(c *gin.Context) {
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

	var req struct {
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		W          float64 `json:"w"`
		H          float64 `json:"h"`
		Z          float64 `json:"z"`
		Title      string  `json:"title"`
		Body       string  `json:"body"`
		BodyBold   bool    `json:"body_bold"`
		BodyStrike bool    `json:"body_strike"`
		BodySize   string  `json:"body_size"`
		Color      string  `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	note := db.RoomStageNote{
		RoomID:     room.RoomID,
		UserID:     currentUser.UserID,
		X:          req.X,
		Y:          req.Y,
		W:          req.W,
		H:          req.H,
		Z:          req.Z,
		LegacyText: req.Body,
		Title:      req.Title,
		Body:       req.Body,
		BodyBold:   req.BodyBold,
		BodyStrike: req.BodyStrike,
		BodySize:   req.BodySize,
		Color:      req.Color,
		IsPinned:   false,
	}
	if note.Color == "" {
		note.Color = "amber"
	}
	if err := db.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room stage note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Created room stage note",
		"note":    buildRoomStageNoteState(note),
	})
}

func UpdateRoomStageNote(c *gin.Context) {
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

	note, err := resolveRoomStageNote(room.RoomID, c.Param("noteId"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room stage note not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room stage note"})
		}
		return
	}

	var req struct {
		X          *float64 `json:"x"`
		Y          *float64 `json:"y"`
		W          *float64 `json:"w"`
		H          *float64 `json:"h"`
		Z          *float64 `json:"z"`
		Title      *string  `json:"title"`
		Body       *string  `json:"body"`
		BodyBold   *bool    `json:"body_bold"`
		BodyStrike *bool    `json:"body_strike"`
		BodySize   *string  `json:"body_size"`
		Color      *string  `json:"color"`
		IsPinned   *bool    `json:"is_pinned"`
		IsLocked   *bool    `json:"is_locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if req.X != nil {
		note.X = *req.X
	}
	if req.Y != nil {
		note.Y = *req.Y
	}
	if req.W != nil {
		note.W = *req.W
	}
	if req.H != nil {
		note.H = *req.H
	}
	if req.Z != nil {
		note.Z = *req.Z
	}
	if req.Title != nil {
		note.Title = *req.Title
	}
	if req.Body != nil {
		note.LegacyText = *req.Body
		note.Body = *req.Body
	}
	if req.BodyBold != nil {
		note.BodyBold = *req.BodyBold
	}
	if req.BodyStrike != nil {
		note.BodyStrike = *req.BodyStrike
	}
	if req.BodySize != nil {
		note.BodySize = *req.BodySize
	}
	if req.Color != nil {
		note.Color = *req.Color
	}
	if req.IsPinned != nil {
		note.IsPinned = *req.IsPinned
	}
	if req.IsLocked != nil {
		note.IsLocked = *req.IsLocked
	}

	if err := db.DB.Save(note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room stage note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated room stage note",
		"note":    buildRoomStageNoteState(*note),
	})
}

func DeleteRoomStageNote(c *gin.Context) {
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

	note, err := resolveRoomStageNote(room.RoomID, c.Param("noteId"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room stage note not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room stage note"})
		}
		return
	}

	if err := db.DB.Delete(note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room stage note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted room stage note", "note_id": note.ID})
}

func CreateRoomStageImage(c *gin.Context) {
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

	var req struct {
		X        float64 `json:"x"`
		Y        float64 `json:"y"`
		W        float64 `json:"w"`
		H        float64 `json:"h"`
		Z        float64 `json:"z"`
		Src      string  `json:"src"`
		FileName string  `json:"file_name"`
		Caption  string  `json:"caption"`
		BgMode   string  `json:"bg_mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if req.Src == "" || len(req.Src) > 12_000_000 || req.W <= 0 || req.H <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image payload"})
		return
	}

	image := db.RoomStageImage{
		RoomID:   room.RoomID,
		UserID:   currentUser.UserID,
		X:        req.X,
		Y:        req.Y,
		W:        req.W,
		H:        req.H,
		Z:        req.Z,
		Src:      req.Src,
		FileName: req.FileName,
		Caption:  req.Caption,
		BgMode:   req.BgMode,
		IsPinned: false,
	}
	if image.BgMode == "" {
		image.BgMode = "grid"
	}
	if err := db.DB.Create(&image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room stage image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Created room stage image",
		"image":   buildRoomStageImageState(image),
	})
}

func UpdateRoomStageImage(c *gin.Context) {
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

	image, err := resolveRoomStageImage(room.RoomID, c.Param("imageId"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room stage image not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room stage image"})
		}
		return
	}

	var req struct {
		X        *float64 `json:"x"`
		Y        *float64 `json:"y"`
		W        *float64 `json:"w"`
		H        *float64 `json:"h"`
		Z        *float64 `json:"z"`
		FileName *string  `json:"file_name"`
		Caption  *string  `json:"caption"`
		BgMode   *string  `json:"bg_mode"`
		IsPinned *bool    `json:"is_pinned"`
		IsLocked *bool    `json:"is_locked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if req.X != nil {
		image.X = *req.X
	}
	if req.Y != nil {
		image.Y = *req.Y
	}
	if req.W != nil {
		image.W = *req.W
	}
	if req.H != nil {
		image.H = *req.H
	}
	if req.Z != nil {
		image.Z = *req.Z
	}
	if req.FileName != nil {
		image.FileName = *req.FileName
	}
	if req.Caption != nil {
		image.Caption = *req.Caption
	}
	if req.BgMode != nil {
		image.BgMode = *req.BgMode
	}
	if req.IsPinned != nil {
		image.IsPinned = *req.IsPinned
	}
	if req.IsLocked != nil {
		image.IsLocked = *req.IsLocked
	}

	if err := db.DB.Save(image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update room stage image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Updated room stage image",
		"image":   buildRoomStageImageState(*image),
	})
}

func DeleteRoomStageImage(c *gin.Context) {
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

	image, err := resolveRoomStageImage(room.RoomID, c.Param("imageId"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room stage image not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room stage image"})
		}
		return
	}

	if err := db.DB.Delete(image).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room stage image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deleted room stage image", "image_id": image.ID})
}
