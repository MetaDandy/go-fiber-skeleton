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
		Limit:       uint(limit),
		Offset:      uint(offset),
		ShowDeleted: c.QueryBool("show_deleted", false),
		OnlyDeleted: c.QueryBool("only_deleted", false),
	}
}

func ApplyFindAllOptions(query *gorm.DB, opts *FindAllOptions) (*gorm.DB, int64) {
	var total int64

	if opts == nil {
		query = query.Order("created_at asc")
		query.Count(&total)
		return query, total
	}

	orderBy := opts.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}

	sort := "asc"
	if opts.Sort == "desc" {
		sort = "desc"
	}

	query = query.Order(orderBy + " " + sort)

	if opts.OnlyDeleted {
		query = query.Unscoped().Where("deleted_at IS NOT NULL")
	} else if opts.ShowDeleted {
		query = query.Unscoped() // trae todos
	}

	if opts.Search != "" {
		query = query.Where("name ILIKE ?", "%"+opts.Search+"%")
	}

	query.Count(&total)
	query = query.Limit(int(opts.Limit)).Offset(int(opts.Offset))
	return query, total
}
