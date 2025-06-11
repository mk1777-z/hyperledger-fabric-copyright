package conf

import (
	"time" //确保导入 time 包

	"github.com/dgrijalva/jwt-go"
)

type Mysql struct {
	DbUser     string `yaml:"dbUser"`
	DbPassword string `yaml:"dbPassword"`
	DbName     string `yaml:"dbName"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	Mysql  Mysql  `yaml:"mysql"`
	Jwtkey []byte `yaml:"jwtkey"`
	APIKey string `yaml:"APIKey"`
}

type Upload struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Simple_dsc string  `json:"simple_dsc"`
	Dsc        string  `json:"dsc"`
	Price      float32 `json:"price"`
	Img        string  `json:"img"`
	On_sale    bool    `json:"on_sale"`
	Category   int     `json:"category"`
}

type UpdateItem struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float32 `json:"price"`
	Dsc         string  `json:"dsc"`
	Sale        bool    `json:"on_sale"`
}

type Createtrans struct {
	ID        string
	Name      string
	Seller    string
	Purchaser string
	Price     float64
	Transtime string
}

type Asset struct {
	Owner   string
	Balance float64
}

// UserClaims 用于 JWT 的声明

type UserClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type CopyrightItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SimpleDsc string `json:"simple_dsc"`
	Owner     string `json:"owner"`
	Price     string `json:"price"`
	Img       string `json:"img"`
}

type AuditRecord struct {
	TradeID   string `json:"tradeId"`
	Decision  string `json:"decision"` // APPROVE/REJECT
	Comment   string `json:"comment"`
	Regulator string `json:"regulator"`
	Timestamp int64  `json:"timestamp"`
}

// ChatMessage represents a message in a conversation
type ChatMessage struct {
	ID             uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ConversationID string    `json:"conversation_id" gorm:"index"` // Unique ID for a conversation pair
	SenderUserID   string    `json:"sender_user_id" gorm:"index"`
	ReceiverUserID string    `json:"receiver_user_id" gorm:"index"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	IsRead         bool      `json:"is_read" gorm:"default:false"`
}

// Conversation represents an overview of a chat for the conversation list
type Conversation struct {
	ConversationID       string    `json:"conversation_id"`
	OtherUserID          string    `json:"other_user_id"`  // The user the current user is chatting with
	OtherUsername        string    `json:"other_username"` // Display name of the other user
	LastMessage          string    `json:"last_message"`   // Snippet of the last message
	LastMessageTimestamp time.Time `json:"last_message_timestamp"`
	UnreadCount          int       `json:"unread_count"` // Unread messages for the current user in this conversation
	// Consider adding OtherUserProfilePic string `json:"other_user_profile_pic"` if you implement user profiles
}
