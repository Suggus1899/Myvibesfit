// Package progression calcula el siguiente objetivo de carga/repeticiones
// para un ejercicio segun una estrategia de progresion. Son funciones puras:
// no tocan la base de datos, no conocen sesiones ni asignaciones. El
// llamador arma el Input a partir del historial reciente (fase 6+) y decide
// que hacer con el Output (queda en manos del coach o de una sugerencia de
// IA supervisada, fase 9).
package progression

import (
	"encoding/json"
	"errors"
	"math"
)

type Type string

const (
	TypeDoubleProgression Type = "double_progression"
	TypeLinearLoad        Type = "linear_load"
	TypeRPEAutoregulated  Type = "rpe_autoregulated"
	TypePercentage1RM     Type = "percentage_1rm"
	TypeNone              Type = "none"
)

var ErrUnknownType = errors.New("progression: unknown rule type")

// Input reune todo lo que cualquier estrategia podria necesitar. Cada
// estrategia solo lee los campos que le importan.
type Input struct {
	LastWeightKg  float64
	RepsAchieved  []int // reps logradas en cada serie de trabajo, en orden
	TargetRepsMin int
	TargetRepsMax int
	Success       bool // ¿se cumplio el objetivo prescrito esta sesion?
	TargetRPE     float64
	ActualRPE     float64
	OneRepMaxKg   float64
}

type Output struct {
	NextWeightKg float64
	NextRepsMin  int
	NextRepsMax  int
	Note         string
}

// Apply decodifica params (el jsonb de progression_rule) y aplica la
// estrategia correspondiente. params puede ser nil/vacio: cada estrategia
// tiene defaults razonables.
func Apply(ruleType Type, params json.RawMessage, in Input) (Output, error) {
	switch ruleType {
	case TypeDoubleProgression:
		var p doubleProgressionParams
		if err := decodeParams(params, &p); err != nil {
			return Output{}, err
		}
		return doubleProgression(in, p), nil
	case TypeLinearLoad:
		var p linearLoadParams
		if err := decodeParams(params, &p); err != nil {
			return Output{}, err
		}
		return linearLoad(in, p), nil
	case TypeRPEAutoregulated:
		var p rpeParams
		if err := decodeParams(params, &p); err != nil {
			return Output{}, err
		}
		return rpeAutoregulated(in, p), nil
	case TypePercentage1RM:
		var p percentageParams
		if err := decodeParams(params, &p); err != nil {
			return Output{}, err
		}
		return percentage1RM(in, p), nil
	case TypeNone:
		return Output{NextWeightKg: in.LastWeightKg, NextRepsMin: in.TargetRepsMin, NextRepsMax: in.TargetRepsMax, Note: "sin progresion automatica"}, nil
	default:
		return Output{}, ErrUnknownType
	}
}

func decodeParams(raw json.RawMessage, dst any) error {
	if len(raw) == 0 {
		return nil // dst se queda en sus valores cero, cada funcion aplica su propio default
	}
	return json.Unmarshal(raw, dst)
}

// --- doble progresion ---
// Sube reps dentro del rango objetivo; cuando todas las series tocan el
// techo del rango, sube peso y reinicia el rango al piso.

type doubleProgressionParams struct {
	IncrementKg float64 `json:"increment_kg"`
	RoundToKg   float64 `json:"round_to_kg"`
}

func doubleProgression(in Input, p doubleProgressionParams) Output {
	increment := orDefault(p.IncrementKg, 2.5)
	roundTo := orDefault(p.RoundToKg, 2.5)

	if len(in.RepsAchieved) == 0 || in.TargetRepsMax == 0 {
		return Output{NextWeightKg: in.LastWeightKg, NextRepsMin: in.TargetRepsMin, NextRepsMax: in.TargetRepsMax, Note: "sin datos suficientes, se mantiene"}
	}

	allHitTop := true
	for _, reps := range in.RepsAchieved {
		if reps < in.TargetRepsMax {
			allHitTop = false
			break
		}
	}

	if allHitTop {
		return Output{
			NextWeightKg: round(in.LastWeightKg+increment, roundTo),
			NextRepsMin:  in.TargetRepsMin,
			NextRepsMax:  in.TargetRepsMax,
			Note:         "techo del rango alcanzado en todas las series: sube peso, reinicia reps",
		}
	}

	return Output{
		NextWeightKg: in.LastWeightKg,
		NextRepsMin:  in.TargetRepsMin,
		NextRepsMax:  in.TargetRepsMax,
		Note:         "sigue progresando reps dentro del rango",
	}
}

