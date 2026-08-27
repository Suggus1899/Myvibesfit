package service

import (
	"testing"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

func f(v float64) *float64 { return &v }

func TestAggregateTrends(t *testing.T) {
	ex1 := uuid.New()
	ex2 := uuid.New()
	sets := []domain.RecentWorkingSet{
		{ExerciseID: ex1, ExerciseName: "Sentadilla", WeightKg: f(80), RPE: f(7)},
		{ExerciseID: ex1, ExerciseName: "Sentadilla", WeightKg: f(85), RPE: f(8)},
		{ExerciseID: ex2, ExerciseName: "Press banca", WeightKg: nil, RPE: nil},
	}

	names := map[uuid.UUID]string{}
	trends := aggregateTrends(sets, names)

	if len(trends) != 2 {
		t.Fatalf("esperaba 2 tendencias, dio %d", len(trends))
	}

	var squat *struct {
		count       int
		first, last *float64
		avgRPE      *float64
	}
	for _, tr := range trends {
		if tr.ExerciseName != "Sentadilla" {
			continue
		}
		squat = &struct {
			count       int
			first, last *float64
			avgRPE      *float64
		}{tr.SetCount, tr.FirstWeightKg, tr.LastWeightKg, tr.AvgRPE}
	}
	if squat == nil {
		t.Fatal("no se encontro la tendencia de Sentadilla")
	}
	if squat.count != 2 {
		t.Errorf("set_count esperado 2, dio %d", squat.count)
	}
	if squat.first == nil || *squat.first != 80 {
		t.Errorf("first_weight_kg esperado 80, dio %v", squat.first)
	}
	if squat.last == nil || *squat.last != 85 {
		t.Errorf("last_weight_kg esperado 85, dio %v", squat.last)
	}
	if squat.avgRPE == nil || *squat.avgRPE != 7.5 {
		t.Errorf("avg_rpe esperado 7.5, dio %v", squat.avgRPE)
	}
	if names[ex1] != "Sentadilla" || names[ex2] != "Press banca" {
		t.Errorf("mapa de nombres incompleto: %v", names)
	}
}
