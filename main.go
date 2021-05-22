package main

import (
	"fmt"
	"log"
	"mpmclient/client"
	"os"

	"github.com/alecthomas/kong"
)

type SignupCmd struct {
}

type LoginCmd struct {
	RequestType string `arg:"" help:"Can be add, get or del" type:"request" name:"request"`
}

var cli struct {
	Signup SignupCmd `cmd:"" help:"Signup for the mpm service"`
	Login  LoginCmd  `cmd:"" help:"Login to mpm service"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("mpmclient"),
		kong.Description("Client for the mpm server"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))
	switch ctx.Command() {
	case "signup":
		if err := client.SignUp(); err != nil {
			log.Fatal(err)
		}
	case "login <request>":
		if cli.Login.RequestType == "add" {
			if err := client.Login(0); err != nil {
				log.Fatal(err)
			}
		} else if cli.Login.RequestType == "get" {
			if err := client.Login(1); err != nil {
				log.Fatal(err)
			}
		} else if cli.Login.RequestType == "del" {
			if err := client.Login(2); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("request can only be add, get or del")
			os.Exit(1)
		}
	}
}
