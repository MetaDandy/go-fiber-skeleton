package helper

import (
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type FindAllOptions struct {
	Search      string `query:"search"`
	Limit       uint   `query:"limit,default:10"`
	Offset      uint   `query:"offset,default:0"`
	ShowDeleted bool   `query:"show_deleted,default:false"`
	OnlyDeleted bool   `query:"only_deleted,default:false"`
}

func NewFindAllOptionsFromQuery(c fiber.Ctx) *FindAllOptions {
	q := new(FindAllOptions)

	if err := c.Bind().Query(q); err != nil {
		q = &FindAllOptions{
			Limit:  10,
			Offset: 0,
		}
	}

	return &FindAllOptions{
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
		query.Count(&total)
		return query, total
	}

	if f.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else if f.ShowDeleted {
		query = query.Unscoped()
	}

	query.Count(&total)
	query = query.Limit(int(f.Limit)).Offset(int(f.Offset))
	return query, total
}
