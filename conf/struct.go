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

// 推荐系统的ORM
// 不要直接放数据库，用api,不然推荐系统可能不会刷新
// type Items struct {
// 	ItemId     string    `gorm:"primaryKey;type:varchar(256);not null;column:item_id"`
// 	IsHidden   bool      `gorm:"type:tinyint(1);not null;default:0;column:is_hidden"`
// 	Categories string    `gorm:"type:json;not null;column:categories"`
// 	TimeStamp  time.Time `gorm:"type:datetime;not null;column:time_stamp"`
// 	Labels     string    `gorm:"type:json;not null;column:labels"`
// 	Comment    string    `gorm:"type:text;not null;column:comment"`
// }

// type Feedback struct {
// 	FeedbackType string    `gorm:"primaryKey;type:varchar(256);not null;column:feedback_type"`
// 	UserId       string    `gorm:"primaryKey;type:varchar(256);not null;column:user_id"`
// 	ItemId       string    `gorm:"primaryKey;type:varchar(256);not null;column:item_id"`
// 	TimeStamp    time.Time `gorm:"type:datetime;not null;column:time_stamp"`
// 	Comment      string    `gorm:"type:text;not null;column:comment"`
// }

// type Users struct {
// 	UserId    string `gorm:"primaryKey;column:user_id;type:varchar(256);not null"`
// 	Labels    string `gorm:"type:json;not null"`
// 	Subscribe string `gorm:"type:json;not null"`
// 	Comment   string `gorm:"type:text;not null"`
// }

type Item struct {
	Id        int       `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar(50);not null"`
	Owner     string    `gorm:"column:owner;type:varchar(255);not null"`
	SimpleDsc string    `gorm:"column:simple_dsc;type:varchar(30);default:null"`
	Dsc       string    `gorm:"column:dsc;type:text;default:null"`
	Price     int       `gorm:"column:price;type:int;not null"`
	Img       []byte    `gorm:"column:img;type:longblob;default:null"`
	OnSale    bool      `gorm:"column:on_sale;type:tinyint(1);default:0"`
	StartTime time.Time `gorm:"column:start_time;type:date;default:null"`
	TransID   string    `gorm:"column:transID;type:text;default:null"`
	Category  string    `gorm:"column:category;type:varchar(50);default:'其他'"`
	Decision  string    `gorm:"column:decision;type:varchar(255);default:null"`
}

func (Item) TableName() string {
	return "item"
}

type DbUser struct {
	Username          string    `gorm:"primaryKey;column:username;type:varchar(255);not null"`
	Password          string    `gorm:"column:password;type:varchar(255);default:NULL"`
	Location          string    `gorm:"column:location;type:varchar(255);default:'未知地点'"`
	Last_active_time  time.Time `gorm:"column:last_active_time;type:datetime;default:CURRENT_TIMESTAMP"`
	Registration_time time.Time `gorm:"column:registration_time;type:datetime;default:NULL"`
}

func (DbUser) TableName() string {
	return "user"
}

type Favorites struct {
	Id         int       `gorm:"primaryKey;autoIncrement;column:id"`
	Username   string    `gorm:"type:varchar(255);not null;index;column:username"`
	ItemId     int       `gorm:"type:int;not null;index;column:item_id"`
	CreateTime time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP;autoCreateTime;column:create_time"`
}

func (Favorites) TableName() string {
	return "favorites"
}
