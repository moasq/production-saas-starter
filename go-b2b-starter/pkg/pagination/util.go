package listingshared

// PaginationCalc converts offset and limit to page number and page size.
// Assumes default values if not specified.
func PaginationCalc(offset, limit int) (page, pageSize int) {
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}

	page = (offset / limit) + 1
	pageSize = limit

	return page, pageSize
}
