package middle

import (
	"context"
	"encoding/json"
	"fmt"
	"hyperledger-fabric-copyright/conf"
	"io"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	ark "github.com/sashabaranov/go-openai"
)

var upgrader = websocket.HertzUpgrader{}

type ChatRequest struct {
	Message string `json:"message"`
	Prompt  string `json:"prompt"`
}

type ChatResponse struct {
	Response string `json:"response"`
}

func ChatWebsocket(_ context.Context, c *app.RequestContext) {
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in websocket handler: %v", r)
			}
			if err := conn.Close(); err != nil {
				log.Println("Error closing websocket connection:", err)
			} else {
				log.Println("WebSocket connection closed normally.")
			}
		}()

		config := ark.DefaultConfig(conf.Con.APIKey)
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
		client := ark.NewClientWithConfig(config)

		log.Println("WebSocket connection established from:", conn.RemoteAddr()) // 添加连接建立日志

		for {
			messageType, p, readErr := conn.ReadMessage()
			if readErr != nil {
				log.Printf("read error from %s: %v, messageType: %d", conn.RemoteAddr(), readErr, messageType) // 更详细的读取错误日志
				break                                                                                          // 退出循环，连接已断开
			}

			if messageType == websocket.TextMessage {
				var chatRequest ChatRequest
				if unmarshalErr := json.Unmarshal(p, &chatRequest); unmarshalErr != nil {
					log.Printf("unmarshal error from %s: %v, message: %s", conn.RemoteAddr(), unmarshalErr, string(p)) // 包含消息内容的Unmarshal错误日志
					errorMessage := ChatResponse{Response: "Invalid request format"}
					errorResponseJson, _ := json.Marshal(errorMessage)
					if writeErr := conn.WriteMessage(websocket.TextMessage, errorResponseJson); writeErr != nil {
						log.Printf("write error to %s: %v, message: %s", conn.RemoteAddr(), writeErr, string(errorResponseJson)) // 错误响应写入日志
					}
					continue // 继续下一次循环，处理新的消息
				}
				log.Printf("Received message from %s: %+v", conn.RemoteAddr(), chatRequest) // 记录接收到的消息

				// API 请求前打印日志
				log.Printf("Sending API request to AI service from %s, prompt: %s, message: %s", conn.RemoteAddr(), chatRequest.Prompt, chatRequest.Message)
				stream, createStreamErr := client.CreateChatCompletionStream(
					context.Background(),
					ark.ChatCompletionRequest{
						Model: "ep-20250210154913-vx8xj",
						Messages: []ark.ChatCompletionMessage{
							{
								Role:    ark.ChatMessageRoleSystem,
								Content: chatRequest.Prompt,
							},
							{
								Role:    ark.ChatMessageRoleUser,
								Content: chatRequest.Message,
							},
						},
					},
				)
				if createStreamErr != nil {
					log.Printf("API stream request error from %s: %v", conn.RemoteAddr(), createStreamErr) // 更详细的API请求错误日志
					errorMessage := ChatResponse{Response: "Error communicating with AI service"}
					errorResponseJson, _ := json.Marshal(errorMessage)
					if writeErr := conn.WriteMessage(websocket.TextMessage, errorResponseJson); writeErr != nil {
						log.Printf("write error to %s: %v, message: %s", conn.RemoteAddr(), writeErr, string(errorResponseJson)) // 错误响应写入日志
					}
					continue // 继续下一次循环，处理新的消息
				}
				defer stream.Close()
				log.Println("API stream request successful, stream started for:", conn.RemoteAddr()) // 记录stream开始

				completed := false // 初始化变量以检查响应流是否完成

				for {
					recv, recvErr := stream.Recv()
					if recvErr == io.EOF {
						log.Println("Stream finished for:", conn.RemoteAddr()) // 记录stream正常结束
						break                                                  // stream 完成
					}
					if recvErr != nil {
						log.Printf("Stream recv error from %s: %v", conn.RemoteAddr(), recvErr) // 更详细的stream接收错误日志
						errorMessage := ChatResponse{Response: "Error receiving data from AI service"}
						errorResponseJson, _ := json.Marshal(errorMessage)
						if writeErr := conn.WriteMessage(websocket.TextMessage, errorResponseJson); writeErr != nil {
							log.Printf("write error to %s: %v, message: %s", conn.RemoteAddr(), writeErr, string(errorResponseJson)) // 错误响应写入日志
						}
						break // 退出 stream 接收循环
					}

					if len(recv.Choices) > 0 && recv.Choices[0].Delta.Content != "" { // 检查 Content 是否为空
						chatResponseChunk := ChatResponse{
							Response: recv.Choices[0].Delta.Content,
						}
						responseJsonChunk, _ := json.Marshal(chatResponseChunk)
						if writeErr := conn.WriteMessage(websocket.TextMessage, responseJsonChunk); writeErr != nil {
							log.Printf("write error to %s: %v, message: %s", conn.RemoteAddr(), writeErr, string(responseJsonChunk)) // 响应数据写入错误日志
							break                                                                                                    // 写入错误，退出 stream 接收循环
						}
						fmt.Print(recv.Choices[0].Delta.Content)
					}

					if !completed && recvErr == io.EOF { // 再次检查 recvErr 是否为 io.EOF 是多余的，因为在循环开始时已经判断过
						completed = true
						finalMessage := ChatResponse{Response: "Full response stream handled."}
						finalMessageJson, _ := json.Marshal(finalMessage)
						if writeErr := conn.WriteMessage(websocket.TextMessage, finalMessageJson); writeErr != nil {
							log.Printf("write final message error to %s: %v, message: %s", conn.RemoteAddr(), writeErr, string(finalMessageJson)) // 最终消息写入错误日志
							break                                                                                                                 // 最终消息写入错误，退出 stream 接收循环
						}
						log.Println("Final response message sent to:", conn.RemoteAddr()) // 记录最终消息发送
					}
				}
				log.Println("Stream processing finished for message from:", conn.RemoteAddr()) // 记录消息处理完成
			}
		}
	})

	if err != nil {
		log.Printf("websocket upgrade error: %v", err) // 记录 upgrade 错误
		if _, ok := err.(websocket.HandshakeError); ok {
			c.String(http.StatusBadRequest, "Not a websocket handshake")
		} else {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
}
