package openai

import (
	"strings"
)

func response_wrapper(response chan<- string) chan<- string {
	wrapper := make(chan string)
	go func() {
		message := ""
		for {

			received, ok := <-wrapper
			if !ok {
				return
			}

			if strings.Contains(received, "`") {
				message += received
			}
			if message != "" {
				if !strings.Contains(message, "```") {
					continue
				}
				response <- message
				message = ""
			} else {
				response <- received
			}
		}
	}()
	return wrapper
}
