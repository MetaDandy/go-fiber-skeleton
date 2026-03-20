package helper

import (
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type FindAllOptions struct {
	OrderBy     string `query:"order_by,default:created_at"`
	Sort        string `query:"sort,default:desc"`
	Search      string `query:"search"`
	Limit       uint   `query:"limit,default:10"`
	Offset      uint   `query:"offset,default:0"`
	ShowDeleted bool   `query:"show_deleted,default:false"`
	OnlyDeleted bool   `query:"only_deleted,default:false"`
}

func NewFindAllOptionsFromQuery(c fiber.Ctx) *FindAllOptions {
	q := new(FindAllOptions)

	// c.Bind().Query() automatically:
	// - Parses query parameters
	// - Converts types (string -> int, bool, etc)
	// - Applies default values from struct tags
	if err := c.Bind().Query(q); err != nil {
		// On error, initialize with hardcoded defaults
		q = &FindAllOptions{
			OrderBy: "created_at",
			Sort:    "desc",
			Limit:   10,
			Offset:  0,
		}
	}

	return &FindAllOptions{
		OrderBy:     q.OrderBy,
		Sort:        q.Sort,
		Search:      q.Search,
		Limit:       uint(q.Limit),
		Offset:      uint(q.Offset),
		ShowDeleted: q.ShowDeleted,
		OnlyDeleted: q.OnlyDeleted,
	}
}

func (f *FindAllOptions) ApplyFindAllOptions(query *gorm.DB) (*gorm.DB, int64) {
	var total int64

	if f == nil {
		query = query.Order("created_at asc")
		query.Count(&total)
		return query, total
	}

	orderBy := f.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}

	sort := "asc"
	if f.Sort == "desc" {
		sort = "desc"
	}

	query = query.Order(orderBy + " " + sort)

	if f.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else if f.ShowDeleted {
		query = query.Unscoped() // trae todos
	}

	query.Count(&total)
	query = query.Limit(int(f.Limit)).Offset(int(f.Offset))
	return query, total
}
