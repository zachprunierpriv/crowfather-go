package message_handler

import (
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"fmt"
	"math/rand"
	"strings"
)

func Handle(message groupme.Message, oai *open_ai.OpenAIService, gms *groupme.GroupMeService) (string, error) {
	err := validateMessage(message)

	if err != nil {
		return "", err
	}

	message.Text = cleanMessage(message.Text)
	resp, err := processMessage(message, oai)

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

func processMessage(message groupme.Message, oai *open_ai.OpenAIService) (string, error) {
	threadId, err := oai.GetOrCreateThread(message.GroupId)
	lowercasedMessage := strings.ToLower(message.Text)

	if err != nil {
		return "", err
	}

	msg, err := oai.CreateMessage(message.Text, threadId)

	if err != nil {
		return "", err
	}

	if strings.Contains(lowercasedMessage, "hey crowfather") || shouldAttackRandomly() {

		run, err := oai.CreateRun(msg.ThreadID)

		if err != nil {
			return "", err
		}

		resp, err := oai.GetResponse(run, msg.ID)

		if err != nil {
			return "", err
		}

		return resp, nil
	}
	return "", nil
}
