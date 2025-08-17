package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"tailscale.com/tsnet"
)

var (
	addr      = flag.String("addr", ":80", "address to listen on")
	tsAuthKey = flag.String("ts-authkey", "", "Tailscale auth key")
	hostname  = flag.String("hostname", "", "Tailscale hostname for this server")
)

func main() {
	flag.Parse()

	if *tsAuthKey == "" {
		log.Fatal("Please provide a Tailscale auth key via option --ts-authkey")
	}
	if *hostname == "" {
		log.Fatal("Please provide a Tailscale hostname via option --hostname")
	}

	srv := new(tsnet.Server)
	srv.AuthKey = *tsAuthKey
	srv.Hostname = *hostname
	defer srv.Close()
	ln, err := srv.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatal(err)
	}

	if *addr == ":443" {
		ln = tls.NewListener(ln, &tls.Config{
			GetCertificate: lc.GetCertificate,
		})
	}

	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		who, err := lc.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "<html><body><h1>Hello, world!</h1>\n")
		fmt.Fprintf(w, "<p>You are <b>%s</b> from <b>%s</b> (%s)</p>",
			html.EscapeString(who.UserProfile.LoginName),
			html.EscapeString(firstLabel(who.Node.ComputedName)),
			r.RemoteAddr)
	})))
}

func firstLabel(s string) string {
	s, _, _ = strings.Cut(s, ".")
	return s
}
