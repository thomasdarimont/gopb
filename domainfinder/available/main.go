package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

var marks = map[bool]string{true: "\u2713", false: "\u2717"}

func main() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		domain := s.Text()
		fmt.Print(domain, " ")
		taken, err := exists(domain)

		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(marks[!taken])
		time.Sleep(1 * time.Second)
	}
}

func exists(domain string) (bool, error) {
	const whoisServer string = "com.whois-servers.net"

	conn, err := net.Dial("tcp", whoisServer+":43")
	if err != nil {
		return false, err
	}
	defer conn.Close()

	conn.Write([]byte(domain + "\r\n"))
	s := bufio.NewScanner(conn)

	for s.Scan() {
		if strings.Contains(strings.ToLower(s.Text()), "no match") {
			return false, nil
		}
	}
	return true, nil
}
