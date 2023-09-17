package openai

import (
	"testing"

	"gotest.tools/assert"
)

func TestShouldEchoTextAsIs(t *testing.T) {
	t.Parallel()
	response := make(chan string)
	wrapper := response_wrapper(response)
	wrapper <- "Hello World!"
	assert.Equal(t, <-response, "Hello World!")
}

func TestShouldPauseResponseIfCodeBegun(t *testing.T) {
	t.Parallel()
	response := make(chan string, 20)
	wrapper := response_wrapper(response)
	wrapper <- "`"
	wrapper <- "`` sth"
	wrapper <- "```"
	assert.Equal(t, <-response, "``` sth")
	assert.Equal(t, <-response, "```")
}
