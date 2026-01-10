package templates

import (
	"testing"

	"arabica/internal/models"
)

func TestFormatTemp(t *testing.T) {
	tests := []struct {
		name     string
		temp     float64
		expected string
	}{
		{"zero returns N/A", 0, "N/A"},
		{"celsius range", 93.5, "93.5°C"},
		{"celsius whole number", 90.0, "90.0°C"},
		{"celsius at 100", 100.0, "100.0°C"},
		{"fahrenheit range", 200.0, "200.0°F"},
		{"fahrenheit at 212", 212.0, "212.0°F"},
		{"low temp celsius", 20.5, "20.5°C"},
		{"just over 100 is fahrenheit", 100.1, "100.1°F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTemp(tt.temp)
			if got != tt.expected {
				t.Errorf("formatTemp(%v) = %q, want %q", tt.temp, got, tt.expected)
			}
		})
	}
}

func TestFormatTempValue(t *testing.T) {
	tests := []struct {
		name     string
		temp     float64
		expected string
	}{
		{"zero", 0, "0.0"},
		{"whole number", 93.0, "93.0"},
		{"decimal", 93.5, "93.5"},
		{"high precision rounds", 93.55, "93.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTempValue(tt.temp)
			if got != tt.expected {
				t.Errorf("formatTempValue(%v) = %q, want %q", tt.temp, got, tt.expected)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{"zero returns N/A", 0, "N/A"},
		{"seconds only", 30, "30s"},
		{"exactly one minute", 60, "1m"},
		{"minutes and seconds", 90, "1m 30s"},
		{"multiple minutes", 180, "3m"},
		{"multiple minutes and seconds", 185, "3m 5s"},
		{"large time", 3661, "61m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.seconds)
			if got != tt.expected {
				t.Errorf("formatTime(%v) = %q, want %q", tt.seconds, got, tt.expected)
			}
		})
	}
}

func TestFormatRating(t *testing.T) {
	tests := []struct {
		name     string
		rating   int
		expected string
	}{
		{"zero returns N/A", 0, "N/A"},
		{"rating 1", 1, "1/10"},
		{"rating 5", 5, "5/10"},
		{"rating 10", 10, "10/10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRating(tt.rating)
			if got != tt.expected {
				t.Errorf("formatRating(%v) = %q, want %q", tt.rating, got, tt.expected)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 123, "123"},
		{"large number", 99999, "99999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatID(tt.id)
			if got != tt.expected {
				t.Errorf("formatID(%v) = %q, want %q", tt.id, got, tt.expected)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		name     string
		val      int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 42, "42"},
		{"negative", -5, "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatInt(tt.val)
			if got != tt.expected {
				t.Errorf("formatInt(%v) = %q, want %q", tt.val, got, tt.expected)
			}
		})
	}
}

func TestFormatRoasterID(t *testing.T) {
	t.Run("nil returns null", func(t *testing.T) {
		got := formatRoasterID(nil)
		if got != "null" {
			t.Errorf("formatRoasterID(nil) = %q, want %q", got, "null")
		}
	})

	t.Run("valid pointer", func(t *testing.T) {
		id := 123
		got := formatRoasterID(&id)
		if got != "123" {
			t.Errorf("formatRoasterID(&123) = %q, want %q", got, "123")
		}
	})

	t.Run("zero pointer", func(t *testing.T) {
		id := 0
		got := formatRoasterID(&id)
		if got != "0" {
			t.Errorf("formatRoasterID(&0) = %q, want %q", got, "0")
		}
	})
}

func TestPoursToJSON(t *testing.T) {
	tests := []struct {
		name     string
		pours    []*models.Pour
		expected string
	}{
		{
			name:     "empty pours",
			pours:    []*models.Pour{},
			expected: "[]",
		},
		{
			name:     "nil pours",
			pours:    nil,
			expected: "[]",
		},
		{
			name: "single pour",
			pours: []*models.Pour{
				{WaterAmount: 50, TimeSeconds: 30},
			},
			expected: `[{"water":50,"time":30}]`,
		},
		{
			name: "multiple pours",
			pours: []*models.Pour{
				{WaterAmount: 50, TimeSeconds: 30},
				{WaterAmount: 100, TimeSeconds: 60},
				{WaterAmount: 150, TimeSeconds: 90},
			},
			expected: `[{"water":50,"time":30},{"water":100,"time":60},{"water":150,"time":90}]`,
		},
		{
			name: "zero values",
			pours: []*models.Pour{
				{WaterAmount: 0, TimeSeconds: 0},
			},
			expected: `[{"water":0,"time":0}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := poursToJSON(tt.pours)
			if got != tt.expected {
				t.Errorf("poursToJSON() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		p := ptr(42)
		if *p != 42 {
			t.Errorf("ptr(42) = %v, want 42", *p)
		}
	})

	t.Run("string", func(t *testing.T) {
		p := ptr("hello")
		if *p != "hello" {
			t.Errorf("ptr(\"hello\") = %v, want \"hello\"", *p)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		p := ptr(0)
		if *p != 0 {
			t.Errorf("ptr(0) = %v, want 0", *p)
		}
	})
}

func TestPtrEquals(t *testing.T) {
	t.Run("nil pointer returns false", func(t *testing.T) {
		var p *int = nil
		if ptrEquals(p, 42) {
			t.Error("ptrEquals(nil, 42) should be false")
		}
	})

	t.Run("matching value returns true", func(t *testing.T) {
		val := 42
		if !ptrEquals(&val, 42) {
			t.Error("ptrEquals(&42, 42) should be true")
		}
	})

	t.Run("non-matching value returns false", func(t *testing.T) {
		val := 42
		if ptrEquals(&val, 99) {
			t.Error("ptrEquals(&42, 99) should be false")
		}
	})

	t.Run("string comparison", func(t *testing.T) {
		s := "hello"
		if !ptrEquals(&s, "hello") {
			t.Error("ptrEquals(&\"hello\", \"hello\") should be true")
		}
		if ptrEquals(&s, "world") {
			t.Error("ptrEquals(&\"hello\", \"world\") should be false")
		}
	})
}

func TestPtrValue(t *testing.T) {
	t.Run("nil int returns zero", func(t *testing.T) {
		var p *int = nil
		if ptrValue(p) != 0 {
			t.Errorf("ptrValue(nil) = %v, want 0", ptrValue(p))
		}
	})

	t.Run("valid int returns value", func(t *testing.T) {
		val := 42
		if ptrValue(&val) != 42 {
			t.Errorf("ptrValue(&42) = %v, want 42", ptrValue(&val))
		}
	})

	t.Run("nil string returns empty", func(t *testing.T) {
		var p *string = nil
		if ptrValue(p) != "" {
			t.Errorf("ptrValue(nil string) = %v, want \"\"", ptrValue(p))
		}
	})

	t.Run("valid string returns value", func(t *testing.T) {
		s := "hello"
		if ptrValue(&s) != "hello" {
			t.Errorf("ptrValue(&\"hello\") = %v, want \"hello\"", ptrValue(&s))
		}
	})
}
