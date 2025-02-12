package middle

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
		defer conn.Close()
		config := ark.DefaultConfig(os.Getenv("ARK_API_KEY"))
		config.BaseURL = "https://ark.cn-beijing.volces.com/api/v3"
		client := ark.NewClientWithConfig(config)

		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}

			if messageType == websocket.TextMessage {
				var chatRequest ChatRequest
				if err := json.Unmarshal(p, &chatRequest); err != nil {
					log.Println("unmarshal error:", err)
					errorMessage := ChatResponse{Response: "Invalid request format"}
					errorResponseJson, _ := json.Marshal(errorMessage)
					if err := conn.WriteMessage(websocket.TextMessage, errorResponseJson); err != nil {
						log.Println("write error:", err)
					}
					continue
				}

				stream, err := client.CreateChatCompletionStream(
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
				if err != nil {
					log.Println("API stream request error:", err)
					errorMessage := ChatResponse{Response: "Error communicating with AI service"}
					errorResponseJson, _ := json.Marshal(errorMessage)
					if err := conn.WriteMessage(websocket.TextMessage, errorResponseJson); err != nil {
						log.Println("write error:", err)
					}
					continue
				}
				defer stream.Close()

				// Initialize variable to check when the response stream is complete
				var completed bool

				for {
					recv, err := stream.Recv()
					if err == io.EOF {
						fmt.Println("Stream finished.")
						break
					}
					if err != nil {
						log.Printf("Stream recv error: %v\n", err)
						errorMessage := ChatResponse{Response: "Error receiving data from AI service"}
						errorResponseJson, _ := json.Marshal(errorMessage)
						if err := conn.WriteMessage(websocket.TextMessage, errorResponseJson); err != nil {
							log.Println("write error:", err)
						}
						break
					}

					if len(recv.Choices) > 0 {
						chatResponseChunk := ChatResponse{
							Response: recv.Choices[0].Delta.Content,
						}
						responseJsonChunk, _ := json.Marshal(chatResponseChunk)
						if err := conn.WriteMessage(websocket.TextMessage, responseJsonChunk); err != nil {
							log.Println("write error:", err)
							break
						}
						fmt.Print(recv.Choices[0].Delta.Content)
					}

					if !completed && err == io.EOF {
						completed = true
						finalMessage := ChatResponse{Response: "Full response stream handled."}
						finalMessageJson, _ := json.Marshal(finalMessage)
						if err := conn.WriteMessage(websocket.TextMessage, finalMessageJson); err != nil {
							log.Println("write error:", err)
							break
						}
					}
				}
			}
		}
	})

	if err != nil {
		if _, ok := err.(websocket.HandshakeError); ok {
			c.String(http.StatusBadRequest, "Not a websocket handshake")
		} else {
			c.String(http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
}
