package dto

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

func (q PageQuery) Normalize() PageQuery {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}
	return q
}

func (q PageQuery) Offset() int {
	normalized := q.Normalize()
	return (normalized.Page - 1) * normalized.PageSize
}
