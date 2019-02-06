package api

import (
	"log"
	"net/http"
	"path"
	"runtime"
	"server/datatypes"
	"server/listener"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// global datapoint subscriber
var points chan *datatypes.Datapoint

// Write strings to these websockets
var activeWebsockets []*websocket.Conn
var socketLock sync.Mutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// ChatDefault is the default handler for the /chat path
func (api *API) ChatDefault(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/chat/static/index.html", http.StatusFound)
}

// SubscribeDriverStatus will listen to the datapoint publisher
// and filter all data points except for driver ACK/NACK,
// and relay such messages to each WebSocket currently connected
func SubscribeDriverStatus() {
	points = make(chan *datatypes.Datapoint)
	listener.Subscribe(points)
	for {
		point := <-points
		if point.Metric == "Driver_ACK_Status" {
			socketLock.Lock()
			for _, conn := range activeWebsockets {
				msg := ""
				if point.Value == 0.0 {
					msg = "Driver Response NACK"
				} else if point.Value == 1.0 {
					msg = "Driver Response ACK"
				} else {
					// don't send NULL
					continue
				}
				err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					log.Println("Failed to write message to websocket")
					continue
				}
			}
			socketLock.Unlock()
		}
	}
}

func addSocketConn(conn *websocket.Conn) {
	socketLock.Lock()
	defer socketLock.Unlock()
	activeWebsockets = append(activeWebsockets, conn)
	log.Printf("Chat Websocket connected. %s\n", conn.RemoteAddr())
	log.Printf("Active websocket connections: %d\n", len(activeWebsockets))
}

func removeSocketConn(conn *websocket.Conn) {
	socketLock.Lock()
	defer socketLock.Unlock()
	for i, c := range activeWebsockets {
		if c == conn {
			log.Printf("Client Chat Websocket Disconnected. %s\n", conn.RemoteAddr())
			activeWebsockets = append(activeWebsockets[:i], activeWebsockets[i+1:]...)
			log.Printf("Active websocket connections: %d\n", len(activeWebsockets))
			return
		}
	}
	log.Fatalf("Connection not found in activeWebsockets!")
}

// ChatSocket initializes the WebSocket for a connection.
// The websocket is responsible for relaying driver responses (Yes, No)
// to the client as well
func (api *API) ChatSocket(res http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		log.Println("Client Chat Websocket failed to initialize.")
		return
	}

	socketLock.Lock()
	if points == nil {
		go SubscribeDriverStatus()
	}
	socketLock.Unlock()

	// Subscribe this socket to ACK/NACK changes
	addSocketConn(conn)
	defer removeSocketConn(conn)

	// read messages from each websocket
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if len(string(msg)) > 35 {
			err := conn.WriteMessage(msgType, []byte("Message length too long"))
			if err != nil {
				return
			}
		} else {
			// Print the message to the console
			log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
			// upload message to car
			uploadTCPMessage(string(msg))
		}
	}
}

// uploadTCPMessage sends our message to the listener
// which will then relay to the car
func uploadTCPMessage(message string) {
	// send GTSR, the length of the string, and the string itself
	msg := []byte("GTSR")
	messageLength := []byte{byte(len(message))}
	msg = append(msg, messageLength...)
	msg = append(msg, []byte(message)...)
	listener.Write(msg)
}

// RegisterChatRoutes registers the routes for the chat service
func (api *API) RegisterChatRoutes(router *mux.Router) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)
	router.PathPrefix("/chat/static/").Handler(http.StripPrefix("/chat/static/", http.FileServer(http.Dir(path.Join(dir, "chat")))))

	router.HandleFunc("/chat", api.ChatDefault).Methods("GET")

	router.HandleFunc("/chat/socket", api.ChatSocket)
}
