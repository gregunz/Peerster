package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"log"
	"net/http"
)

type WebServer struct {
	gossiper   *gossiper.Gossiper
	clientChan chan *ClientChannelElement
	allRumors  []*packets_gossiper.RumorMessage
	clients    map[Writer]*client
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebServer(g *gossiper.Gossiper) *WebServer {
	return &WebServer{
		gossiper:   g,
		clientChan: make(chan *ClientChannelElement, 1),
		allRumors:  []*packets_gossiper.RumorMessage{},
		clients:    map[Writer]*client{},
	}
}

func (server *WebServer) Start() {

	router := mux.NewRouter()

	// Configure websocket route
	router.HandleFunc("/ws", server.handleConnections)

	// GET
	router.HandleFunc("/id", server.getIdHandler).Methods("GET")

	// Create a simple file server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gui")))

	go server.handleClientPacket()
	go server.handleRumorSubscriptions()

	// Start the server on localhost port 8000 and log any errors
	port := fmt.Sprintf(":%d", server.gossiper.GUIPort)
	log.Print("WebServer running...\n")
	err := http.ListenAndServe(port, router)
	if err != nil {
		common.HandleAbort("ListenAndServe: ", err)
	}
}

func (server *WebServer) handleClientPacket() {
	for {
		elem, ok := <-server.clientChan
		if ok {
			if err := elem.Packet.Check(); err != nil {
				common.HandleAbort("error in client packet", err)
				continue
			}
			go server.handlePacket(elem.Packet, elem.Writer, false)
		}
	}
}

func (server *WebServer) handleRumorSubscriptions() {
	for {
		rumor, ok := <-server.gossiper.VectorClock().LatestRumorChan
		if ok {
			server.allRumors = append(server.allRumors, rumor)
			for w, c := range server.clients {
				if c.IsSubscribedToMessage {
					common.HandleError(w.WriteJSON(rumor))
				}
			}
		}
	}
}

func (server *WebServer) handlePacket(packet *packets_client.ClientPacket, w Writer, isRest bool) {

	packet.AckPrint()

	client := server.clients[w]

	if packet.IsGetId() {
		common.HandleError(w.WriteJSON(server.gossiper.Name))
		return
	}
	if packet.IsGetMessage() {
		common.HandleError(w.WriteJSON(server.allRumors))
		return
	}
	if packet.IsPostMessage() {
		server.gossiper.ClientChan <- packet.PostMessage
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}
	if packet.IsGetNode() {
		common.HandleError(w.WriteJSON(server.gossiper.PeersSet().ToStrings()))
		return
	}
	if packet.IsPostNode() {
		server.gossiper.PeersSet().AddPeer(packet.PostNode.ToPeer())
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}
	if packet.IsSubscribeMessage() {
		if !client.IsSubscribedToMessage {
			client.IsSubscribedToMessage = true
			if packet.SubscribeMessage.WithPrevious {
				for _, rumor := range server.allRumors {
					common.HandleError(w.WriteJSON(rumor))
				}
			}
		}
		return
	}

	common.HandleAbort("an unexpected event occurred while handling ClientPacket", nil)
}

func (server *WebServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		common.HandleAbort("could not upgrade the connection to websocket", err)
		return
	}

	// Make sure we close the connection when the function returns
	defer func() {
		ws.Close()
		delete(server.clients, ws)
	}()

	c, ok := server.clients[ws]
	if !ok {
		c = NewClient()
		server.clients[ws] = c
		log.Printf("<web-server> new client just arrived")
	}

	for {
		var packet packets_client.ClientPacket
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&packet)
		if err != nil {
			common.HandleAbort("error while reading json of websocket", err)
			break
		}
		// Send the newly received message to the broadcast channel
		server.clientChan <- &ClientChannelElement{
			Packet: &packet,
			Writer: ws,
		}
	}

}

func handlerToWriter(w http.ResponseWriter, r *http.Request) Writer {
	return &ProtoWriter{
		writeJSON: func(v interface{}) error {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			return json.NewEncoder(w).Encode(v)
		},
	}
}

func (server *WebServer) getIdHandler(w http.ResponseWriter, r *http.Request) {
	packet := &packets_client.ClientPacket{
		GetId: &packets_client.GetIdPacket{},
	}
	server.handlePacket(packet, handlerToWriter(w, r), true)
}
