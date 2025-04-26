package message

import (
	"crowfather/internal/groupme"
	"crowfather/internal/open_ai"
	"strings"
)

func MessageHandler(message groupme.Message, oai *open_ai.OpenAIService, gms *groupme.GroupMeService) (string, error) {
	lowercasedMessage := strings.ToLower(message.Text)

	if strings.Contains(lowercasedMessage, "hey crowfather") {
		cleanedMessage := strings.TrimPrefix(message.Text, "hey crowfather")
		cleanedMessage = strings.TrimPrefix(cleanedMessage, ",")
		cleanedMessage = strings.TrimSpace(cleanedMessage)

		message.Text = cleanedMessage

		resp, err := HandleMessage(message, oai)

		if err != nil {
			return "", err
		}

		if message.UserId != "" {
			_, err = gms.SendMessage(message, resp)

			if err != nil {
				return "", err
			}
		}

		return resp, nil
	}

	return "", nil
}

func HandleMessage(message groupme.Message, oai *open_ai.OpenAIService) (string, error) {
	err := oai.SetThreadId()

	if err != nil {
		return "", err
	}

	msg, err := oai.CreateMessage(message.Text)

	if err != nil {
		return "", err
	}

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
