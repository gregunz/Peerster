package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/packets"
	"github.com/gregunz/Peerster/utils"
	"io/ioutil"
	"net/http"
)

type WebServer struct {
	gossiper *gossiper.Gossiper
}

func NewWebServer(g *gossiper.Gossiper) *WebServer {
	fmt.Println("Creating WebServer")
	return &WebServer{
		gossiper: g,
	}
}

func (server *WebServer) Start() {
	router := mux.NewRouter()

	// GET
	router.PathPrefix("/gui/").Handler(http.StripPrefix("/gui/", http.FileServer(http.Dir("./gui/dist"))))
	router.HandleFunc("/id", server.getIdHandler).Methods("GET")
	router.HandleFunc("/node", server.getNodeHandler).Methods("GET")
	router.HandleFunc("/message", server.getMessageHandler).Methods("GET")

	// POST
	router.HandleFunc("/node", server.postNodeHandler).Methods("POST")
	router.HandleFunc("/message", server.postMessageHandler).Methods("POST")

	http.ListenAndServe(fmt.Sprintf(":%d", server.gossiper.GUIPort), router)
}

func (server *WebServer) getIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.Name())
}

func (server *WebServer) getNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.PeersSet().ToStrings())
}

func (server *WebServer) postNodeHandler(w http.ResponseWriter, r *http.Request) {

	var jsonMap map[string]interface{}
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	ipPort, ok := jsonMap["peer"].(string)

	if !ok {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	udpAddr := utils.IpPortToUDPAddr(ipPort)
	if udpAddr == nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	server.gossiper.PeersSet().AddIpPort(ipPort)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}

func (server *WebServer) getMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.VectorClock().GetLatestMessages())
}

func (server *WebServer) postMessageHandler(w http.ResponseWriter, r *http.Request) {

	var jsonMap map[string]interface{}
	data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	text, ok := jsonMap["message"].(string)

	if !ok {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	clientPacket := packets.ClientPacket{
		Message: text,
	}
	server.gossiper.HandleClient(&clientPacket)

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write([]byte("{}"))
}
