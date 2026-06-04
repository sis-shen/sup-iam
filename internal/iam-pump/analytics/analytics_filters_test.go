package analytics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticFilters_HasFilter(t *testing.T) {
	tests := []struct {
		name    string
		filters *AnalyticFilters
		want    bool
	}{
		{
			name:    "both nil",
			filters: &AnalyticFilters{},
			want:    false,
		},
		{
			name: "only Usernames set",
			filters: &AnalyticFilters{
				Usernames: []string{"alice"},
			},
			want: true,
		},
		{
			name: "only SkippedUsernames set",
			filters: &AnalyticFilters{
				SkippedUsernames: []string{"bob"},
			},
			want: true,
		},
		{
			name: "both set",
			filters: &AnalyticFilters{
				Usernames:        []string{"alice"},
				SkippedUsernames: []string{"bob"},
			},
			want: true,
		},
		{
			name: "both empty slices (not nil)",
			filters: &AnalyticFilters{
				Usernames:        []string{},
				SkippedUsernames: []string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filters.HasFilter()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnalyticFilters_ShouldFilter_NoFilters(t *testing.T) {
	f := &AnalyticFilters{}
	assert.False(t, f.ShouldFilter("anyone"), "no filters should never filter")
	assert.False(t, f.ShouldFilter(""), "no filters should never filter")
}

func TestAnalyticFilters_ShouldFilter_SkippedUsernames(t *testing.T) {
	f := &AnalyticFilters{
		SkippedUsernames: []string{"bob", "alice"},
	}

	assert.True(t, f.ShouldFilter("bob"), "bob should be filtered")
	assert.True(t, f.ShouldFilter("alice"), "alice should be filtered")
	assert.False(t, f.ShouldFilter("charlie"), "charlie should not be filtered")
	assert.False(t, f.ShouldFilter(""), "empty should not be filtered")
}

func TestAnalyticFilters_ShouldFilter_UsernamesWhitelist(t *testing.T) {
	f := &AnalyticFilters{
		Usernames: []string{"alice", "bob"},
	}

	assert.False(t, f.ShouldFilter("alice"), "alice should NOT be filtered (in whitelist)")
	assert.False(t, f.ShouldFilter("bob"), "bob should NOT be filtered (in whitelist)")
	assert.True(t, f.ShouldFilter("charlie"), "charlie should be filtered (not in whitelist)")
	assert.True(t, f.ShouldFilter(""), "empty should be filtered (not in whitelist)")
}

func TestAnalyticFilters_ShouldFilter_BlacklistOverridesWhitelist(t *testing.T) {
	// 黑名单优先级高于白名单
	f := &AnalyticFilters{
		Usernames:        []string{"alice", "bob", "mallory"},
		SkippedUsernames: []string{"mallory"},
	}

	assert.False(t, f.ShouldFilter("alice"), "alice: in whitelist, not blacklisted -> keep")
	assert.False(t, f.ShouldFilter("bob"), "bob: in whitelist, not blacklisted -> keep")
	assert.True(t, f.ShouldFilter("mallory"), "mallory: in whitelist AND blacklisted -> filtered (blacklist wins)")
	assert.True(t, f.ShouldFilter("charlie"), "charlie: not in whitelist -> filtered")
}

func TestAnalyticFilters_ShouldFilter_OnlyBlacklist_NoWhitelist(t *testing.T) {
	// Usernames == nil, SkippedUsernames != nil
	f := &AnalyticFilters{
		SkippedUsernames: []string{"bob"},
	}

	assert.True(t, f.ShouldFilter("bob"), "bob explicitly skipped")
	assert.False(t, f.ShouldFilter("alice"), "alice not skipped, no whitelist -> keep")
	assert.False(t, f.ShouldFilter(""), "empty not skipped -> keep")
}

func TestAnalyticFilters_ShouldFilter_OnlyWhitelist_NoBlacklist(t *testing.T) {
	// Usernames != nil, SkippedUsernames == nil
	f := &AnalyticFilters{
		Usernames: []string{"alice"},
	}

	assert.False(t, f.ShouldFilter("alice"), "alice in whitelist -> keep")
	assert.True(t, f.ShouldFilter("bob"), "bob not in whitelist -> filtered")
}

func TestAnalyticFilters_ShouldFilter_EmptyLists(t *testing.T) {
	f := &AnalyticFilters{
		Usernames:        []string{},
		SkippedUsernames: []string{},
	}

	assert.False(t, f.ShouldFilter("anyone"), "with empty lists, everyone is not filtered")
}

func TestAnalyticFilters_ShouldFilter_OnceInitialization(t *testing.T) {
	// Verify that adding to Usernames AFTER first ShouldFilter call doesn't change behavior
	f := &AnalyticFilters{
		Usernames: []string{"alice"},
	}

	// First call triggers once.Do
	f.ShouldFilter("alice")

	// Mutate the slice (this should NOT affect the cached set)
	f.Usernames = append(f.Usernames, "bob")

	assert.False(t, f.ShouldFilter("alice"), "alice should still pass")
	assert.True(t, f.ShouldFilter("bob"), "bob was added after init, should still be filtered (cached)")
}
