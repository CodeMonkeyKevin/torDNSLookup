package main

import (
	"github.com/miekg/dns"
	"net"
	"os"
	"log"
	"fmt"
	"regexp"
	"bufio"
	"time"
	"go.uber.org/ratelimit"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	rateLimit()
}

func dnsQuery(domain string, dnsClient *dns.Client, config *dns.ClientConfig) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	m.RecursionDesired = true

	r, _, err := dnsClient.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil {
		fmt.Printf("%s,%s,%s\n", domain, "", err.Error())
	}

	if r.Rcode != dns.RcodeSuccess {
		fmt.Printf("%s,%s,%s\n", domain, "", "invalid response")
	}
	// Stuff must be in the answer section
	var regexIP = regexp.MustCompile(`(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)`)

	for _, a := range r.Answer {
		// fmt.Printf("%v\n", a)
		r := regexIP.FindAllString(a.String(), -1)

		if len(r) != 0 {
			fmt.Printf("%s,%s,%s\n", domain, r[0], "success")
		}
	}
}

func rateLimit() {
	config, _ := dns.ClientConfigFromFile("resolve.conf")
	c := new(dns.Client)

	file, err := os.Open("input.txt")

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// per second
	rl := ratelimit.New(1000)
	prev := time.Now()

	for scanner.Scan() {
		now := rl.Take()
		now.Sub(prev)

		domain := scanner.Text()
		dnsQuery(domain, c, config)
		prev = now
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

