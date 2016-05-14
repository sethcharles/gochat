# Gochat IRC Client

Gochat is a Go software package that implements RFC 1459 and RFC 2812
standards for an IRC client.

## IRC Bot Example 

### Create the client
    client := gochat.Client{}
    client.Config = gochat.ClientCfg{"irc.freenode.net:6667", "gochat-bot"}

### Connect to the network
    err := client.Connect()
    
    if err != nil {
        fmt.Println("Could not connect")
        return
    }

### Join a channel
    client.Join("#go-nuts")
