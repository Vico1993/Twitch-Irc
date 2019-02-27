package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type twitchLine struct {
	Badges      map[string]bool `mapstructure:"@badges"`
	Color       string          `mapstructure:"color"`
	DisplayName string          `mapstructure:"display-name"`
	Emotes      string          `mapstructure:"emotes"`
	Flags       string          `mapstructure:"flags"`
	ID          string          `mapstructure:"id"`
	Mod         string          `mapstructure:"mod"`
	RoomID      string          `mapstructure:"room-id"`
	Subscriber  string          `mapstructure:"subscriber"`
	Timestamp   int64           `mapstructure:"tmi-sent-ts"`
	Turbo       int             `mapstructure:"turbo"`
	UserID      int             `mapstructure:"user-id"`
	UserType    string          `mapstructure:"user-type"`
	Message     string          `mapstructure:"message"`
}

func getMessage(message string, CHANNEL string) string {
	// Bytes to removes the first part of the message.
	var b = []byte(message)
	b = b[bytes.IndexRune(b, '#'):]

	return strings.TrimPrefix(string(b), CHANNEL)
}

func main() {

	// Get ViperConfig ( IMPROVE )
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.twitchIRC")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s \n", err))
	}

	// GET INFO inside config.json
	CHANNEL := viper.GetString("CHANNEL")
	PASS := viper.GetString("PASS")
	USERNAME := viper.GetString("USERNAME")

	var conn net.Conn

	dialer := &net.Dialer{
		KeepAlive: time.Second * 10,
	}

	conn, err = tls.DialWithDialer(dialer, "tcp", "irc.chat.twitch.tv:443", &tls.Config{})
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		return
	}

	conn.Write([]byte("PASS " + PASS + "\r\n"))
	conn.Write([]byte("NICK " + USERNAME + "\r\n"))
	conn.Write([]byte("CAP REQ :twitch.tv/tags\r\n"))
	// conn.Write([]byte("CAP REQ :twitch.tv/commands\r\n"))
	conn.Write([]byte("JOIN " + CHANNEL + "\r\n"))

	onComma := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		for i := 0; i < len(data); i++ {
			if data[i] == ';' {
				return i + 1, data[:i], nil
			}
		}

		return 0, data, bufio.ErrFinalToken
	}

	for {
		line, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("ERROR readline: " + err.Error())
			return
		}

		scanner := bufio.NewScanner(strings.NewReader(line))
		scanner.Split(onComma)

		twl := twitchLine{}
		// Initial declaration
		m := map[string]string{}

		for scanner.Scan() {
			l := scanner.Text()

			// If twitch send US PING need to reply. PONG
			if strings.Contains(l, "PING") {
				pong := strings.Replace(l, "PING", "PONG", 1)
				conn.Write([]byte(pong + "\r\n"))
			}

			if strings.Contains(l, "=") {
				split := strings.Split(l, "=")

				m[split[0]] = split[1]
			} else {
				fmt.Printf(l)
			}
		}

		if m["user-id"] != "" {
			if m["user-type"] != "" {
				s := strings.SplitN(m["user-type"], ":", 2)

				m["message"] = s[1]
				m["user-type"] = s[0]
			}
			mapstructure.Decode(m, &twl)
			message := strings.Replace(m["message"], strings.ToLower(twl.DisplayName)+`!`+strings.ToLower(twl.DisplayName)+`@`+strings.ToLower(twl.DisplayName)+` .tmi.twitch.tv PRIVMSG`+CHANNEL, "", -1)

			fmt.Printf(twl.DisplayName + ": " + getMessage(message, CHANNEL))
			fmt.Println("- - - - - - - - - - - - - - - - - - - - - -")
			fmt.Println("Color :" + twl.Color)
			fmt.Println("DisplayName :" + twl.DisplayName)
			fmt.Println("Emotes :" + twl.Emotes)
			fmt.Println("Mod :" + twl.Mod)
			fmt.Println("User - Type :" + twl.UserType)
			fmt.Println("- - - - - - - - - - - - - - - - - - - - - -")
		}
	}

}
