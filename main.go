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
	RequestType string `arg:"" help:"Can be add, get, del, update or delAcc" type:"request" name:"request"`
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
		} else if cli.Login.RequestType == "update" {
			if err := client.Login(3); err != nil {
				log.Fatal(err)
			}
		} else if cli.Login.RequestType == "delAcc" {
			if err := client.Login(4); err != nil {
				log.Fatal(err)
			}
		} else {
			fmt.Println("request can only be add, get or del")
			os.Exit(1)
		}
	}
}
