package api

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"server/datatypes"
	"server/listener"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/websocket"
	"github.com/nlopes/slack"
)

var hashKey = []byte(securecookie.GenerateRandomKey(32))
var blockKey = []byte(securecookie.GenerateRandomKey(32))

const cookieName = "login-cookie"

const maxMessageLength = 35

const requireAuthorization = false

var authorizedUsers = map[string]bool{
	"U0JSX098T": true, // Jared
	"U2PVAD9B7": true, // Alex
	"U709MM717": true, // Steven
	"UCRCSKHBN": true, // Noah
	"UCR9YCZCN": true, // Kat
}

var slck *slack.Client

func postSlackMessage(message string) {
	if slck != nil {
		slck.PostMessage("chat", slack.MsgOptionText(message, false))
	} else {
		log.Printf("No slack key - message: %s", message)
	}
}

var sc = securecookie.New(hashKey, blockKey)

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Write strings to these websockets
var activeWebsockets []*websocket.Conn
var socketLock sync.Mutex

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// SetCookieHandler Sets a cookie for a given request
// gets set in the /login route
// If production, set secure to true
func SetCookieHandler(w http.ResponseWriter, r *http.Request) {
	token := randToken()
	value := map[string]string{
		"id": token,
	}

	useSecure := false
	if os.Getenv("PRODUCTION") == "true" {
		useSecure = true
	}

	log.Printf("set cookie: %s\n", token)
	if encoded, err := sc.Encode(cookieName, value); err == nil {
		cookie := &http.Cookie{
			Name:     cookieName,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
			Secure:   useSecure,
		}
		http.SetCookie(w, cookie)
	}
}

// ReadCookieHandler Reads the cookie and returns false if it is invalid
func ReadCookieHandler(r *http.Request) bool {
	if cookie, err := r.Cookie(cookieName); err == nil {
		value := make(map[string]string)
		if err = sc.Decode(cookieName, cookie.Value, &value); err == nil {
			log.Printf("Cookie Accepted: %s\n", cookie)
			return true
		}
		log.Printf("Cookie Rejected: %s\n", cookie)
	}
	return false
}

// ChatLogin will issue a token. It is secured in Basic Auth at the NGINX layer
func (api *API) ChatLogin(res http.ResponseWriter, req *http.Request) {
	// Grant a login token via a cookie if it does not exist
	if !ReadCookieHandler(req) {
		SetCookieHandler(res, req)
	}

	// Then redirect to the chat index
	http.Redirect(res, req, "/chat/static/index.html", http.StatusFound)
}

func checkAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if ReadCookieHandler(req) {
			h.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, "/chat/login", http.StatusFound)
		}

	}
}

func checkAuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if ReadCookieHandler(req) {
			h.ServeHTTP(res, req)
		} else {
			http.Redirect(res, req, "/chat/login", http.StatusFound)
		}
	})
}

// ChatDefault is the default handler for the /chat path
func (api *API) ChatDefault(res http.ResponseWriter, req *http.Request) {
	// Check if login session exists, if it doesn't, then redirect to login
	if ReadCookieHandler(req) {
		http.Redirect(res, req, "/chat/static/index.html", http.StatusFound)
	} else {
		http.Redirect(res, req, "/chat/login", http.StatusFound)
	}
}

// ChatSlashCommand is the handler for the chat slash command, which sends
// a message to the car
func (api *API) ChatSlashCommand(res http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	user := req.Form.Get("user_id")
	if requireAuthorization && !authorizedUsers[user] {
		slackResponse("Unauthorized user", res)
		return
	}
	channel := req.Form.Get("channel_name")
	if channel != "chat" && channel != "testing" {
		slackResponse("Requests must be made from #chat", res)
		return
	}
	msg := req.Form.Get("text")
	if len(msg) == 0 {
		slackResponse("Please provide a message", res)
		return
	}
	if len(msg) > maxMessageLength {
		slackResponse("Message exceeds maximum length", res)
		return
	}
	uploadTCPMessage(msg)
	slackResponse("Message sent", res)
}

