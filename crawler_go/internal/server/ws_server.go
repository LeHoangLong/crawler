package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type WsConnectionWrapper struct {
	conn     *websocket.Conn
	isClosed bool
}

func MakeWsConnecionWrapper(
	iConn *websocket.Conn,
) WsConnectionWrapper {
	return WsConnectionWrapper{
		conn:     iConn,
		isClosed: false,
	}
}

func (c *WsConnectionWrapper) writeTimeout(ctx context.Context, timeout time.Duration, msg []byte) error {
	if c.isClosed {
		return fmt.Errorf("already closed")
	}
	// err := ctx.Err()
	// fmt.Println("err ", err)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := c.conn.Write(ctx, websocket.MessageText, msg)

	if err != nil {
		c.isClosed = true
	}
	return err
}

type MessageType = int

const (
	MessageTypeText   MessageType = 0
	MessageTypeBinary MessageType = 1
)

func (c *WsConnectionWrapper) Read(ctx context.Context) (MessageType, []byte, error) {
	messageType, data, err := c.conn.Read(ctx)

	if err != nil {
		c.isClosed = true
	}

	convertedMesageType := MessageTypeText
	if messageType == websocket.MessageBinary {
		convertedMesageType = MessageTypeBinary
	}

	return convertedMesageType, data, err
}

type WsResponseWriter struct {
	Id   uint32
	Type string
	conn *WsConnectionWrapper
	ctx  *context.Context
}

type WsRequestHandlerI interface {
	HandleWsRequest(ctx context.Context, request map[string]interface{}, response WsResponseWriter) error
}

type WsServer struct {
	handlers map[string]WsRequestHandlerI
}

func MakeWsResponseWriter(
	iId uint32,
	iType string,
	iConn *WsConnectionWrapper,
	iCtx context.Context,
) WsResponseWriter {
	return WsResponseWriter{
		Id:   iId,
		Type: iType,
		conn: iConn,
		ctx:  &iCtx,
	}
}

type WsError = string

const (
	InvalidFormat       WsError = "InvalidFormat"
	UnsupportedEndpoint WsError = "UnsupportedEndpoint"
	InternalError       WsError = "InternalError"
)

type WsClientRequest struct {
	Id   uint32                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type WsClientResponse struct {
	Id   uint32      `json:"id"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func MakeWsServer() WsServer {
	return WsServer{
		handlers: map[string]WsRequestHandlerI{},
	}
}

func (w *WsResponseWriter) SendError(wsError WsError) error {
	response := WsClientResponse{
		w.Id,
		"error",
		fmt.Sprintf("%s", wsError),
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		str := "internal server error"
		err = w.conn.writeTimeout(*w.ctx, time.Second*5, []byte(str))
	} else {
		err = w.conn.writeTimeout(*w.ctx, time.Second*5, responseJson)
	}

	return err
}

func (w *WsResponseWriter) SendData(data interface{}) error {
	response := WsClientResponse{
		w.Id,
		w.Type,
		data,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		str := "internal server error"
		err = w.conn.writeTimeout(*w.ctx, time.Second*5, []byte(str))
	} else {
		err = w.conn.writeTimeout(*w.ctx, time.Second*5, responseJson)
	}

	return err
}

func (s *WsServer) RegisterEndpoint(iEndpoint string, iHandler WsRequestHandlerI) {
	s.handlers[iEndpoint] = iHandler
}

func (s *WsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("serving http")
	options := websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}
	c, err := websocket.Accept(w, r, &options)
	if err != nil {
		fmt.Println("err ", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	fmt.Println("ws connected ", w.Header())
	ctx := r.Context()

	conn := MakeWsConnecionWrapper(c)

	for !conn.isClosed {
		messageType, data, err := conn.Read(ctx)
		fmt.Println("", conn.isClosed, err)

		if messageType != MessageTypeText {
			response := MakeWsResponseWriter(0, "", &conn, ctx)
			response.SendError(InvalidFormat)
			continue
		}

		request := WsClientRequest{}
		err = json.Unmarshal(data, &request)
		if err != nil {
			response := MakeWsResponseWriter(0, "", &conn, ctx)
			response.SendError(InvalidFormat)
			continue
		}

		if request.Id == 0 {
			response := MakeWsResponseWriter(0, "", &conn, ctx)
			response.SendError(InvalidFormat)
			continue
		}

		response := MakeWsResponseWriter(request.Id, request.Type, &conn, ctx)
		if handler, ok := s.handlers[request.Type]; !ok {
			response.SendError(UnsupportedEndpoint)
			continue
		} else {
			err = handler.HandleWsRequest(ctx, request.Data, response)
			if err != nil {
				response.SendError(InternalError)
			}
		}
	}

	fmt.Println("done")
}
