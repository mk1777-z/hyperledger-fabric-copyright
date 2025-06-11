package middle

import (
	"database/sql"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"log"
	"sort"
	"strings"
	"time"
	// "gorm.io/gorm" // 移除 GORM 导入
)

type ChatService struct{}

// NewChatService creates a new ChatService
func NewChatService() *ChatService {
	return &ChatService{}
}

// generateConversationID creates a unique and consistent ID for a conversation between two users.
func (s *ChatService) generateConversationID(userID1, userID2 string) string {
	ids := []string{userID1, userID2}
	sort.Strings(ids) // Sort to ensure consistency (userA_userB is same as userB_userA)
	return strings.Join(ids, "_")
}

// SendMessage saves a new chat message to the database using database/sql.
// It returns the saved message (with ID) or an error.
func (s *ChatService) SendMessage(senderID, receiverID, content string) (conf.ChatMessage, error) {
	if senderID == "" || receiverID == "" || content == "" {
		return conf.ChatMessage{}, fmt.Errorf("senderID, receiverID, and content cannot be empty")
	}
	if senderID == receiverID {
		return conf.ChatMessage{}, fmt.Errorf("sender and receiver cannot be the same user")
	}

	conversationID := s.generateConversationID(senderID, receiverID)
	timestamp := time.Now()

	stmt, err := conf.DB.Prepare("INSERT INTO chat_messages(conversation_id, sender_user_id, receiver_user_id, content, timestamp, is_read) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return conf.ChatMessage{}, fmt.Errorf("failed to prepare statement for sending message: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(conversationID, senderID, receiverID, content, timestamp, false)
	if err != nil {
		return conf.ChatMessage{}, fmt.Errorf("failed to execute statement for sending message: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		// Not all drivers support LastInsertId. If MySQL, it should work.
		// Log error but proceed, as message is inserted.
		log.Printf("Warning: Could not get last insert ID: %v", err)
	}

	message := conf.ChatMessage{
		ID:             uint(id), // Cast int64 to uint
		ConversationID: conversationID,
		SenderUserID:   senderID,
		ReceiverUserID: receiverID,
		Content:        content,
		Timestamp:      timestamp,
		IsRead:         false,
	}
	return message, nil
}

// GetMessages retrieves messages for a given conversationID using database/sql.
// It also marks the messages as read for the requesting user.
// userID is the ID of the user requesting the messages.
func (s *ChatService) GetMessages(conversationID string, userID string, page, pageSize int) ([]conf.ChatMessage, int64, error) {
	if conversationID == "" || userID == "" {
		return nil, 0, fmt.Errorf("conversationID and userID cannot be empty")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20 // Default page size
	}
	offset := (page - 1) * pageSize

	var messages []conf.ChatMessage
	var total int64

	// Count total messages for pagination
	row := conf.DB.QueryRow("SELECT COUNT(*) FROM chat_messages WHERE conversation_id = ?", conversationID)
	err := row.Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count messages: %w", err)
	}

	if total == 0 {
		return messages, 0, nil // No messages, return empty slice
	}

	// Get messages
	query := "SELECT id, conversation_id, sender_user_id, receiver_user_id, content, timestamp, is_read FROM chat_messages WHERE conversation_id = ? ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	rows, err := conf.DB.Query(query, conversationID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg conf.ChatMessage
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderUserID, &msg.ReceiverUserID, &msg.Content, &msg.Timestamp, &msg.IsRead); err != nil {
			log.Printf("Error scanning message row: %v", err) // Log and continue if possible
			continue
		}
		messages = append(messages, msg)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating message rows: %w", err)
	}

	// Mark messages as read for the current user in this conversation
	stmt, err := conf.DB.Prepare("UPDATE chat_messages SET is_read = ? WHERE conversation_id = ? AND receiver_user_id = ? AND is_read = ?")
	if err != nil {
		log.Printf("Warning: failed to prepare statement for marking messages as read: %v\n", err)
	} else {
		defer stmt.Close()
		_, updateErr := stmt.Exec(true, conversationID, userID, false)
		if updateErr != nil {
			fmt.Printf("Warning: failed to mark messages as read for conversation %s, user %s: %v\n", conversationID, userID, updateErr)
		}
	}

	// Reverse messages to display in chronological order (oldest first)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, nil
}