func slackResponse(text string, res http.ResponseWriter) {
	response := make(map[string]string)
	response["response_type"] = "in_channel"
	response["text"] = text
	res.Header().Set("Content-Type", "application/json")
	json.NewEncoder(res).Encode(response)
}

// ChatSocket initializes the WebSocket for a connection.
// The websocket is responsible for relaying driver responses (Yes, No)
// to the client as well
func (api *API) ChatSocket(res http.ResponseWriter, req *http.Request) {
	// Only upgrade connection with proper cookie.

	if !ReadCookieHandler(req) {
		log.Println("Socket Connection Attempt Rejected!")
		return
	}

	conn, err := upgrader.Upgrade(res, req, nil)

	if err != nil {
		log.Println("Client Chat Websocket failed to initialize.")
		return
	}

	// Subscribe this socket to ACK/NACK changes
	addSocketConn(conn)
	defer removeSocketConn(conn)

	// ping client every 10 seconds to keep alive
	go pingClient(conn)
	// read messages from websocket
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}

		if len(string(msg)) > maxMessageLength {
			err := conn.WriteMessage(msgType, []byte("Message length too long"))
			if err != nil {
				log.Println(err)
				return
			}
		} else {
			// Print the message to the console
			log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))
			postSlackMessage("Strategy: " + string(msg))
			// upload message to car
			uploadTCPMessage(string(msg))
		}
	}
}

// SubscribeDriverStatus will listen to the datapoint publisher
// and filter all data points except for driver ACK/NACK,
// and relay such messages to each WebSocket currently connected
func SubscribeDriverStatus() {
	points := make(chan *datatypes.Datapoint)
	listener.Subscribe(points)
	for {
		point := <-points
		if point.Metric == "Driver_ACK_Status" {
			msg := ""
			if point.Value == 0.0 {
				msg = "Driver Response NACK"
				postSlackMessage("Driver: NACK")
			} else if point.Value == 1.0 {
				msg = "Driver Response ACK"
				postSlackMessage("Driver: ACK")
			}
			if msg == "" {
				continue
			}
			socketLock.Lock()
			for _, conn := range activeWebsockets {
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

func pingClient(conn *websocket.Conn) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		<-ticker.C
		err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second*10))
		if err != nil {
			log.Println(err)
			return
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
	go SubscribeDriverStatus()
	go MonitorConnection()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	dir := path.Dir(filename)

	if token, err := ioutil.ReadFile("/secrets/slack_token.txt"); err == nil {
		slck = slack.New(strings.TrimSpace(string(token)))
	} else {
		log.Printf("Unable to find Slack credentials: %s\n", err)
	}

	router.PathPrefix("/chat/static").Handler(checkAuthHandler(http.StripPrefix("/chat/static/", http.FileServer(http.Dir(path.Join(dir, "chat"))))))
	router.PathPrefix("/chat-login/static").Handler(http.StripPrefix("/chat-login/static/", http.FileServer(http.Dir(path.Join(dir, "chat-login")))))
	router.HandleFunc("/chat/login", api.ChatLogin).Methods("GET")
	router.HandleFunc("/chat", checkAuth(api.ChatDefault)).Methods("GET")
	router.HandleFunc("/chat/socket", api.ChatSocket)
	router.HandleFunc("/chatSlashCommand", api.ChatSlashCommand).Methods("GET")
}

// MonitorConnection listens for connection status messages and
// posts Slack messages in response
func MonitorConnection() {
	c := make(chan *datatypes.Datapoint, 10)
	err := listener.Subscribe(c)
	if err != nil {
		log.Fatalf("Error getting datapoint publisher")
	}
	for point := range c {
		if point.Metric == "Connection_Status" {
			if point.Value == 1 {
				postSlackMessage("Connection established")
			} else {
				postSlackMessage("Connection lost")
			}
		}
	}
}
