package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/files"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/packets/responses_client"
	"github.com/microcosm-cc/bluemonday"
	"log"
	"net/http"
)

type WebServer struct {
	gossiper        *gossiper.Gossiper
	clientChan      chan *ClientChannelElement
	allRumors       []*packets_gossiper.RumorMessage
	allPrivate      []*packets_gossiper.PrivateMessage
	downloadedFiles []*files.FileType
	clients         map[Writer]*client
	policy          *bluemonday.Policy
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebServer(g *gossiper.Gossiper) *WebServer {
	p := bluemonday.NewPolicy()
	// Require URLs to be parseable by net/url.Parse and either:
	//   mailto: http:// or https://
	p.AllowStandardURLs()
	// We only allow <p> and <a href="">
	p.AllowAttrs("href").OnElements("a")
	p.AllowElements("p")

	return &WebServer{
		gossiper:   g,
		clientChan: make(chan *ClientChannelElement, 1),
		clients:    map[Writer]*client{},
		policy:     p,
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
	go server.handlePrivateSubscriptions()
	go server.handleNodeSubscriptions()
	go server.handleOriginsSubscriptions()
	go server.handleIndexedFilesSubscriptions()
	go server.handleDownloadedFilesSubscriptions()

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
		msg := server.gossiper.RumorChan.GetRumor()
		if msg != nil {
			server.allRumors = append(server.allRumors, msg)
			for w, c := range server.clients {
				if c.IsSubscribedToMessage {
					common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(msg, server.policy)))
				}
			}
		}
	}
}

func (server *WebServer) handlePrivateSubscriptions() {
	for {
		msg := server.gossiper.PrivateMsgChan.GetPrivateMsg()
		if msg != nil {
			for w, c := range server.clients {
				if c.IsSubscribedToMessage {
					common.HandleError(w.WriteJSON(responses_client.NewPrivateResponse(msg, server.policy)))
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
					common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer, server.policy)))
				}
			}
		}
	}
}

func (server *WebServer) handleOriginsSubscriptions() {
	for {
		o := server.gossiper.OriginChan.GetOrigin()
		if o != "" && o != server.gossiper.Origin {
			for w, c := range server.clients {
				if c.IsSubscribedToOrigin {
					common.HandleError(w.WriteJSON(responses_client.NewContactResponse(o, server.policy)))
				}
			}
		}
	}
}

func (server *WebServer) handleIndexedFilesSubscriptions() {
	for {
		file := server.gossiper.IndexedFilesChan.Get()
		if file != nil {
			for w, c := range server.clients {
				if c.IsSubscribedToFiles {
					common.HandleError(w.WriteJSON(responses_client.NewIndexedFileResponse(file, server.policy)))
				}
			}
		}
	}
}
func (server *WebServer) handleDownloadedFilesSubscriptions() {
	for {
		file := server.gossiper.DownloadedFilesChan.Get()
		if file != nil {
			server.downloadedFiles = append(server.downloadedFiles, file)
			for w, c := range server.clients {
				if c.IsSubscribedToFiles {
					common.HandleError(w.WriteJSON(responses_client.NewDownloadedFileResponse(file, server.policy)))
				}
			}
		}
	}
}

func (server *WebServer) handlePacket(packet *packets_client.ClientPacket, w Writer, isRest bool) {

	fmt.Printf("[GUI] received packet <%s>\n", packet.String())

	client := server.clients[w]

	if packet.IsGetId() {
		common.HandleError(w.WriteJSON(responses_client.NewGetIdResponse(server.gossiper.Origin, server.policy)))
		return
	}
	if packet.IsPostMessage() {
		go func() { server.gossiper.FromClientChan <- packet.PostMessage.ToClientPacket() }()
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}
	if packet.IsPostNode() {
		peer := packet.PostNode.ToPeer()
		if peer != nil &&
			!server.gossiper.GossipAddr.Equals(peer.Addr) &&
			!server.gossiper.ClientAddr.Equals(peer.Addr) {
			server.gossiper.PeersSet.Add(peer)
		} else {
			common.HandleAbort("cannot add node", nil)
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
					common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(rumor, server.policy)))
				}
				for _, msg := range server.gossiper.Conversations.GetAll() {
					common.HandleError(w.WriteJSON(responses_client.NewPrivateResponse(msg, server.policy)))
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
					common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer, server.policy)))
				}
			}
		} else if client.IsSubscribedToNode && !packet.SubscribeNode.Subscribe {
			client.IsSubscribedToNode = false
		}
		return
	}

	if packet.IsSubscribeOrigin() {
		if !client.IsSubscribedToOrigin && packet.SubscribeOrigin.Subscribe {
			client.IsSubscribedToOrigin = true
			if packet.SubscribeOrigin.WithPrevious {
				for _, o := range server.gossiper.RoutingTable.GetOrigins() {
					if o != server.gossiper.Origin {
						common.HandleError(w.WriteJSON(responses_client.NewContactResponse(o, server.policy)))
					}
				}
			}
		} else if client.IsSubscribedToOrigin && !packet.SubscribeOrigin.Subscribe {
			client.IsSubscribedToOrigin = false
		}
		return
	}

	if packet.IsIndexFile() {
		go func() { server.gossiper.FromClientChan <- packet.IndexFile.ToClientPacket() }()
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}

	if packet.IsRequestFile() {
		go func() { server.gossiper.FromClientChan <- packet.RequestFile.ToClientPacket() }()
		if isRest {
			common.HandleError(w.WriteJSON(nil))
		}
		return
	}

	if packet.IsSubscribeFile() {
		if !client.IsSubscribedToFiles && packet.SubscribeFile.Subscribe {
			client.IsSubscribedToFiles = true
			if packet.SubscribeFile.WithPrevious {
				for _, file := range server.gossiper.FilesUploader.GetAllFiles() {
					common.HandleError(w.WriteJSON(responses_client.NewIndexedFileResponse(file, server.policy)))
				}
				for _, filename := range server.downloadedFiles {
					common.HandleError(w.WriteJSON(responses_client.NewDownloadedFileResponse(filename, server.policy)))
				}
			}
		} else if client.IsSubscribedToFiles && !packet.SubscribeFile.Subscribe {
			client.IsSubscribedToFiles = false
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
			if _, ok := err.(*websocket.CloseError); ok {
				common.HandleError(err)
				break
			} else {
				common.HandleAbort("error while reading json of websocket", err)
				continue
			}
		}
		// Send the newly received message to the broadcast channel
		server.clientChan <- &ClientChannelElement{
			Packet: &packet,
			Writer: w,
		}
	}

}

func restToWriter(w http.ResponseWriter) Writer {
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
	server.handlePacket(packet, restToWriter(w), true)
}
