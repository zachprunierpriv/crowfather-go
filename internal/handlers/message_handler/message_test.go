package message_handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Bug 3: cleanMessage must strip the prefix case-insensitively and handle a
// space before the trailing comma.

func TestCleanMessageCaseInsensitivePrefix(t *testing.T) {
	// Original: strings.TrimPrefix(message, "hey crowfather") is case-sensitive,
	// so a capitalised trigger is not stripped before being sent to OpenAI.
	assert.Equal(t, "what time is it?", cleanMessage("Hey Crowfather, what time is it?"))
}

func TestCleanMessageSpaceBeforeComma(t *testing.T) {
	// Original: TrimPrefix(",") silently skips the comma when whitespace precedes it.
	assert.Equal(t, "what time is it?", cleanMessage("hey crowfather , what time is it?"))
}
