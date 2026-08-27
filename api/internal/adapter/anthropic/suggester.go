// Package anthropic implementa domain.SuggestionProposer contra la API de
// Claude. La sugerencia nunca se aplica sola: el worker solo la guarda como
// 'pending' en ai_suggestion, y un coach humano tiene que aprobarla desde
// el panel (ver docs/PHASES.md fase 9).
package anthropic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"

	"myvibesfit/api/internal/domain"
)

const systemPrompt = `Eres el asistente de un coach de gimnasio. Se te da un resumen
estructurado (JSON) del progreso reciente de un cliente. Debes proponer UNA
sola sugerencia de ajuste al plan, basada solo en los datos del resumen —
nunca inventes datos que no esten ahi. La sugerencia SIEMPRE la revisa el
coach antes de aplicarse: nunca digas que ya se aplico, solo proponla.
Usa la herramienta propose_suggestion para responder. La razon debe ser una
sola frase en español, clara para que el coach decida rapido.`

var proposeSuggestionTool = anthropic.ToolParam{
	Name:        "propose_suggestion",
	Description: anthropic.String("Propone un ajuste al plan de entrenamiento del cliente, para que el coach lo apruebe o rechace."),
	InputSchema: anthropic.ToolInputSchemaParam{
		Properties: map[string]any{
			"kind": map[string]any{
				"type": "string",
				"enum": []string{"volume_adjust", "load_adjust", "exercise_swap", "deload", "rest_day", "habit_nudge"},
			},
			"rationale": map[string]any{
				"type":        "string",
				"description": "Una frase en español explicando el motivo, para mostrar al coach",
			},
			"confidence": map[string]any{
				"type":        "number",
				"description": "Que tan seguro estas de esta sugerencia, entre 0 y 1",
			},
			"payload": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_exercise_id":      map[string]any{"type": "string"},
					"delta_percent":           map[string]any{"type": "number"},
					"suggested_exercise_name": map[string]any{"type": "string"},
					"note":                    map[string]any{"type": "string"},
				},
			},
		},
		Required: []string{"kind", "rationale", "confidence", "payload"},
	},
}

type Suggester struct {
	client *anthropic.Client
	model  string
}

func NewSuggester(client *anthropic.Client, model string) *Suggester {
	if model == "" {
		model = domain.DefaultSuggestionModel
	}
	return &Suggester{client: client, model: model}
}

var _ domain.SuggestionProposer = (*Suggester)(nil)

// Propose llama a Claude con el resumen de la asignacion y devuelve una
// sugerencia validada contra el esquema. Nunca guarda ni aplica nada — eso
// es responsabilidad del caller.
func (s *Suggester) Propose(ctx context.Context, summary domain.AssignmentSummary) (domain.Suggestion, error) {
	input, err := json.Marshal(summary)
	if err != nil {
		return domain.Suggestion{}, fmt.Errorf("marshal summary: %w", err)
	}

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(string(input))),
		},
		Tools:      []anthropic.ToolUnionParam{{OfTool: &proposeSuggestionTool}},
		ToolChoice: anthropic.ToolChoiceParamOfTool("propose_suggestion"),
	})
	if err != nil {
		return domain.Suggestion{}, fmt.Errorf("claude request: %w", err)
	}

	for _, block := range resp.Content {
		toolUse, ok := block.AsAny().(anthropic.ToolUseBlock)
		if !ok || toolUse.Name != "propose_suggestion" {
			continue
		}
		var sug domain.Suggestion
		if err := json.Unmarshal([]byte(toolUse.JSON.Input.Raw()), &sug); err != nil {
			return domain.Suggestion{}, fmt.Errorf("parse tool input: %w", err)
		}
		if err := sug.Validate(); err != nil {
			return domain.Suggestion{}, fmt.Errorf("respuesta invalida contra el esquema: %w", err)
		}
		return sug, nil
	}
	return domain.Suggestion{}, fmt.Errorf("claude no llamo a propose_suggestion (stop_reason=%s)", resp.StopReason)
}
