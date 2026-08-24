package progression

import "testing"

func TestDoubleProgression(t *testing.T) {
	base := Input{LastWeightKg: 60, TargetRepsMin: 8, TargetRepsMax: 12}

	t.Run("todas las series tocan el techo: sube peso", func(t *testing.T) {
		in := base
		in.RepsAchieved = []int{12, 12, 12}
		out, err := Apply(TypeDoubleProgression, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 62.5 {
			t.Errorf("NextWeightKg = %v, want 62.5", out.NextWeightKg)
		}
	})

	t.Run("una serie no llega al techo: mantiene peso", func(t *testing.T) {
		in := base
		in.RepsAchieved = []int{12, 10, 12}
		out, err := Apply(TypeDoubleProgression, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 60 {
			t.Errorf("NextWeightKg = %v, want 60 (sin cambio)", out.NextWeightKg)
		}
	})

	t.Run("respeta increment_kg custom", func(t *testing.T) {
		in := base
		in.RepsAchieved = []int{12, 12}
		out, err := Apply(TypeDoubleProgression, []byte(`{"increment_kg":5,"round_to_kg":5}`), in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 65 {
			t.Errorf("NextWeightKg = %v, want 65", out.NextWeightKg)
		}
	})
}

func TestLinearLoad(t *testing.T) {
	base := Input{LastWeightKg: 100}

	t.Run("exito: sube peso", func(t *testing.T) {
		in := base
		in.Success = true
		out, err := Apply(TypeLinearLoad, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 102.5 {
			t.Errorf("NextWeightKg = %v, want 102.5", out.NextWeightKg)
		}
	})

	t.Run("fallo: mantiene peso", func(t *testing.T) {
		in := base
		in.Success = false
		out, err := Apply(TypeLinearLoad, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 100 {
			t.Errorf("NextWeightKg = %v, want 100 (sin cambio)", out.NextWeightKg)
		}
	})
}

func TestRPEAutoregulated(t *testing.T) {
	t.Run("costo mas de lo esperado: baja peso", func(t *testing.T) {
		in := Input{LastWeightKg: 100, TargetRPE: 8, ActualRPE: 9.5}
		out, err := Apply(TypeRPEAutoregulated, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg >= 100 {
			t.Errorf("NextWeightKg = %v, want < 100 (RPE mas alto de lo esperado)", out.NextWeightKg)
		}
	})

	t.Run("costo menos de lo esperado: sube peso", func(t *testing.T) {
		in := Input{LastWeightKg: 100, TargetRPE: 8, ActualRPE: 6}
		out, err := Apply(TypeRPEAutoregulated, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg <= 100 {
			t.Errorf("NextWeightKg = %v, want > 100 (RPE mas bajo de lo esperado)", out.NextWeightKg)
		}
	})

	t.Run("sin RPE reportado: mantiene peso", func(t *testing.T) {
		in := Input{LastWeightKg: 100, TargetRPE: 8}
		out, err := Apply(TypeRPEAutoregulated, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 100 {
			t.Errorf("NextWeightKg = %v, want 100 (sin cambio)", out.NextWeightKg)
		}
	})
}

func TestPercentage1RM(t *testing.T) {
	t.Run("calcula porcentaje redondeado", func(t *testing.T) {
		in := Input{OneRepMaxKg: 140}
		out, err := Apply(TypePercentage1RM, []byte(`{"percent":0.8,"round_to_kg":2.5}`), in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 112.5 {
			t.Errorf("NextWeightKg = %v, want 112.5 (140*0.8=112, redondeado a 112.5)", out.NextWeightKg)
		}
	})

	t.Run("sin 1RM: mantiene peso previo", func(t *testing.T) {
		in := Input{LastWeightKg: 80}
		out, err := Apply(TypePercentage1RM, nil, in)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if out.NextWeightKg != 80 {
			t.Errorf("NextWeightKg = %v, want 80 (sin cambio)", out.NextWeightKg)
		}
	})
}

func TestApplyUnknownType(t *testing.T) {
	_, err := Apply("no-existe", nil, Input{})
	if err != ErrUnknownType {
		t.Errorf("Apply() error = %v, want ErrUnknownType", err)
	}
}
