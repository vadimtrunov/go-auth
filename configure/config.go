package configure

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Read file from "path" and parse it into the config sturcture
func process(path string) (cnf map[string]string, fatal error) {
	file, fatal := os.Open(path)
	if fatal != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	lineNumber := 1
	cnf = make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			parcedLine := strings.Split(line, "=")
			if len(parcedLine) > 2 {
				panic(fmt.Sprintf("Invalid configuration at line %d", lineNumber))
			}

			cnf[parcedLine[0]] = parcedLine[1]
		}

		lineNumber++

		if err != nil {
			break
		}
	}
	return
}

// HTTPServer configurates server according config file. Returns new http.Server
func HTTPServer(path string) (*http.Server, error) {
	cnf, err := process(path)
	if err != nil {
		return nil, err
	}
	var server http.Server
	for key, value := range cnf {
		switch key {
		case "Addr":
			server.Addr = value
		case "ReadTimeout":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			server.ReadTimeout = time.Millisecond * time.Duration(v)
		case "WriteTimeout":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			server.WriteTimeout = time.Millisecond * time.Duration(v)
		case "MaxHeaderBytes":
			v, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			server.MaxHeaderBytes = v
		}
	}
	return &server, nil
}
