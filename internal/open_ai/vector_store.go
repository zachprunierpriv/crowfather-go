package open_ai

import (
	"bytes"
	"context"
	"fmt"

	"github.com/openai/openai-go"
)

// CreateVectorStore creates a new OpenAI vector store and returns its ID.
func (oai *OpenAIService) CreateVectorStore(ctx context.Context, name string) (string, error) {
	client := openai.NewClient(oai.Options...)
	vs, err := client.VectorStores.New(ctx, openai.VectorStoreNewParams{
		Name: openai.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to create vector store: %w", err)
	}
	return vs.ID, nil
}

// UploadFilesToVectorStore uploads the provided documents to a vector store in a
// single batch and polls until all files are processed.
// docs is a map of filename → file content (Markdown bytes).
func (oai *OpenAIService) UploadFilesToVectorStore(ctx context.Context, vsID string, docs map[string][]byte) error {
	if len(docs) == 0 {
		return nil
	}

	client := openai.NewClient(oai.Options...)

	fileParams := make([]openai.FileNewParams, 0, len(docs))
	for _, content := range docs {
		fileParams = append(fileParams, openai.FileNewParams{
			File:    bytes.NewReader(content),
			Purpose: openai.FilePurposeAssistants,
		})
	}

	batch, err := client.VectorStores.FileBatches.UploadAndPoll(
		ctx, vsID, fileParams, nil, 5000, oai.Options...,
	)
	if err != nil {
		return fmt.Errorf("failed to upload files to vector store: %w", err)
	}

	if batch.Status == "failed" {
		return fmt.Errorf("vector store file batch failed: %d file(s) errored", batch.FileCounts.Failed)
	}

	return nil
}

// AttachVectorStoreToAssistant updates the assistant to use the given vector store
// for file_search. It also ensures the file_search tool is enabled.
func (oai *OpenAIService) AttachVectorStoreToAssistant(ctx context.Context, assistantID, vsID string) error {
	client := openai.NewClient(oai.Options...)
	_, err := client.Beta.Assistants.Update(ctx, assistantID, openai.BetaAssistantUpdateParams{
		Tools: []openai.AssistantToolUnionParam{
			{OfFileSearch: &openai.FileSearchToolParam{}},
		},
		ToolResources: openai.BetaAssistantUpdateParamsToolResources{
			FileSearch: openai.BetaAssistantUpdateParamsToolResourcesFileSearch{
				VectorStoreIDs: []string{vsID},
			},
		},
	}, oai.Options...)
	if err != nil {
		return fmt.Errorf("failed to attach vector store %s to assistant %s: %w", vsID, assistantID, err)
	}
	return nil
}

// DeleteVectorStore deletes a vector store by ID.
func (oai *OpenAIService) DeleteVectorStore(ctx context.Context, vsID string) error {
	client := openai.NewClient(oai.Options...)
	_, err := client.VectorStores.Delete(ctx, vsID, oai.Options...)
	if err != nil {
		return fmt.Errorf("failed to delete vector store %s: %w", vsID, err)
	}
	return nil
}
