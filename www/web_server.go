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
	"github.com/gregunz/Peerster/models/packets/responses_client"
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
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gui/public")))

	go server.handleClientPacket()
	go server.handleRumorSubscriptions()
	go server.handleNodeSubscriptions()

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
		rumor := server.gossiper.RumorChan.GetRumor()
		if rumor != nil {
			server.allRumors = append(server.allRumors, rumor)
			for w, c := range server.clients {
				if c.IsSubscribedToMessage {
					common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(rumor)))
				}
			}
		}
	}
}

func (server *WebServer) handleNodeSubscriptions() {
	for {
		peer := server.gossiper.NodeChan.GetNode()
		if peer != nil {
			for w, c := range server.clients {
				if c.IsSubscribedToNode {
					common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer.Addr.ToIpPort())))
				}
			}
		}
	}
}

func (server *WebServer) handlePacket(packet *packets_client.ClientPacket, w Writer, isRest bool) {

	packet.AckPrint()

	client := server.clients[w]

	if packet.IsGetId() {
		common.HandleError(w.WriteJSON(responses_client.NewGetIdResponse(server.gossiper.Name)))
		return
	}
	if packet.IsPostMessage() {
		go func() { server.gossiper.ClientChan <- packet.PostMessage }()
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}
	if packet.IsPostNode() {
		peer := packet.PostNode.ToPeer()
		if !peer.Addr.Equals(server.gossiper.Addr) {
			server.gossiper.PeersSet.Add(peer)
		} else {
			common.HandleAbort("cannot add node with same address as gossiper", nil)
		}
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}
	if packet.IsSubscribeMessage() {
		if !client.IsSubscribedToMessage && packet.SubscribeMessage.Subscribe {
			client.IsSubscribedToMessage = true
			if packet.SubscribeMessage.WithPrevious {
				for _, rumor := range server.allRumors {
					common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(rumor)))
				}
			}
		} else if client.IsSubscribedToMessage && !packet.SubscribeMessage.Subscribe {
			client.IsSubscribedToMessage = false
		}
		return
	}
	if packet.IsSubscribeNode() {
		if !client.IsSubscribedToNode && packet.SubscribeNode.Subscribe {
			client.IsSubscribedToNode = true
			if packet.SubscribeNode.WithPrevious {
				for _, peer := range server.gossiper.PeersSet.GetSlice() {
					common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer.Addr.ToIpPort())))
				}
			}
		} else if client.IsSubscribedToNode && !packet.SubscribeNode.Subscribe {
			client.IsSubscribedToNode = false
		}
		return
	}

	common.HandleAbort("an unexpected event occurred while handling ClientPacket", nil)
}

func (server *WebServer) handleConnections(writer http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(writer, r, nil)
	if err != nil {
		common.HandleAbort("could not upgrade the connection to websocket", err)
		return
	}

	w := websocketToWriter(ws)
	c, ok := server.clients[w]

	// Make sure we close the connection when the function returns
	defer func() {
		ws.Close()
		delete(server.clients, w)
	}()

	if !ok {
		c = NewClient()
		server.clients[w] = c
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
			Writer: w,
		}
	}

}

func handlerToWriter(w http.ResponseWriter, r *http.Request) Writer {
	return &ProtoWriter{
		writeJSON: func(v *responses_client.ClientResponse) error {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			return json.NewEncoder(w).Encode(v)
		},
	}
}

func websocketToWriter(ws *websocket.Conn) Writer {
	return &ProtoWriter{
		writeJSON: func(v *responses_client.ClientResponse) error {
			return ws.WriteJSON(v)
		},
	}
}

func (server *WebServer) getIdHandler(w http.ResponseWriter, r *http.Request) {
	packet := &packets_client.ClientPacket{
		GetId: &packets_client.GetIdPacket{},
	}
	server.handlePacket(packet, handlerToWriter(w, r), true)
}
