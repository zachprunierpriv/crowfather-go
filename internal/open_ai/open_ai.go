package open_ai

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/pagination"
	"github.com/openai/openai-go/packages/param"
	"os"
	"strings"
	"time"
)

type Config struct {
	ClientId     string
	ClientSecret string
	Uri          string
}

type OpenAIService struct {
	ThreadClient *openai.BetaThreadService
	ThreadId     string
}

var opts = []option.RequestOption{
	option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	option.WithBaseURL("https://api.openai.com/v1"),
}

func NewOpenAIService() *OpenAIService {
	threadC := openai.NewBetaThreadService(
		opts[0],
	)

	return &OpenAIService{
		ThreadClient: &threadC,
	}
}

func (oai *OpenAIService) SetThreadId() error {
	if oai.ThreadId == "" {
		thread, err := oai.CreateThread()

		if err != nil {
			return fmt.Errorf("failed to create thread %v", err)
		}

		oai.ThreadId = thread
	}
	return nil
}

func (oai *OpenAIService) GetThreadId() (string, error) {
	return oai.ThreadId, nil
}

func (oai *OpenAIService) CreateThread() (string, error) {
	t, err := oai.ThreadClient.New(context.Background(), openai.BetaThreadNewParams{}, opts...)

	if err != nil {
		return "", fmt.Errorf("failed to create thread %v", err)
	}

	return t.ID, nil
}
func (oai *OpenAIService) CreateMessage(message string) (openai.Message, error) {
	msg, err := oai.ThreadClient.Messages.New(context.Background(), oai.ThreadId, openai.BetaThreadMessageNewParams{
		Role: "user",
		Content: openai.BetaThreadMessageNewParamsContentUnion{
			OfString: param.NewOpt(message),
		},
	}, opts...)

	if err != nil {
		return openai.Message{}, fmt.Errorf("failed to create message  %v", err)
	}
	return *msg, nil
}

func (oai *OpenAIService) CreateRun(threadId string) (openai.Run, error) {
	run, err := oai.ThreadClient.Runs.New(context.Background(), threadId, openai.BetaThreadRunNewParams{
		AssistantID: os.Getenv("ASSISTANT_ID"),
	}, opts...)

	if err != nil {
		return openai.Run{}, fmt.Errorf("failed to create run %v", err)
	}

	return *run, nil
}

func (oai *OpenAIService) GetResponse(run openai.Run, messageId string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	defer cancel()

	done := false

	for !done {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("operation timed out")
		default:
			resp, err := oai.ThreadClient.Runs.Get(ctx, run.ThreadID, run.ID, opts...)

			if err != nil {
				return "", fmt.Errorf("failed to retrieve run info %v", err)
			}

			switch resp.Status {
			case openai.RunStatusInProgress, openai.RunStatusQueued:
				time.Sleep(5 * time.Second)
				continue
			case openai.RunStatusCompleted:
				done = true
			case openai.RunStatusFailed:
			case openai.RunStatusCancelled:
			case openai.RunStatusRequiresAction:
			default:
				return "", fmt.Errorf("failed to get run status with status %s", resp.Status)
			}

			messages, err := oai.ThreadClient.Messages.List(ctx, run.ThreadID, openai.BetaThreadMessageListParams{Before: param.NewOpt(messageId)}, opts...)

			if err != nil {
				return "", fmt.Errorf("failed to list messages %v", err)
			}

			if !validateResponse(messages) {
				return "", fmt.Errorf("failed to validate response")
			}

			return CleanResponse(messages.Data[0].Content[0].Text.Value), nil
		}
	}
	return "", fmt.Errorf("something went wrong")
}

func CleanResponse(s string) string {
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
