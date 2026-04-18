package claude

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/superplanehq/superplane/pkg/configuration"
	"github.com/superplanehq/superplane/pkg/core"
)

const PromptAIPayloadType = "claude.promptAI"

type PromptAI struct{}

type PromptAISpec struct {
	Model string `json:"model"`
}

type PromptAIPayload struct {
	Text string `json:"text"`
}

func (c *PromptAI) Name() string {
	return "claude.promptAI"
}

func (c *PromptAI) Label() string {
	return "Prompt AI"
}

func (c *PromptAI) Description() string {
	return "Automatically analyzes upstream alert data from Grafana using Claude"
}

func (c *PromptAI) Documentation() string {
	return "Reads the upstream Grafana alert payload from ctx.Data and sends it to Claude for analysis."
}

func (c *PromptAI) Icon() string {
	return "message-circle"
}

func (c *PromptAI) Color() string {
	return "purple"
}

func (c *PromptAI) ExampleOutput() map[string]any {
	return map[string]any{
		"text": "The alert indicates high CPU usage on the worker nodes...",
	}
}

func (c *PromptAI) OutputChannels(configuration any) []core.OutputChannel {
	return []core.OutputChannel{core.DefaultOutputChannel}
}

func (c *PromptAI) Configuration() []configuration.Field {
	return []configuration.Field{
		{
			Name:        "model",
			Label:       "Model",
			Type:        configuration.FieldTypeIntegrationResource,
			Required:    true,
			Default:     "claude-opus-4-6",
			Placeholder: "Select a Claude model",
			TypeOptions: &configuration.TypeOptions{
				Resource: &configuration.ResourceTypeOptions{
					Type: "model",
				},
			},
		},
	}
}

func (c *PromptAI) Setup(ctx core.SetupContext) error {
	spec := PromptAISpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}
	if spec.Model == "" {
		return fmt.Errorf("model is required")
	}
	return nil
}

func (c *PromptAI) Execute(ctx core.ExecutionContext) error {
	spec := PromptAISpec{}
	if err := mapstructure.Decode(ctx.Configuration, &spec); err != nil {
		return fmt.Errorf("failed to decode configuration: %v", err)
	}
	if spec.Model == "" {
		return fmt.Errorf("model is required")
	}

	const systemPrompt = "You are an SRE assistant. Analyze the following alert data and provide a concise summary of what went wrong, the likely cause, and suggested remediation steps."

	combined := systemPrompt
	if ctx.Data != nil {
		alertJSON, err := json.Marshal(ctx.Data)
		if err == nil {
			combined = fmt.Sprintf("%s\n\nGrafana Alert Payload:\n%s", systemPrompt, string(alertJSON))
		}
	}

	client, err := NewClient(ctx.HTTP, ctx.Integration)
	if err != nil {
		return err
	}

	resp, err := client.CreateMessage(CreateMessageRequest{
		Model:     spec.Model,
		MaxTokens: 4096,
		Messages:  []Message{{Role: "user", Content: combined}},
	})
	if err != nil {
		return fmt.Errorf("claude api error: %w", err)
	}

	return ctx.ExecutionState.Emit(
		core.DefaultOutputChannel.Name,
		PromptAIPayloadType,
		[]any{PromptAIPayload{Text: extractMessageText(resp)}},
	)
}

func (c *PromptAI) Cancel(ctx core.ExecutionContext) error {
	return nil
}

func (c *PromptAI) Actions() []core.Action {
	return []core.Action{}
}

func (c *PromptAI) HandleAction(ctx core.ActionContext) error {
	return nil
}

func (c *PromptAI) HandleWebhook(ctx core.WebhookRequestContext) (int, *core.WebhookResponseBody, error) {
	return http.StatusOK, nil, nil
}

func (c *PromptAI) ProcessQueueItem(ctx core.ProcessQueueContext) (*uuid.UUID, error) {
	return ctx.DefaultProcessing()
}

func (c *PromptAI) Cleanup(ctx core.SetupContext) error {
	return nil
}
