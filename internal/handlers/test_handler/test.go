package test_handler

import (
	"crowfather/internal/open_ai"
)

func Handle(message string, oai *open_ai.OpenAIService, assistantID string) (string, error) {
	resp, err := processMessage(message, oai, assistantID)

	if err != nil {
		return "", err
	}

	return resp, nil
}

func processMessage(message string, oai *open_ai.OpenAIService, assistantID string) (string, error) {
	threadId, err := oai.GetOrCreateThread("temp")

	if err != nil {
		return "", err
	}

	msg, err := oai.CreateMessage(message, threadId)

	if err != nil {
		return "", err
	}

	run, err := oai.CreateRun(msg.ThreadID, assistantID)

	if err != nil {
		return "", err
	}

	resp, err := oai.GetResponse(run, msg.ID)

	if err != nil {
		return "", err
	}

	return resp, nil
}
