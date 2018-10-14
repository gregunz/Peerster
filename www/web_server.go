package www

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gregunz/Peerster/common"
	"github.com/gregunz/Peerster/gossiper"
	"github.com/gregunz/Peerster/models/packets"
	"io"
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
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./gui/dist")))
	router.HandleFunc("/id", server.GetIdHandler).Methods("GET")
	router.HandleFunc("/node", server.GetNodeHandler).Methods("GET")
	router.HandleFunc("/message", server.GetMessageHandler).Methods("GET")

	// POST
	router.HandleFunc("/node", server.PostNodeHandler).Methods("POST")
	router.HandleFunc("/message", server.PostMessageHandler).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func (server *WebServer) GetIdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.Name())
}

func (server *WebServer) GetNodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.PeersSet().ToStrings())
}

func (server *WebServer) PostNodeHandler(w http.ResponseWriter, r *http.Request) {

}

func (server *WebServer) GetMessageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(server.gossiper.VectorClock().GetLatestMessages())
}

func (server *WebServer) PostMessageHandler(w http.ResponseWriter, r *http.Request) {
	var clientPacket packets.ClientPacket
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		common.HandleAbort("error when reading post request", err)
	}
	if err := r.Body.Close(); err != nil {
		common.HandleAbort("error when closing body of post request", err)
	}
	if err := json.Unmarshal(body, &clientPacket); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	server.gossiper.HandleClient(&clientPacket)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode("success"); err != nil {
		panic(err)
	}
}

/*
func StringHandler(s string) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, s)
	}
}
func JsonHandler(val ToJson) func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(val.toJson())
	}
}

type ToJson interface {
	toJson() string
}
*/