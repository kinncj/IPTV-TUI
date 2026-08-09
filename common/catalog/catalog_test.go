package catalog

import "testing"

func TestBuildGroupsAndSorts(t *testing.T) {
	chs := []Channel{
		{Name: "Zeta", Group: "Brazil"},
		{Name: "Alpha", Group: "Brazil"},
		{Name: "Gamma", Group: "Argentina"},
		{Name: "Orphan", Group: ""},
	}

	cat := Build(chs)

	if cat.Total != 4 {
		t.Fatalf("Total = %d, want 4", cat.Total)
	}
	if len(cat.Groups) != 3 {
		t.Fatalf("groups = %d, want 3 (Argentina, Brazil, Uncategorized)", len(cat.Groups))
	}

	// Groups sorted alphabetically.
	if cat.Groups[0].Name != "Argentina" || cat.Groups[1].Name != "Brazil" {
		t.Errorf("groups not sorted: %s, %s", cat.Groups[0].Name, cat.Groups[1].Name)
	}

	// Empty group bucketed under Uncategorized.
	if cat.Groups[2].Name != "Uncategorized" {
		t.Errorf("orphan group = %q, want Uncategorized", cat.Groups[2].Name)
	}

	// Channels within a group sorted alphabetically.
	brazil := cat.Groups[1]
	if brazil.Channels[0].Name != "Alpha" || brazil.Channels[1].Name != "Zeta" {
		t.Errorf("channels not sorted: %+v", brazil.Channels)
	}
}

func TestBuildByCategoryAndFlatten(t *testing.T) {
	chs := []Channel{
		{Name: "A", Group: "Brazil", Category: "news"},
		{Name: "B", Group: "Japan", Category: "news"},
		{Name: "C", Group: "Brazil", Category: "movies"},
		{Name: "D", Group: "Japan", Category: ""}, // no category
	}
	byCat := BuildByCategory(chs)
	got := map[string]int{}
	for _, g := range byCat.Groups {
		got[g.Name] = len(g.Channels)
	}
	if got["news"] != 2 || got["movies"] != 1 || got["Uncategorized"] != 1 {
		t.Errorf("category grouping wrong: %+v", got)
	}

	// Flatten of a country catalog round-trips the channel count.
	byCountry := Build(chs)
	if len(byCountry.Flatten()) != 4 {
		t.Errorf("flatten lost channels: %d", len(byCountry.Flatten()))
	}
	// Re-grouping the flattened set by category matches the direct build.
	if len(BuildByCategory(byCountry.Flatten()).Groups) != len(byCat.Groups) {
		t.Errorf("flatten+regroup mismatch")
	}
}
