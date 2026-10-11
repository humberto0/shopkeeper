package pagination

import (
	"strconv"
	"testing"
)

func TestOffsetPagination_Normalize(t *testing.T) {
	tests := []struct {
		name      string
		in        OffsetPagination
		wantPage  int
		wantLimit int
	}{
		{"zero values use defaults", OffsetPagination{}, 1, defaultLimit},
		{"negative page becomes first page", OffsetPagination{Page: -3, Limit: 10}, 1, 10},
		{"negative limit uses default", OffsetPagination{Page: 2, Limit: -1}, 2, defaultLimit},
		{"limit above max is capped", OffsetPagination{Page: 1, Limit: maxLimit + 1}, 1, maxLimit},
		{"valid values are kept", OffsetPagination{Page: 3, Limit: 15}, 3, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in.Normalize()
			if got.Page != tt.wantPage || got.Limit != tt.wantLimit {
				t.Errorf("Normalize() = {Page:%d Limit:%d}, want {Page:%d Limit:%d}",
					got.Page, got.Limit, tt.wantPage, tt.wantLimit)
			}
		})
	}
}

func TestOffsetPagination_Offset(t *testing.T) {
	p := OffsetPagination{Page: 3, Limit: 20}
	if got := p.Offset(); got != 40 {
		t.Errorf("Offset() = %d, want 40", got)
	}
}

func TestNewOffsetPage(t *testing.T) {
	tests := []struct {
		name           string
		page, limit    int
		totalItems     int
		wantTotalPages int
		wantNext       bool
		wantPrevious   bool
	}{
		{"empty result", 1, 20, 0, 0, false, false},
		{"exact multiple", 1, 10, 30, 3, true, false},
		{"partial last page", 3, 10, 25, 3, false, true},
		{"middle page", 2, 10, 25, 3, true, true},
		{"page past the end", 5, 10, 25, 3, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOffsetPage([]int{}, OffsetPagination{Page: tt.page, Limit: tt.limit}, tt.totalItems)
			if got.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", got.TotalPages, tt.wantTotalPages)
			}
			if got.HasNext() != tt.wantNext {
				t.Errorf("HasNext() = %v, want %v", got.HasNext(), tt.wantNext)
			}
			if got.HasPrevious() != tt.wantPrevious {
				t.Errorf("HasPrevious() = %v, want %v", got.HasPrevious(), tt.wantPrevious)
			}
		})
	}
}

func TestNewOffsetPage_NormalizesPagination(t *testing.T) {
	got := NewOffsetPage([]int{}, OffsetPagination{}, 5)
	if got.Page != 1 || got.Limit != defaultLimit {
		t.Errorf("got Page=%d Limit=%d, want Page=1 Limit=%d", got.Page, got.Limit, defaultLimit)
	}
}

func TestNewCursorPage(t *testing.T) {
	cursorOf := func(i int) string { return strconv.Itoa(i) }

	t.Run("more items than limit", func(t *testing.T) {
		// Repositories fetch limit+1 rows; the extra row only signals HasMore.
		got := NewCursorPage([]int{1, 2, 3}, 2, cursorOf)
		if len(got.Items) != 2 {
			t.Fatalf("len(Items) = %d, want 2", len(got.Items))
		}
		if !got.HasMore {
			t.Error("HasMore = false, want true")
		}
		if got.NextCursor != "2" {
			t.Errorf("NextCursor = %q, want %q", got.NextCursor, "2")
		}
	})

	t.Run("last page", func(t *testing.T) {
		got := NewCursorPage([]int{1, 2}, 2, cursorOf)
		if got.HasMore {
			t.Error("HasMore = true, want false")
		}
		if got.NextCursor != "" {
			t.Errorf("NextCursor = %q, want empty", got.NextCursor)
		}
	})
}

func TestCursorPagination_Normalize(t *testing.T) {
	got := CursorPagination{Cursor: "abc", Limit: 0}.Normalize()
	if got.Limit != defaultLimit || got.Cursor != "abc" {
		t.Errorf("Normalize() = %+v, want Cursor=abc Limit=%d", got, defaultLimit)
	}
}
