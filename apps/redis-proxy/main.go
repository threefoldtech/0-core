package main

import (
	"os"

	"github.com/codegangsta/cli"
	logging "github.com/op/go-logging"
	"github.com/threefoldtech/0-core/base/utils"
)

const (
	/*SupportOrganization all people of this organization are allowed to have full control
	of the machine, if support flag is set.

	--- With great power comes great responsibility ---
		      __
	     /  l
	   .'   :               __.....__..._  ____
	  /  /   \          _.-"        "-.  ""    "-.
	 (`-: .---:    .--.'          _....J.         "-.
	  """y     \,.'    \  __..--""       `+""--.     `.
	    :     .'/    .-"""-. _.            `.   "-.    `._.._
	    ;  _.'.'  .-j       `.               \     "-.   "-._`.
	    :    / .-" :          \  `-.          `-      "-.      \
	     ;  /.'    ;          :;               ."        \      `,
	     :_:/      ::\        ;:     (        /   .-"   .')      ;
	       ;-"      ; "-.    /  ;           .^. .'    .' /    .-"
	      /     .-  :    `. '.  : .- / __.-j.'.'   .-"  /.---'
	     /  /      `,\.  .'   "":'  /-"   .'       \__.'
	    :  :         ,\""       ; .'    .'      .-""
	   _J  ;         ; `.      /.'    _/    \.-"
	  /  "-:        /"--.b-..-'     .'       ;
	 /     /  ""-..'            .--'.-'/  ,  :
	:`.   :     / :             `-i" ,',_:  _ \
	:  \  '._  :__;             .'.-"; ; ; j `.l
	 \  \          "-._         `"  :_/ :_/
	  `.;\             "-._
	    :_"-._             "-.
	      `.  l "-.     )     `.
	        ""^--""^-. :        \
	                  ";         \
	                  :           `._
	                  ; /    \ `._   ""---.
	                 / /   _      `.--.__.'
	                : :   / ;  :".  \
	                ; ;  :  :  ;  `. `.
	               /  ;  :   ; :    `. `.
	              /  /:  ;   :  ;     "-'
	             :_.' ;  ;    ; :
	                 /  /     :_l
	                 `-'
	*/
	SupportOrganization = "threefold.sysadmin"
)

var (
	log = logging.MustGetLogger("redis-proxy")
)

func main() {
	app := cli.NewApp()

	app.Name = "redis-proxy"
	app.Usage = "add tls and custom authentication on top of redis"
	app.Version = "1.0"

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "organization, o",
			Usage: "IYO organization that has to be valid in the jwt claims, if not provided, it will be parsed from kernel cmdline, otherwise no authentication will be applied",
		},
		cli.StringFlag{
			Name:  "listen, l",
			Value: "0.0.0.0:6379",
			Usage: "listing address (default: 0.0.0.0:6379)",
		},
		cli.StringFlag{
			Name:  "redis, r",
			Value: "/var/run/redis.sock",
			Usage: "redis unix socket to proxy",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
	}

	app.Before = func(ctx *cli.Context) error {
		if ctx.GlobalBool("debug") {
			logging.SetLevel(logging.DEBUG, "")
		} else {
			logging.SetLevel(logging.INFO, "")
		}

		return nil
	}

	app.Action = func(ctx *cli.Context) error {
		var organizations []string
		/*
			if we are running in development mode, we should accept all connections
			with no jwt validation required.
		*/
		if !utils.GetKernelOptions().Is("development") {
			/*
				otherwise, we only accept connections from the given organization
				either the one given via command line, if not given we use the ones
				given to the kernel
			*/
			organizations = ctx.StringSlice("organization")
			if len(organizations) == 0 {
				if orgs, ok := utils.GetKernelOptions().Get("organization"); ok {
					organizations = orgs
				}
			}

			if utils.GetKernelOptions().Is("support") {
				//and finally we add our spiderman orgnaization
				organizations = append(organizations, SupportOrganization)
			}
		}

		return Proxy(ctx.String("listen"), ctx.String("redis"), organizations)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
