package gochat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// IRC client config type used during the connection process
type ClientCfg struct {
	Network string
	Nick    string
}

// IRC client type with in/out data channels
type Client struct {
	Config ClientCfg
	In     chan string
	Out    chan string
	conn   net.Conn
}

// Open a TCP connection to the specified server
func (c *Client) Connect() error {
	var err error
	c.conn, err = net.Dial("tcp", c.Config.Network)

	if err != nil {
		return err
	}

	c.In = make(chan string)
	c.Out = make(chan string)

	go c.receiver()
	go c.transmitter()

	// Send NICK and USER messages to server
	// USER message just re-uses NICK for now
	c.Nick(c.Config.Nick)
	c.User(c.Config.Nick, c.Config.Nick)

	return nil
}

// Close the TCP connection
func (c *Client) Close() {
	c.conn.Close()
	close(c.In)
	close(c.Out)
}

// Sends a NICK message to the server
func (c *Client) Nick(nick string) {
	c.Out <- "NICK " + nick
}

func (c *Client) User(nick, name string) {
	c.Out <- "USER " + nick + " 0 * :" + name
}

// Joins an IRC channel
func (c *Client) Join(channel string) {
	c.Out <- "JOIN " + channel
}

// Receiver functionality for the IRC client
// Sends raw IRC messages to the parser
func (c *Client) receiver() {
	for {
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection!")
			return
		}
		go c.ParseMessage(message)
	}
}

// Transmitter functionality for the IRC client
// Sends raw IRC messages to the server
func (c *Client) transmitter() {
	for message := range c.Out {
		fmt.Fprintf(c.conn, "%v\n", message)
		fmt.Printf("[TX] %v\n", message)
	}
}

// Creates and sends a PONG message
func (c *Client) pong(s string) {
	c.Out <- "PONG " + s
}

// Parses raw IRC messages
func (c *Client) ParseMessage(message string) {
	message = strings.TrimSpace(message)
	switch {
	case message[0] == ':':
		parts := strings.SplitN(message, " ", 2)
		if len(parts) == 2 {
			fmt.Println("[" + parts[0] + "] " + parts[1])
		} else {
			fmt.Println("=>" + message)
		}

	default:
		parts := strings.SplitN(message, " ", 2)
		cmd := strings.ToUpper(parts[0])

		switch {
		case cmd == "PING":
			if len(parts) == 2 {
				c.pong(parts[1])
			} else {
				fmt.Println("[ERR] Do not know to to PONG!")
			}
		default:
			fmt.Println(message)
		}
	}
}
