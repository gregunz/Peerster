package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/packets"
	"log"
	"net/http"
)

type WebServer struct {
	gossiper   *gossiper.Gossiper
	clientChan chan *ClientChannelElement
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebServer(g *gossiper.Gossiper) *WebServer {
	fmt.Printf("Creating WebServer on port %d", g.GUIPort)
	return &WebServer{
		gossiper:   g,
		clientChan: make(chan *ClientChannelElement, 1),
	}
}

func (server *WebServer) Start() {

	router := mux.NewRouter()

	// Create a simple file server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gui")))

	// Configure websocket route
	router.HandleFunc("/ws", server.handleConnections)

	// GET
	router.HandleFunc("/id", server.getIdHandler).Methods("GET")
	router.HandleFunc("/node", server.getNodeHandler).Methods("GET")
	router.HandleFunc("/message", server.getMessageHandler).Methods("GET")

	// POST
	router.HandleFunc("/node", server.postNodeHandler).Methods("POST")
	router.HandleFunc("/message", server.postMessageHandler).Methods("POST")

	// Start the server on localhost port 8000 and log any errors
	port := fmt.Sprintf(":%d", server.gossiper.GUIPort)
	log.Printf("http server started on %s\n", port)
	err := http.ListenAndServe(port, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (server *WebServer) handleClientPacket() {
	for {
		elem := <-server.clientChan
		packet, _ := elem.Packet, elem.Writer

		if err := packet.Check(); err != nil {
			common.HandleAbort("error in client packet", err)
			continue
		}

		if packet.IsText() {
			server.gossiper.ClientChan <- packet.Text
		}
		if packet.IsAddNode() {
			server.gossiper.PeersSet().AddPeer(packet.AddNode.ToPeer())
		}
	}
}

func (server *WebServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	for {
		var packet packets.ClientPacket
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&packet)
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		// Send the newly received message to the broadcast channel
		server.clientChan <- &ClientChannelElement{
			Packet: &packet,
			Writer: ws,
		}
	}

}

func (server *WebServer) getIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.Name)
}

func (server *WebServer) getNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.PeersSet().ToStrings())
}

func (server *WebServer) getMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.VectorClock().GetAllMessages())
}
