package domain

import "testing"

func TestSuggestionValidate(t *testing.T) {
	valid := func() Suggestion {
		return Suggestion{Kind: "load_adjust", Rationale: "subio RPE", Confidence: 0.8}
	}

	tests := []struct {
		name    string
		mutate  func(*Suggestion)
		wantErr bool
	}{
		{"sugerencia valida", func(*Suggestion) {}, false},
		{"kind invalido", func(s *Suggestion) { s.Kind = "inventado" }, true},
		{"rationale vacio", func(s *Suggestion) { s.Rationale = "" }, true},
		{"confidence negativa", func(s *Suggestion) { s.Confidence = -0.1 }, true},
		{"confidence mayor a 1", func(s *Suggestion) { s.Confidence = 1.5 }, true},

		// Guardrails de magnitud (docs/ARCHITECTURE.md §4).
		{"carga en el tope pasa", func(s *Suggestion) { s.Payload.DeltaPercent = 10 }, false},
		{"carga sobre el tope se descarta", func(s *Suggestion) { s.Payload.DeltaPercent = 12 }, true},
		{"bajada de carga grande tambien se descarta", func(s *Suggestion) { s.Payload.DeltaPercent = -25 }, true},
		{"volumen usa su propio tope", func(s *Suggestion) {
			s.Kind = "volume_adjust"
			s.Payload.DeltaPercent = 25
		}, false},
		{"volumen sobre el tope se descarta", func(s *Suggestion) {
			s.Kind = "volume_adjust"
			s.Payload.DeltaPercent = 40
		}, true},
		{"deload no mira delta_percent", func(s *Suggestion) {
			s.Kind = "deload"
			s.Payload.DeltaPercent = 90
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := valid()
			tt.mutate(&s)
			err := s.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("esperaba error, no hubo (%+v)", s)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("no esperaba error, hubo: %v", err)
			}
		})
	}
}
