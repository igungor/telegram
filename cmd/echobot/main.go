package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/igungor/tlbot"
)

// flags
var (
	token   = flag.String("token", "", "telegram bot token")
	webhook = flag.String("webhook", "", "webhook url")
	host    = flag.String("host", "127.0.0.1", "host to listen to")
	port    = flag.String("port", "1986", "port to listen to")
	debug   = flag.Bool("d", false, "debug mode (*very* verbose)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "testbot is an echo server for testing Telegram bots\n\n")
	fmt.Fprintf(os.Stderr, "usage:\n")
	fmt.Fprintf(os.Stderr, "  testbot -token <insert-your-telegrambot-token> -url <insert-your-webhook-url>\n\n")
	fmt.Fprintf(os.Stderr, "flags:\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("testbot: ")
	flag.Usage = usage
	flag.Parse()

	if *webhook == "" {
		log.Printf("missing webhook parameter\n\n")
		flag.Usage()
	}

	if *token == "" {
		log.Printf("missing token parameter\n\n")
		flag.Usage()
	}

	b := tlbot.New(*token)
	err := b.SetWebhook(*webhook)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := b.Listen(net.JoinHostPort(*host, *port))
	if err != nil {
		log.Fatal(err)
	}

	// spew.Dump uses String() method if a type implements Stringer interface.
	// Since Message type is a Stringer, enable more verbose output by
	// disabling this behaviour.
	if *debug {
		spew.Config.DisableMethods = true
	}
	for msg := range ch {
		spew.Dump(msg)
		err := b.SendMessage(msg.From, msg.Text, tlbot.ModeNone, false, nil)
		if err != nil {
			log.Println(err)
		}
	}
}
