package listingshared

type SearchableParams struct {
	Q     string `form:"q" binding:"required,min=1,max=100"`
	Page  int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100" default:"10"`
	Lang  string `form:"lang" binding:"omitempty,oneof=en ar" default:"en"`
}

func (s *SearchableParams) Validate() error {

	// Then, set default values for empty fields
	if s.Page == 0 {
		s.Page = 1
	}
	if s.Limit == 0 {
		s.Limit = 10
	}
	if s.Lang == "" {
		s.Lang = "en"
	}

	return nil
}

type ListableParams struct {
	Page  int    `form:"page" binding:"omitempty,min=1" default:"1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100" default:"10"`
	Lang  string `form:"lang" binding:"omitempty,oneof=en ar" default:"en"`
}

func (s *ListableParams) Validate() error {

	// Then, set default values for empty fields
	if s.Page == 0 {
		s.Page = 1
	}
	if s.Limit == 0 {
		s.Limit = 10
	}
	if s.Lang == "" {
		s.Lang = "en"
	}

	return nil
}