// GetConversations retrieves a list of conversations for a given userID using database/sql.
func (s *ChatService) GetConversations(userID string) ([]conf.Conversation, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	var conversations []conf.Conversation

	// Step 1: Find all unique conversation IDs involving the user
	// This query gets the latest message for each conversation_id involving the user.
	// It's a bit complex, might be optimizable or done in multiple steps for clarity.
	query := `
		SELECT
			m.conversation_id,
			m.sender_user_id,
			m.receiver_user_id,
			m.content,
			m.timestamp
		FROM chat_messages m
		INNER JOIN (
			SELECT conversation_id, MAX(timestamp) AS max_timestamp
			FROM chat_messages
			WHERE sender_user_id = ? OR receiver_user_id = ?
			GROUP BY conversation_id
		) lm ON m.conversation_id = lm.conversation_id AND m.timestamp = lm.max_timestamp
		WHERE m.sender_user_id = ? OR m.receiver_user_id = ?
		ORDER BY m.timestamp DESC;
	`
	rows, err := conf.DB.Query(query, userID, userID, userID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []conf.Conversation{}, nil
		}
		return nil, fmt.Errorf("failed to get latest messages for conversations: %w", err)
	}
	defer rows.Close()

	processedConvIDs := make(map[string]bool) // To handle cases where the join might return duplicates if timestamps are identical for different messages in a conversation (unlikely with precise timestamps but good for safety)

	for rows.Next() {
		var convID, senderID, receiverID, lastMsgContent string
		var lastMsgTimestamp time.Time

		if err := rows.Scan(&convID, &senderID, &receiverID, &lastMsgContent, &lastMsgTimestamp); err != nil {
			log.Printf("Error scanning conversation row: %v", err)
			continue
		}

		if processedConvIDs[convID] {
			continue
		}
		processedConvIDs[convID] = true

		otherUserID := ""
		if senderID == userID {
			otherUserID = receiverID
		} else {
			otherUserID = senderID
		}

		var unreadCount int64
		unreadQuery := "SELECT COUNT(*) FROM chat_messages WHERE conversation_id = ? AND receiver_user_id = ? AND is_read = ?"
		unreadRow := conf.DB.QueryRow(unreadQuery, convID, userID, false)
		if err := unreadRow.Scan(&unreadCount); err != nil {
			log.Printf("Failed to count unread messages for conversation %s: %v", convID, err)
			// Continue with unreadCount as 0 if error
		}

		// Placeholder for OtherUsername. You'll need to query your users table/profile table
		// based on otherUserID to get the actual username.
		otherUsername := otherUserID // Replace with actual username lookup

		conversations = append(conversations, conf.Conversation{
			ConversationID:       convID,
			OtherUserID:          otherUserID,
			OtherUsername:        otherUsername,
			LastMessage:          lastMsgContent,
			LastMessageTimestamp: lastMsgTimestamp,
			UnreadCount:          int(unreadCount),
		})
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating conversation rows: %w", err)
	}

	// Sorting is already handled by the SQL query (ORDER BY m.timestamp DESC)

	return conversations, nil
}

// MarkMessagesAsRead marks all messages in a conversation as read for a specific user using database/sql.
func (s *ChatService) MarkMessagesAsRead(conversationID string, readerUserID string) error {
	if conversationID == "" || readerUserID == "" {
		return fmt.Errorf("conversationID and readerUserID cannot be empty")
	}

	stmt, err := conf.DB.Prepare("UPDATE chat_messages SET is_read = ? WHERE conversation_id = ? AND receiver_user_id = ? AND is_read = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement for marking messages as read: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(true, conversationID, readerUserID, false)
	if err != nil {
		return fmt.Errorf("failed to execute statement for marking messages as read: %w", err)
	}
	return nil
}

// Helper function to get username (if you have a separate users table with more details)
// You would implement this to fetch username from your 'users' or 'user_profiles' table.
// func (s *ChatService) getUsername(userID string) (string, error) {
// 	var username string
// 	query := "SELECT username FROM users WHERE id = ?" // Adjust table and column names
// 	row := conf.DB.QueryRow(query, userID)
// 	if err := row.Scan(&username); err != nil {
// 		if err == sql.ErrNoRows {
// 			return userID, fmt.Errorf("user not found: %s", userID) // Or return userID as fallback
// 		}
// 		return userID, err
// 	}
// 	return username, nil
// }
