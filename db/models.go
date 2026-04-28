package db

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the shared GORM database handle for the application.
var DB *gorm.DB

// User represents a user in the system
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       string    `gorm:"unique;not null" json:"user_id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"password_hash"`
	Name         string    `gorm:"type:text" json:"name"`
	Email        string    `gorm:"type:text" json:"email"`
	IsOnline     bool      `gorm:"default:false" json:"is_online"` // Stub for online status
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Friend represents a friendship between two users
type Friend struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"not null" json:"user_id"`
	FriendID  string    `gorm:"not null" json:"friend_id"`
	IsPinned  bool      `gorm:"default:false" json:"is_pinned"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// FriendRequest stores a pending or resolved friend request.
type FriendRequest struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID string    `gorm:"not null" json:"from_user_id"`
	ToUserID   string    `gorm:"not null" json:"to_user_id"`
	Status     string    `gorm:"default:'pending';not null" json:"status"` // pending, accepted, declined
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// Room represents a room
type Room struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	RoomID                  string    `gorm:"unique;not null" json:"room_id"` // UUID
	UserID                  string    `gorm:"not null" json:"user_id"`        // Creator's user UUID
	Name                    string    `gorm:"not null" json:"name"`
	Type                    string    `gorm:"not null" json:"type"`      // public, private, secret
	Password                string    `gorm:"type:text" json:"password"` // null if not password-protected
	SharedStageLayout       string    `gorm:"type:text" json:"shared_stage_layout"`
	SharedStageLayoutLocked bool      `gorm:"default:false" json:"shared_stage_layout_locked"`
	CreatedAt               time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// RoomMember represents a member in a room
type RoomMember struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	RoomID   string    `gorm:"not null;index:idx_room_member_user,unique" json:"room_id"` // Room's UUID
	UserID   string    `gorm:"not null;index:idx_room_member_user,unique" json:"user_id"` // User's UUID
	Role     string    `gorm:"not null" json:"role"`                                      // Role in the room (admin, member, viewer)
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}

// RoomInvite stores a pending or resolved invitation to a room.
type RoomInvite struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RoomID        string    `gorm:"not null" json:"room_id"`
	InviterUserID string    `gorm:"not null" json:"inviter_user_id"`
	InvitedUserID string    `gorm:"not null" json:"invited_user_id"`
	Status        string    `gorm:"default:'pending';not null" json:"status"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// RoomVoiceParticipant represents explicit voice presence inside a room.
// Presence is room-scoped and separate from general room membership.
type RoomVoiceParticipant struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	RoomID          string    `gorm:"not null;index:idx_room_voice_user,unique" json:"room_id"`
	UserID          string    `gorm:"not null;index:idx_room_voice_user,unique" json:"user_id"`
	IsMicEnabled    bool      `gorm:"default:false" json:"is_mic_enabled"`
	IsCameraEnabled bool      `gorm:"default:false" json:"is_camera_enabled"`
	IsScreenSharing bool      `gorm:"default:false" json:"is_screen_sharing"`
	JoinedAt        time.Time `gorm:"autoCreateTime" json:"joined_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// RoomStageNote stores a shared note tile on the room stage workspace.
type RoomStageNote struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RoomID     string    `gorm:"not null;index" json:"room_id"`
	UserID     string    `gorm:"not null;index" json:"user_id"`
	X          float64   `gorm:"not null" json:"x"`
	Y          float64   `gorm:"not null" json:"y"`
	W          float64   `gorm:"not null" json:"w"`
	H          float64   `gorm:"not null" json:"h"`
	Z          float64   `gorm:"not null" json:"z"`
	LegacyText string    `gorm:"column:text;type:text;not null;default:''" json:"-"`
	Title      string    `gorm:"type:text;not null;default:''" json:"title"`
	Body       string    `gorm:"type:text;not null;default:''" json:"body"`
	BodyBold   bool      `gorm:"default:false" json:"body_bold"`
	BodyStrike bool      `gorm:"default:false" json:"body_strike"`
	BodySize   string    `gorm:"type:text;not null;default:'md'" json:"body_size"`
	Color      string    `gorm:"type:text;not null;default:'amber'" json:"color"`
	IsPinned   bool      `gorm:"default:false" json:"is_pinned"`
	IsLocked   bool      `gorm:"default:false" json:"is_locked"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// RoomStageImage stores a shared image tile on the room stage workspace.
type RoomStageImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    string    `gorm:"not null;index" json:"room_id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	X         float64   `gorm:"not null" json:"x"`
	Y         float64   `gorm:"not null" json:"y"`
	W         float64   `gorm:"not null" json:"w"`
	H         float64   `gorm:"not null" json:"h"`
	Z         float64   `gorm:"not null" json:"z"`
	Src       string    `gorm:"type:text;not null" json:"src"`
	FileName  string    `gorm:"type:text;not null;default:''" json:"file_name"`
	Caption   string    `gorm:"type:text;not null;default:''" json:"caption"`
	BgMode    string    `gorm:"column:bg_mode;type:text;not null;default:'grid'" json:"bg_mode"`
	IsPinned  bool      `gorm:"default:false" json:"is_pinned"`
	IsLocked  bool      `gorm:"default:false" json:"is_locked"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// Message stores a direct chat message between two users.
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   string    `gorm:"not null" json:"sender_id"`   // UUID отправителя
	ReceiverID string    `gorm:"not null" json:"receiver_id"` // UUID получателя
	Text       string    `gorm:"type:text" json:"text"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// InitDatabase initializes the SQLite database using GORM
func InitDatabase(path string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate all models
	err = DB.AutoMigrate(
		&User{},
		&Friend{},
		&FriendRequest{},
		&Room{},
		&RoomMember{},
		&RoomInvite{},
		&RoomVoiceParticipant{},
		&RoomStageNote{},
		&RoomStageImage{},
		&Message{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database schema:", err)
	}

	migrateRoomStageNotesLegacyText()
}

func migrateRoomStageNotesLegacyText() {
	type columnInfo struct {
		Name string `gorm:"column:name"`
	}

	var columns []columnInfo
	if err := DB.Raw("PRAGMA table_info(room_stage_notes)").Scan(&columns).Error; err != nil {
		log.Printf("Failed to inspect room_stage_notes schema: %v", err)
		return
	}

	hasLegacyText := false
	hasTitle := false
	hasBody := false
	for _, column := range columns {
		switch column.Name {
		case "text":
			hasLegacyText = true
		case "title":
			hasTitle = true
		case "body":
			hasBody = true
		}
	}

	if !hasLegacyText || !hasBody || !hasTitle {
		return
	}

	if err := DB.Exec(`
		UPDATE room_stage_notes
		SET body = CASE
			WHEN (body IS NULL OR body = '') AND text IS NOT NULL THEN text
			ELSE body
		END,
		    title = CASE
			WHEN (title IS NULL OR title = '') THEN ''
			ELSE title
		END
	`).Error; err != nil {
		log.Printf("Failed to migrate legacy room_stage_notes text into body: %v", err)
	}
}

// BeforeCreate assigns a UUID before a room is persisted.
func (r *Room) BeforeCreate(tx *gorm.DB) (err error) {
	r.RoomID = uuid.New().String()
	return
}

// BeforeCreate assigns a UUID before a user is persisted.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.UserID = uuid.New().String()
	return
}
