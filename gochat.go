package gochat

import (
	"bufio"
	"errors"
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

// IRC message
type Message struct {
	Raw     string
	Prefix  string
	Command string
	Params  string
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
		data, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection!")
			return
		}

		go func() {
			message, err := c.ParseMessage(data)
			if err != nil {
				fmt.Println(err)
			} else {
				switch {
				case message.Command == "PING":
					c.pong(message.Prefix)
				default:
					fmt.Printf("%v %v\n", message.Prefix, message.Command)
				}
			}
		}()
	}
}

// Transmitter functionality for the IRC client
// Sends raw IRC messages to the server
func (c *Client) transmitter() {
	for message := range c.Out {
		fmt.Fprintf(c.conn, "%v\n", message)
	}
}

// Creates and sends a PONG message
func (c *Client) pong(s string) {
	c.Out <- "PONG " + s
}

// Parses raw IRC messages
func (c *Client) ParseMessage(data string) (*Message, error) {
	if len(data) == 0 {
		return nil, errors.New("Empty IRC message!")
	}

	message := Message{Raw: data}

	if data[0] == ':' {
		if end := strings.Index(data, " "); end != -1 {
			message.Prefix = data[1:end]
			data = data[end:]
		} else {
			return nil, errors.New("Expected a command or parameter after the prefix!")
		}
	}

	if end := strings.IndexAny(data, " \n"); end != -1 {
		message.Command = data[:end]
		data = data[end:]
	} else {
		return nil, errors.New("IRC message not terminated")
	}

	if len(data) > 0 {
		message.Params = data
	}

	return &message, nil
}
