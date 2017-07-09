package main

import(
	"net/http"
	"strconv"
	"os"
	"bufio"
	"github.com/gorilla/websocket"
)

var peers []*websocket.Conn

func main() {
    file, err := os.Open("peers.txt")
    defer file.Close()
    if err != nil { println(err.Error()) }
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
		//peer := scanner.Text()
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial("ws://"+scanner.Text()+"/ws", nil)
		if err != nil {
			//println(err.Error())
		} else {
			go wsListen(conn)
			if len(peers) > 1 { break }
		}
    }
    if err := scanner.Err(); err != nil {
        println(err.Error())
    }

    http.HandleFunc("/ws", wsHandler)
	//handler := http.FileServer(http.Dir("public"))
    //http.Handle("/", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "public/index.html")
	})

    port := 10742
    println("Listening on port " + strconv.Itoa(port))
    err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
    for err != nil {
        println(err.Error())
		port++
		println("Listening on port " + strconv.Itoa(port))
		err = http.ListenAndServe(":"+strconv.Itoa(port), nil)
    }
}
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool {return true}
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        println(err.Error())
        return
    }
	wsListen(conn)

}
func wsListen(conn *websocket.Conn) {
	peers = append(peers, conn)
	println(conn.RemoteAddr().String() + " Connected")
	println("Peers: " + strconv.Itoa(len(peers)))
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			println(conn.RemoteAddr().String() + " Disconnected")
			removePeer(conn)
			println("Peers: " + strconv.Itoa(len(peers)))
			return
		}
		println(string(p))
		if err = conn.WriteMessage(messageType, p); err != nil {
			println(err.Error())
			return
		}
	}
}

func removePeer(conn *websocket.Conn) bool {
    for i, v := range peers {
        if (v == conn) {
			peers = append(peers[:i], peers[i+1:]...)
            return true
        }
    }
    return false
}