# Peerster
A peer-to-peer application in the Go programming language

## How to
### Main application
To display help, from the root directory:
```
go run *.go -h
```
```
  -GUI
    	whether GUI is enabled (set to true if GUIPort != 0)
  -GUIPort uint
    	port for the GUI client (if 0, a port is randomly assigned)
  -UIPort uint
    	port for the UI client (default 8080)
  -gossipAddr value
    	ip:port for the gossiper (default "127.0.0.1:5000")
  -name string
    	name of the gossiper
  -peers value
    	comma-separated list of peers of the form ip:port
  -rtimer uint
    	route rumors sending period in seconds, 0 to disable sending of route rumors
  -simple
    	run gossiper in simple broadcast mode
```
For example:
```
go run *.go -UIPort=8080 -GUIPort=8080 -gossipAddr=localhost:5000 -name=GossiperName -peers=localhost:5001
```
### Client (command line)
First `cd` to the client directory
```
cd client
```
Then to display help:
```
go run *.go -h
```
```
  -UIPort uint
    	port for the UI client (default 8080)
  -dest string
    	destination for the private message or file request
  -file string
    	filename to be indexed by the gossiper
  -msg string
    	message to be sent
  -request string
    	request metafile of this hash
```
For example:
```
go run *.go -UIPort=8080 -msg="message content here"
```

### GUI
The Graphical User Interface (GUI) will be available at:
```
http://localhost:8080/
```
when running the main application (depending on the port for GUIPort you choose).

## Notes
This project was done during this course:

Decentralized Systems Engineering (CS-438 @EPFL) Fall 2018 (https://dedis.epfl.ch)
