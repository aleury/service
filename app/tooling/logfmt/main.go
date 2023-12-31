// This program takes the structured log output and makes it readable.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var service string

func init() {
	flag.StringVar(&service, "service", "", "filter which service to see")
}

func main() {
	flag.Parse()
	var b strings.Builder

	service := strings.ToLower(service)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		m := make(map[string]string)
		err := json.Unmarshal([]byte(line), &m)
		if err != nil {
			if service == "" {
				fmt.Println("ERROR:", err)
				fmt.Println(line)
			}
			continue
		}

		// If a service filter was provided, check.
		if service != "" && strings.ToLower(m["service"]) != service {
			continue
		}

		// I like always having a trace id present in the logs.
		traceID := "00000000-0000-0000-0000-000000000000"
		if v, ok := m["trace_id"]; ok {
			traceID = fmt.Sprintf("%v", v)
		}

		// Build out the known portions of the log in the order
		// I want them in.
		b.Reset()
		b.WriteString(fmt.Sprintf("%s: %s: %s: %s: %s: %s: ",
			m["service"],
			m["ts"],
			m["level"],
			traceID,
			m["caller"],
			m["msg"],
		))

		// Add the rest of the keys, ignoring the known ones.
		for k, v := range m {
			switch k {
			case "service", "ts", "level", "trace_id", "caller", "msg":
				continue
			}

			// It's nice to see the key[value] in this format
			// especially since map ordering is random.
			b.WriteString(fmt.Sprintf("%s[%v]: ", k, v))
		}

		// Write the new log format, removing the last :
		out := b.String()
		fmt.Println(out[:len(out)-2])
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