// --- carga lineal ---
// Sube un incremento fijo cada sesion exitosa; si falla, mantiene el peso.

type linearLoadParams struct {
	IncrementKg float64 `json:"increment_kg"`
	RoundToKg   float64 `json:"round_to_kg"`
}

func linearLoad(in Input, p linearLoadParams) Output {
	increment := orDefault(p.IncrementKg, 2.5)
	roundTo := orDefault(p.RoundToKg, 2.5)

	if !in.Success {
		return Output{NextWeightKg: in.LastWeightKg, NextRepsMin: in.TargetRepsMin, NextRepsMax: in.TargetRepsMax, Note: "objetivo no cumplido: mantiene peso"}
	}

	return Output{
		NextWeightKg: round(in.LastWeightKg+increment, roundTo),
		NextRepsMin:  in.TargetRepsMin,
		NextRepsMax:  in.TargetRepsMax,
		Note:         "objetivo cumplido: sube peso",
	}
}

// --- autorregulado por RPE ---
// Ajusta el peso proporcional a la diferencia entre el RPE objetivo y el
// reportado: si costo mas de lo esperado, baja peso; si costo menos, sube.

type rpeParams struct {
	AdjustmentPctPerRPE float64 `json:"adjustment_pct_per_rpe"`
	RoundToKg           float64 `json:"round_to_kg"`
}

func rpeAutoregulated(in Input, p rpeParams) Output {
	adjPerPoint := orDefault(p.AdjustmentPctPerRPE, 0.025)
	roundTo := orDefault(p.RoundToKg, 2.5)

	if in.TargetRPE == 0 || in.ActualRPE == 0 {
		return Output{NextWeightKg: in.LastWeightKg, NextRepsMin: in.TargetRepsMin, NextRepsMax: in.TargetRepsMax, Note: "sin RPE reportado, se mantiene"}
	}

	delta := in.TargetRPE - in.ActualRPE // positivo = sesion mas facil de lo esperado
	factor := 1 + delta*adjPerPoint
	return Output{
		NextWeightKg: round(in.LastWeightKg*factor, roundTo),
		NextRepsMin:  in.TargetRepsMin,
		NextRepsMax:  in.TargetRepsMax,
		Note:         "ajustado segun RPE reportado vs objetivo",
	}
}

// --- porcentaje de 1RM ---
// El peso objetivo siempre es un porcentaje fijo del ultimo 1RM conocido.

type percentageParams struct {
	Percent   float64 `json:"percent"`
	RoundToKg float64 `json:"round_to_kg"`
}

func percentage1RM(in Input, p percentageParams) Output {
	roundTo := orDefault(p.RoundToKg, 2.5)
	percent := p.Percent
	if percent == 0 {
		percent = 0.75
	}

	if in.OneRepMaxKg == 0 {
		return Output{NextWeightKg: in.LastWeightKg, NextRepsMin: in.TargetRepsMin, NextRepsMax: in.TargetRepsMax, Note: "sin 1RM registrado, se mantiene"}
	}

	return Output{
		NextWeightKg: round(in.OneRepMaxKg*percent, roundTo),
		NextRepsMin:  in.TargetRepsMin,
		NextRepsMax:  in.TargetRepsMax,
		Note:         "calculado como porcentaje del 1RM",
	}
}

func orDefault(v, fallback float64) float64 {
	if v == 0 {
		return fallback
	}
	return v
}

func round(value, step float64) float64 {
	if step <= 0 {
		return value
	}
	return math.Round(value/step) * step
}
