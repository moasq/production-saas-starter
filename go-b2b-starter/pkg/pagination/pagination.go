package listingshared

import "fmt"

type PagePagination[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

type Meta struct {
	TotalItems         int    `json:"total_items"`
	Page               int    `json:"page"`
	PageSize           int    `json:"page_size"`
	ReturnedItemsCount int    `json:"returned_items_count"`
	HasMore            bool   `json:"has_more"`
	FirstPageURL       string `json:"first_page_url"`
	PreviousPageURL    string `json:"previous_page_url"`
	NextPageURL        string `json:"next_page_url"`
	LastPageURL        string `json:"last_page_url"`
	TotalPages         int    `json:"total_pages"`
}

func NewPagePagination[T any](totalItems, page, pageSize int, items []T) *PagePagination[T] {
	// Calculate total pages properly
	totalPages := (totalItems + pageSize - 1) / pageSize

	p := &PagePagination[T]{
		Meta: Meta{
			TotalItems:         totalItems,
			Page:               page,
			PageSize:           pageSize,
			ReturnedItemsCount: len(items),
			HasMore:            page < totalPages, // Set HasMore
			TotalPages:         totalPages,
		},
		Items: items,
	}

	// First page URL only if we're not on first page
	if page > 1 {
		p.Meta.FirstPageURL = fmt.Sprintf("?page=%d&pageSize=%d", 1, pageSize)
	}

	// Last page URL only if there are multiple pages and we're not on last page
	if page < totalPages {
		p.Meta.LastPageURL = fmt.Sprintf("?page=%d&pageSize=%d", totalPages, pageSize)
	}

	// Previous page URL only if we're not on first page
	if page > 1 {
		p.Meta.PreviousPageURL = fmt.Sprintf("?page=%d&pageSize=%d", page-1, pageSize)
	}

	// Next page URL only if there are more pages
	if page < totalPages {
		p.Meta.NextPageURL = fmt.Sprintf("?page=%d&pageSize=%d", page+1, pageSize)
	}

	return p
}

func PageToOffset(page int, limit int) (int, error) {
	if page < 1 {
		return 0, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 {
		return 0, fmt.Errorf("limit must be greater than 0")
	}

	return (page - 1) * limit, nil
}
