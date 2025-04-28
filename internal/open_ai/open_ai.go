package open_ai

import (
	"context"
	"crowfather/internal/config"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/pagination"
	"github.com/openai/openai-go/packages/param"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ClientId     string
	ClientSecret string
	Uri          string
}

type OpenAIService struct {
	ThreadClient *openai.BetaThreadService
	ThreadIds    map[string]string
	Config       *config.OpenAIConfig
	Options      []option.RequestOption
	mu           sync.RWMutex
}

func NewOpenAIService(config *config.OpenAIConfig) *OpenAIService {
	var opts = []option.RequestOption{
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.BaseURL),
	}

	threadC := openai.NewBetaThreadService(
		opts[0],
	)

	return &OpenAIService{
		ThreadClient: &threadC,
		Config:       config,
		Options:      opts,
		ThreadIds:    make(map[string]string),
	}
}

func (oai *OpenAIService) GetOrCreateThread(contextID string) (string, error) {
	oai.mu.RLock()
	defer oai.mu.RUnlock()

	if threadId, exists := oai.ThreadIds[contextID]; exists {
		return threadId, nil
	}

	threadId, err := oai.CreateThread()

	if err != nil {
		return "", err
	}

	oai.ThreadIds[contextID] = threadId

	return threadId, nil
}

func (oai *OpenAIService) GetThreadId(contextID string) string {
	return oai.ThreadIds[contextID]
}

func (oai *OpenAIService) CreateThread() (string, error) {
	t, err := oai.ThreadClient.New(context.Background(), openai.BetaThreadNewParams{}, oai.Options...)

	if err != nil {
		return "", fmt.Errorf("failed to create thread %v", err)
	}

	return t.ID, nil
}
func (oai *OpenAIService) CreateMessage(message string, threadId string) (openai.Message, error) {
	msg, err := oai.ThreadClient.Messages.New(context.Background(), threadId, openai.BetaThreadMessageNewParams{
		Role: "user",
		Content: openai.BetaThreadMessageNewParamsContentUnion{
			OfString: param.NewOpt(message),
		},
	}, oai.Options...)

	if err != nil {
		return openai.Message{}, fmt.Errorf("failed to create message  %v", err)
	}
	return *msg, nil
}

func (oai *OpenAIService) CreateRun(threadId string, assistantID string) (openai.Run, error) {
	run, err := oai.ThreadClient.Runs.New(context.Background(), threadId, openai.BetaThreadRunNewParams{
		AssistantID: assistantID,
	}, oai.Options...)

	if err != nil {
		return openai.Run{}, fmt.Errorf("failed to create run %v", err)
	}

	return *run, nil
}

func (oai *OpenAIService) GetResponse(run openai.Run, messageId string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), oai.Config.Timeout)

	defer cancel()

	done := false

	for !done {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("operation timed out after 60 seconds")
		default:
			resp, err := oai.ThreadClient.Runs.Get(ctx, run.ThreadID, run.ID, oai.Options...)

			if err != nil {
				return "", fmt.Errorf("failed to retrieve run info %v", err)
			}

			switch resp.Status {
			case openai.RunStatusInProgress, openai.RunStatusQueued:
				time.Sleep(5 * time.Second)
				continue
			case openai.RunStatusCompleted:
				return oai.getCompletedResponse(ctx, run.ThreadID, messageId)
			case openai.RunStatusFailed:
			case openai.RunStatusCancelled:
			case openai.RunStatusRequiresAction:
			default:
				return "", fmt.Errorf("failed to get run status with status %s", resp.Status)
			}
		}
	}
	return "", fmt.Errorf("something went wrong")
}

func cleanResponse(s string) string {
	cleaned := strings.TrimSpace(s)
	cleaned = strings.ReplaceAll(cleaned, "\n\n", "\n")
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	return cleaned
}

func validateResponse(response *pagination.CursorPage[openai.Message]) bool {
	if len(response.Data) == 0 {
		return false
	}

	if len(response.Data[0].Content) == 0 {
		return false
	}

	if response.Data[0].Content[0].Text.Value == "" {
		return false
	}

	return true
}

func (oai *OpenAIService) getCompletedResponse(ctx context.Context, threadId string, messageId string) (string, error) {
	messages, err := oai.ThreadClient.Messages.List(ctx, threadId, openai.BetaThreadMessageListParams{Before: param.NewOpt(messageId)}, oai.Options...)

	if err != nil {
		return "", fmt.Errorf("failed to list messages %v", err)
	}

	if !validateResponse(messages) {
		return "", fmt.Errorf("failed to validate response")
	}
	assistantMessage, err := oai.GetAssitantMessage(messages.Data)

	if err != nil {
		return "", err
	}

	return cleanResponse(assistantMessage.Content[0].Text.Value), nil
}

func (oai *OpenAIService) GetAssitantMessage(messages []openai.Message) (openai.Message, error) {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "assistant" {
			return messages[i], nil
		}
	}
	return openai.Message{}, fmt.Errorf("no assistant message found")
}
