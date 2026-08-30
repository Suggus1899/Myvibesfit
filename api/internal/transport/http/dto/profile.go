package dto

type ClientProfileDTO struct {
	BirthDate          string   `json:"birth_date,omitempty"`
	Sex                string   `json:"sex"`
	HeightCm           *float64 `json:"height_cm,omitempty"`
	Experience         string   `json:"experience"`
	PrimaryGoal        string   `json:"primary_goal"`
	DaysPerWeek        int      `json:"days_per_week"`
	SessionMinutes     int      `json:"session_minutes"`
	AvailableEquipment []string `json:"available_equipment"`
	Limitations        string   `json:"limitations,omitempty"`
	UnitSystem         string   `json:"unit_system"`
	IsOnboarded        bool     `json:"is_onboarded"`
}

type SaveProfileRequest struct {
	BirthDate          string   `json:"birth_date"` // YYYY-MM-DD
	Sex                string   `json:"sex"`
	HeightCm           *float64 `json:"height_cm"`
	Experience         string   `json:"experience"`
	PrimaryGoal        string   `json:"primary_goal"`
	DaysPerWeek        int      `json:"days_per_week"`
	SessionMinutes     int      `json:"session_minutes"`
	AvailableEquipment []string `json:"available_equipment"`
	Limitations        string   `json:"limitations"`
	UnitSystem         string   `json:"unit_system"`
}
