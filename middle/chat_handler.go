package middle

import (
	"context" // 确保导入 context
	"hyperledger-fabric-copyright/conf"
	"log"
	"net/http"
	"strconv"
	"strings" // 确保导入 strings

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/dgrijalva/jwt-go"
)

var chatService = NewChatService() // Initialize the service

// SendMessageHandler handles sending a new chat message.
// POST /api/chat/send
// Request Body: {"receiver_user_id": "user_b", "content": "Hello!"}
func SendMessageHandler(_ context.Context, c *app.RequestContext) {
	log.Printf("=== SendMessageHandler 开始处理请求 ===")

	// JWT认证 - 与其他API保持一致
	tokenString := c.GetHeader("Authorization")
	log.Printf("Authorization header: %s", string(tokenString))

	if string(tokenString) == "" {
		log.Printf("错误: Authorization token 缺失")
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)
	log.Printf("提取的token: %s", token_String)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		log.Printf("JWT解析错误: %v", err)
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		log.Printf("JWT验证失败: ok=%v, valid=%v", ok, token.Valid)
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}

	senderID := claims.Username
	log.Printf("认证成功，发送者: %s", senderID)

	var req struct {
		ReceiverUserID string `json:"receiver_user_id"`
		Content        string `json:"content"`
	}

	if err := c.BindJSON(&req); err != nil {
		log.Printf("请求体解析错误: %v", err)
		c.JSON(http.StatusBadRequest, utils.H{"error": "Invalid request body"})
		return
	}

	log.Printf("接收到消息请求: 接收者=%s, 内容=%s", req.ReceiverUserID, req.Content)

	if req.ReceiverUserID == "" || req.Content == "" {
		log.Printf("缺少必需参数: ReceiverUserID=%s, Content=%s", req.ReceiverUserID, req.Content)
		c.JSON(http.StatusBadRequest, utils.H{"error": "ReceiverUserID and Content are required"})
		return
	}

	message, err := chatService.SendMessage(senderID, req.ReceiverUserID, req.Content)
	if err != nil {
		log.Printf("发送消息失败: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{"error": "Failed to send message", "details": err.Error()})
		return
	}

	log.Printf("消息发送成功: %+v", message)
	c.JSON(http.StatusCreated, message)
}

// GetMessagesHandler handles retrieving messages for a conversation.
// GET /api/chat/messages/:conversation_id?user_id=<user_id_for_marking_read>&page=<page_num>&page_size=<size>
func GetMessagesHandler(_ context.Context, c *app.RequestContext) {
	// JWT认证 - 与其他API保持一致
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}

	userIDForMarkingRead := claims.Username
	conversationID := c.Param("conversation_id")

	// 授权检查：确保当前登录用户是该会话的一部分
	// conversationID 通常是 senderID_receiverID 或 receiverID_senderID
	if !strings.Contains(conversationID, userIDForMarkingRead) {
		// 进一步检查，如果 conversationID 的两部分都不包含 userIDForMarkingRead
		parts := strings.Split(conversationID, "_")
		if len(parts) != 2 || (parts[0] != userIDForMarkingRead && parts[1] != userIDForMarkingRead) {
			c.JSON(http.StatusForbidden, utils.H{"error": "User not authorized for this conversation"})
			return
		}
	}

	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	messages, total, err := chatService.GetMessages(conversationID, userIDForMarkingRead, page, pageSize)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{"error": "Failed to retrieve messages", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"messages":  messages,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetConversationsHandler handles retrieving all conversations for the logged-in user.
// GET /api/chat/conversations
func GetConversationsHandler(_ context.Context, c *app.RequestContext) {
	// JWT认证 - 与其他API保持一致
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}

	userID := claims.Username

	conversations, err := chatService.GetConversations(userID)
	if err != nil {
		log.Printf("Error getting conversations: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{"error": "Failed to retrieve conversations", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, conversations)
}

// MarkAsReadHandler explicitly marks messages in a conversation as read.
// POST /api/chat/markasread
// Request Body: {"conversation_id": "user_a_user_b"}
func MarkAsReadHandler(_ context.Context, c *app.RequestContext) {
	// JWT认证 - 与其他API保持一致
	tokenString := c.GetHeader("Authorization")
	if string(tokenString) == "" {
		c.Status(http.StatusBadRequest)
		c.JSON(http.StatusBadRequest, utils.H{"message": "Authorization token is missing"})
		return
	}

	// 提取 Bearer token
	token_String := strings.Replace(string(tokenString), "Bearer ", "", -1)

	// 解析 token
	token, err := jwt.ParseWithClaims(token_String, &conf.UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return conf.Con.Jwtkey, nil
	})
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token"})
		return
	}

	// 验证 token 是否有效
	claims, ok := token.Claims.(*conf.UserClaims)
	if !ok || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.JSON(http.StatusUnauthorized, utils.H{"message": "Invalid token claims"})
		return
	}

	readerUserID := claims.Username

	var req struct {
		ConversationID string `json:"conversation_id"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{"error": "Invalid request body"})
		return
	}

	if req.ConversationID == "" {
		c.JSON(http.StatusBadRequest, utils.H{"error": "ConversationID is required"})
		return
	}

	// 授权检查：确保当前登录用户是该会话的一部分
	if !strings.Contains(req.ConversationID, readerUserID) {
		parts := strings.Split(req.ConversationID, "_")
		if len(parts) != 2 || (parts[0] != readerUserID && parts[1] != readerUserID) {
			c.JSON(http.StatusForbidden, utils.H{"error": "User not authorized to mark messages in this conversation"})
			return
		}
	}

	err = chatService.MarkMessagesAsRead(req.ConversationID, readerUserID)
	if err != nil {
		log.Printf("Error marking messages as read: %v", err)
		c.JSON(http.StatusInternalServerError, utils.H{"error": "Failed to mark messages as read", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, utils.H{"message": "Messages marked as read"})
}

// Note: WebSocket integration for real-time updates is not fully implemented here.
// You would need to modify your existing chatwebsocket.go or create a new one
// to manage user connections and push messages.
// For example, after a message is saved in SendMessageHandler, you would look up
// the WebSocket connection for the receiverUserID and send the new message through it.
