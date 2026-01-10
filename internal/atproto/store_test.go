package atproto

import (
	"testing"

	"arabica/internal/models"
)

func TestLinkBeansToRoasters(t *testing.T) {
	t.Run("links beans to matching roasters", func(t *testing.T) {
		roasters := []*models.Roaster{
			{RKey: "roaster1", Name: "Roaster One"},
			{RKey: "roaster2", Name: "Roaster Two"},
			{RKey: "roaster3", Name: "Roaster Three"},
		}

		beans := []*models.Bean{
			{RKey: "bean1", Name: "Bean One", RoasterRKey: "roaster1"},
			{RKey: "bean2", Name: "Bean Two", RoasterRKey: "roaster2"},
			{RKey: "bean3", Name: "Bean Three", RoasterRKey: ""}, // No roaster
		}

		LinkBeansToRoasters(beans, roasters)

		// Bean 1 should be linked to Roaster One
		if beans[0].Roaster == nil {
			t.Error("Bean 1 should have roaster linked")
		} else if beans[0].Roaster.Name != "Roaster One" {
			t.Errorf("Bean 1 roaster = %q, want %q", beans[0].Roaster.Name, "Roaster One")
		}

		// Bean 2 should be linked to Roaster Two
		if beans[1].Roaster == nil {
			t.Error("Bean 2 should have roaster linked")
		} else if beans[1].Roaster.Name != "Roaster Two" {
			t.Errorf("Bean 2 roaster = %q, want %q", beans[1].Roaster.Name, "Roaster Two")
		}

		// Bean 3 should have no roaster
		if beans[2].Roaster != nil {
			t.Error("Bean 3 should have no roaster linked")
		}
	})

	t.Run("handles missing roaster gracefully", func(t *testing.T) {
		roasters := []*models.Roaster{
			{RKey: "roaster1", Name: "Roaster One"},
		}

		beans := []*models.Bean{
			{RKey: "bean1", Name: "Bean One", RoasterRKey: "nonexistent"},
		}

		// Should not panic
		LinkBeansToRoasters(beans, roasters)

		// Bean should have nil roaster since reference doesn't match
		if beans[0].Roaster != nil {
			t.Error("Bean with nonexistent roaster ref should have nil Roaster")
		}
	})

	t.Run("handles empty slices", func(t *testing.T) {
		// Should not panic with empty inputs
		LinkBeansToRoasters(nil, nil)
		LinkBeansToRoasters([]*models.Bean{}, []*models.Roaster{})
	})
}
