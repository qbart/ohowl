package cloudh

type Meta struct {
	Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page     int `json:"page,omitempty"`
	PerPage  int `json:"per_page,omitempty"`
	LastPage int `json:"last_page,omitempty"`
	Total    int `json:"total_entries,omitempty"`
}
