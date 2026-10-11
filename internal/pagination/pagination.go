package pagination

const (
	defaultLimit = 20
	maxLimit     = 100
)

func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}

type OffsetPagination struct {
	Page  int
	Limit int
}

func (p OffsetPagination) Normalize() OffsetPagination {
	if p.Page <= 0 {
		p.Page = 1
	}
	p.Limit = clampLimit(p.Limit)
	return p
}

func (p OffsetPagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

type OffsetPage[T any] struct {
	Items      []T
	Page       int
	Limit      int
	TotalItems int
	TotalPages int
}

func NewOffsetPage[T any](items []T, p OffsetPagination, totalItems int) OffsetPage[T] {
	p = p.Normalize()
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + p.Limit - 1) / p.Limit
	}

	return OffsetPage[T]{
		Items:      items,
		Page:       p.Page,
		Limit:      p.Limit,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func (p OffsetPage[T]) HasNext() bool     { return p.Page < p.TotalPages }
func (p OffsetPage[T]) HasPrevious() bool { return p.Page > 1 }

// --- cursor pagination ---

type CursorPagination struct {
	Cursor string
	Limit  int
}

func (p CursorPagination) Normalize() CursorPagination {
	p.Limit = clampLimit(p.Limit)
	return p
}

type CursorPage[T any] struct {
	Items      []T
	NextCursor string
	HasMore    bool
}

func NewCursorPage[T any](items []T, limit int, cursorOf func(T) string) CursorPage[T] {
	limit = clampLimit(limit)
	hasMore := len(items) > limit
	next := ""
	if hasMore {
		items = items[:limit]
		next = cursorOf(items[len(items)-1])
	}

	return CursorPage[T]{
		Items:      items,
		NextCursor: next,
		HasMore:    hasMore,
	}
}
