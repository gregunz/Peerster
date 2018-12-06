package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/logger"
	"github.com/gregunz/Peerster/models/files"
	"github.com/gregunz/Peerster/models/packets/packets_client"
	"github.com/gregunz/Peerster/models/packets/packets_gossiper"
	"github.com/gregunz/Peerster/models/packets/responses_client"
	"github.com/gregunz/Peerster/www/clients"
	"github.com/gregunz/Peerster/www/subscription"
	"github.com/microcosm-cc/bluemonday"
	"net/http"
	"sync"
)

type WebServer struct {
	gossiper        *gossiper.Gossiper
	clientChan      chan *clients.ClientChannelElement
	allRumors       []*packets_gossiper.RumorMessage
	allPrivate      []*packets_gossiper.PrivateMessage
	downloadedFiles []*files.FileType
	clients         *clients.Map
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
		clientChan: make(chan *clients.ClientChannelElement, 1),
		clients:    clients.NewMap(),
		policy:     p,
	}
}

func (server *WebServer) Start(group sync.WaitGroup) {

	router := mux.NewRouter()

	// Configure websocket route
	router.HandleFunc("/ws", server.handleConnections)

	// GET
	router.HandleFunc("/id", server.getIdHandler).Methods("GET")

	// Create a simple file server
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gui")))

	group.Add(1)
	go server.handleClientPacket(group)

	group.Add(1)
	go server.handleRumorSubscriptions(group)

	group.Add(1)
	go server.handlePrivateSubscriptions(group)

	group.Add(1)
	go server.handleNodeSubscriptions(group)

	group.Add(1)
	go server.handleOriginsSubscriptions(group)

	group.Add(1)
	go server.handleIndexedFilesSubscriptions(group)

	group.Add(1)
	go server.handleDownloadedFilesSubscriptions(group)

	group.Add(1)
	go server.handleSearchedFilesSubscriptions(group)

	// Start the server on localhost port 8000 and log any errors
	port := fmt.Sprintf(":%d", server.gossiper.GUIPort)
	logger.Printlnf("WebServer running...")
	err := http.ListenAndServe(port, router)
	if err != nil {
		common.HandleAbort("ListenAndServe: ", err)
	}
}

