package message_handler

import (
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"fmt"
	"github.com/openai/openai-go"
	"math/rand"
	"strings"
)

func Handle(message groupme.Message, oai *open_ai.OpenAIService, gms *groupme.GroupMeService, assistantID string) (string, error) {
	err := validateMessage(message)

	if err != nil {
		return "", err
	}

	message.Text = cleanMessage(message.Text)
	resp, err := processMessage(message, oai, assistantID)

	if err != nil || resp == "" {
		return "", err
	}

	if message.UserId != "" {
		_, err = gms.SendMessage(message, resp)

		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func processMessage(message groupme.Message, oai *open_ai.OpenAIService, assistantID string) (string, error) {
	lowercasedMessage := strings.ToLower(message.Text)

	msg, err := addMessageToThread(message, oai)

	if err != nil {
		return "", err
	}

	if strings.Contains(lowercasedMessage, "hey crowfather") || shouldAttackRandomly() {
		return respondToThread(msg, oai, assistantID)
	}
	return "", nil
}

func addMessageToThread(message groupme.Message, oai *open_ai.OpenAIService) (openai.Message, error) {
	threadId, err := oai.GetOrCreateThread(message.GroupId)

	if err != nil {
		return openai.Message{}, err
	}

	msg, err := oai.CreateMessage(message.Text, threadId)

	if err != nil {
		return openai.Message{}, err
	}
	return msg, nil
}

func respondToThread(message openai.Message, oai *open_ai.OpenAIService, assistantID string) (string, error) {
	run, err := oai.CreateRun(message.ThreadID, assistantID)

	if err != nil {
		return "", err
	}

	resp, err := oai.GetResponse(run, message.ID)

	if err != nil {
		return "", err
	}

	return resp, nil
}

func validateMessage(message groupme.Message) error {
	if message.SenderType != "user" {
		return fmt.Errorf("message is not from a user")
	}

	return nil
}

func shouldAttackRandomly() bool {
	firstRand := rand.Intn(20)
	secondRand := rand.Intn(20)

	return firstRand == secondRand
}

func cleanMessage(message string) string {
	cleanedMessage := strings.TrimPrefix(message, "hey crowfather")
	cleanedMessage = strings.TrimPrefix(cleanedMessage, ",")
	cleanedMessage = strings.TrimSpace(cleanedMessage)

	return cleanedMessage
}
