package helper

import (
	"strconv"
	"strings"

	"api-students/app/model"

	"github.com/gofiber/fiber/v2"
)

func ParamID(c *fiber.Ctx) (int, bool) {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	search := strings.TrimSpace(c.Query("search"))
	cursorStr := strings.TrimSpace(c.Query("cursor"))

	var after *model.Cursor
	if cursorStr != "" {
		decoded, err := DecodeCursor(cursorStr)
		if err != nil {
			return model.CursorQuery{}, err
		}
		after = &decoded
	}

	return model.CursorQuery{
		Limit:  limit,
		Search: search,
		After:  after,
	}, nil
}