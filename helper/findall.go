package helper

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type FindAllOptions struct {
	OrderBy     string
	Sort        string
	Search      string
	SearchValue string
	Limit       uint
	Offset      uint
	ShowDeleted bool
	OnlyDeleted bool
}

func NewFindAllOptionsFromQuery(c *fiber.Ctx) *FindAllOptions {
	limitParam := c.Query("limit", "10")
	offsetParam := c.Query("offset", "0")

	limit, _ := strconv.ParseUint(limitParam, 10, 32)
	offset, _ := strconv.ParseUint(offsetParam, 10, 32)

	return &FindAllOptions{
		OrderBy:     c.Query("order_by", "created_at"),
		Sort:        c.Query("sort", "desc"),
		Search:      c.Query("search", ""),
		SearchValue: c.Query("search_value", ""),
		Limit:       uint(limit),
		Offset:      uint(offset),
		ShowDeleted: c.QueryBool("show_deleted", false),
		OnlyDeleted: c.QueryBool("only_deleted", false),
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

	if f.Search != "" {
		if f.SearchValue == "" {
			f.SearchValue = "name"
		}
		query = query.Where(f.Search+" ILIKE ?", "%"+f.SearchValue+"%")
	}

	query.Count(&total)
	query = query.Limit(int(f.Limit)).Offset(int(f.Offset))
	return query, total
}