func (server *WebServer) handleClientPacket(group sync.WaitGroup) {
	defer group.Done()
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

func (server *WebServer) handleRumorSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		msg := server.gossiper.VectorClock.RumorChan.Get()
		if msg != nil {
			server.allRumors = append(server.allRumors, msg)
			logger.Printlnf(">%s", msg)
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				logger.Printlnf(">>%s", msg)
				logger.Printlnf(">>%s", c)
				if c.IsSubscribedTo(subscription.Message) {
					common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(msg, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handlePrivateSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		msg := server.gossiper.Conversations.PrivateMsgChan.Get()
		if msg != nil {
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.Message) {
					common.HandleError(w.WriteJSON(responses_client.NewPrivateResponse(msg, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handleNodeSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		peer := server.gossiper.PeersSet.NodeChan.Get()
		if peer != nil {
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.Node) {
					common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handleOriginsSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		o := server.gossiper.RoutingTable.OriginChan.Get()
		if o != "" && o != server.gossiper.Origin {
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.Origin) {
					common.HandleError(w.WriteJSON(responses_client.NewContactResponse(o, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handleIndexedFilesSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		file := server.gossiper.FilesUploader.FileChan.Get()
		if file != nil {
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.File) {
					common.HandleError(w.WriteJSON(responses_client.NewIndexedFileResponse(file, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handleDownloadedFilesSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		file := server.gossiper.FilesDownloader.FileChan.Get()
		if file != nil {
			server.downloadedFiles = append(server.downloadedFiles, file)
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.File) {
					common.HandleError(w.WriteJSON(responses_client.NewDownloadedFileResponse(file, server.policy)))
				}
			})
		}
	}
}

func (server *WebServer) handleSearchedFilesSubscriptions(group sync.WaitGroup) {
	defer group.Done()
	for {
		match := server.gossiper.FilesSearcher.MatchChan.Get()
		if match != nil {
			server.clients.Iterate(func(w clients.Writer, c *clients.Client) {
				if c.IsSubscribedTo(subscription.File) {
					for _, metadata := range match.ToSearchMetadata() {
						common.HandleError(w.WriteJSON(responses_client.NewSearchedFileResponse(metadata, server.policy)))
					}
				}
			})
		}
	}
}

func (server *WebServer) handleSubscriptionPacket(packet *packets_client.SubscribePacket, client *clients.Client, sub subscription.Sub) bool {
	//	logger.Printlnf("--->%s ---- %s ---- %s", client.IsSubscribedTo(sub), sub, packet.Subscribe)
	//logger.Printlnf("HELLOOOOOOOO")

	if !client.IsSubscribedTo(sub) && packet.Subscribe {
		client.SetSubscriptionTo(sub, true)
		if packet.WithPrevious {
			return true
		}
	} else if client.IsSubscribedTo(sub) && !packet.Subscribe {
		client.SetSubscriptionTo(sub, false)
	}
	return false
}

func (server *WebServer) handlePacket(packet *packets_client.ClientPacket, w clients.Writer, isRest bool) {

	from := "websocket"
	if isRest {
		from = "rest"
	}
	logger.Printlnf("[GUI] received %s packet <%s>", from, packet.String())

	var client *clients.Client
	if !isRest {
		client = server.clients.Get(w) // we know the client was added
	}

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

	if packet.IsSubscribeMessage() && !isRest {
		if server.handleSubscriptionPacket(packet.SubscribeMessage, client, subscription.Message) {
			for _, rumor := range server.allRumors {
				common.HandleError(w.WriteJSON(responses_client.NewRumorResponse(rumor, server.policy)))
			}
			for _, msg := range server.gossiper.Conversations.GetAll() {
				common.HandleError(w.WriteJSON(responses_client.NewPrivateResponse(msg, server.policy)))
			}
		}
		return
	}
	if packet.IsSubscribeNode() && !isRest {
		if server.handleSubscriptionPacket(packet.SubscribeNode, client, subscription.Node) {
			for _, peer := range server.gossiper.PeersSet.GetSlice() {
				common.HandleError(w.WriteJSON(responses_client.NewPeerResponse(peer, server.policy)))
			}
		}
		return
	}

	if packet.IsSubscribeOrigin() && !isRest {
		if server.handleSubscriptionPacket(packet.SubscribeOrigin, client, subscription.Origin) {
			for _, o := range server.gossiper.RoutingTable.GetOrigins() {
				if o != server.gossiper.Origin {
					common.HandleError(w.WriteJSON(responses_client.NewContactResponse(o, server.policy)))
				}
			}
		}
		return
	}

	if packet.IsSubscribeFile() && !isRest {
		if server.handleSubscriptionPacket(packet.SubscribeFile, client, subscription.File) {
			for _, file := range server.gossiper.FilesUploader.GetAllFiles() {
				common.HandleError(w.WriteJSON(responses_client.NewIndexedFileResponse(file, server.policy)))
			}
			for _, filename := range server.downloadedFiles {
				common.HandleError(w.WriteJSON(responses_client.NewDownloadedFileResponse(filename, server.policy)))
			}

			for _, search := range server.gossiper.FilesSearcher.GetAllSearches() {
				for _, match := range search.GetAllMatches() {
					for _, metadata := range match.ToSearchMetadata() {
						common.HandleError(w.WriteJSON(responses_client.NewSearchedFileResponse(metadata, server.policy)))
					}
				}
			}
		}
	}

	common.HandleAbort("an unexpected event occurred while handling ClientPacket", nil)
}

// Upgrade initial GET request to a websocket
func (server *WebServer) handleConnections(writer http.ResponseWriter, r *http.Request) {

	server.clients.Lock()

	ws, err := upgrader.Upgrade(writer, r, nil)
	if err != nil {
		common.HandleAbort("could not upgrade the connection to websocket", err)
		return
	}

	w := websocketToWriter(ws)
	server.clients.AddUnsafe(w)
	server.clients.Unlock()

	// Make sure we close the connection when the function returns
	defer func() {
		server.clients.Remove(w)
		if err := ws.Close(); err != nil {
			common.HandleAbort("closing socket failed", err)
		}
	}()

	for {
		var packet packets_client.ClientPacket
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&packet)
		if err != nil {
			if _, ok := err.(*websocket.CloseError); ok {
				common.HandleError(err)
				return
			} else {
				common.HandleAbort("error while reading json of websocket", err)
				continue
			}
		}
		// Send the newly received message to the broadcast channel
		server.clientChan <- &clients.ClientChannelElement{
			Packet: &packet,
			Writer: w,
		}
	}

}

func restToWriter(w http.ResponseWriter) clients.Writer {
	return clients.NewWriter(func(v *responses_client.ClientResponse) error {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		return json.NewEncoder(w).Encode(v)
	})
}

func websocketToWriter(ws *websocket.Conn) clients.Writer {
	return clients.NewWriter(func(v *responses_client.ClientResponse) error {
		return ws.WriteJSON(v)
	})
}

func (server *WebServer) getIdHandler(w http.ResponseWriter, r *http.Request) {
	packet := &packets_client.ClientPacket{
		GetId: &packets_client.GetIdPacket{},
	}
	server.handlePacket(packet, restToWriter(w), true)
}
