package dataconverter

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v2"

	commonpb "go.temporal.io/api/common/v1"
)

type PayloadRequest struct {
	RequestID string `json:"requestId"`
	Payload   string `json:"payload"`
}

type PayloadResponse struct {
	RequestID string `json:"requestId"`
	Content   string `json:"content"`
}

func processMessage(c *websocket.Conn) error {
	mt, message, err := c.ReadMessage()
	if err != nil {
		return err
	}

	var payloadRequest PayloadRequest
	err = json.Unmarshal(message, &payloadRequest)
	if err != nil {
		return fmt.Errorf("invalid payload request: %w", err)
	}

	var payload commonpb.Payload
	err = jsonpb.UnmarshalString(payloadRequest.Payload, &payload)
	if err != nil {
		return fmt.Errorf("invalid payload data: %w", err)
	}

	payloadResponse := PayloadResponse{
		RequestID: payloadRequest.RequestID,
		Content:   GetCurrent().ToString(&payload),
	}

	var response []byte
	response, err = json.Marshal(payloadResponse)
	if err != nil {
		return fmt.Errorf("unable to marshal response: %w", err)
	}

	err = c.WriteMessage(mt, response)
	if err != nil {
		return fmt.Errorf("unable to write response: %w", err)
	}

	return nil
}

func buildPayloadHandler(context *cli.Context, origin string) func(http.ResponseWriter, *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			if r.Header.Get("Origin") != origin {
				fmt.Printf("invalid origin: %s\n", origin)
				return false
			}
			return true
		},
	}

	return func(res http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(res, req, nil)
		if err != nil {
			fmt.Printf("data converter websocket upgrade failed: %v\n", err)
			return
		}
		defer c.Close()

		for {
			err := processMessage(c)
			if err != nil {
				if closeError, ok := err.(*websocket.CloseError); ok {
					if closeError.Code == websocket.CloseNoStatusReceived ||
						closeError.Code == websocket.CloseNormalClosure {
						return
					}
				}
				fmt.Fprintln(os.Stderr, fmt.Errorf("data converter websocket error: %w", err))

				return
			}
		}
	}
}
