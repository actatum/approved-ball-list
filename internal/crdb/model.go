package crdb

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/actatum/approved-ball-list/internal/abl"
	"github.com/actatum/errs"
)

type ball struct {
	ID         int64     `db:"id"`
	Brand      string    `db:"brand"`
	Name       string    `db:"name"`
	ApprovedAt time.Time `db:"approved_at"`
	ImageURL   string    `db:"image_url"`
}

type listBallRow struct {
	ball
	Count int `db:"count"`
}

const (
	insertBallsQuery = `
	INSERT INTO balls (
		brand,
		name,
		approved_at,
		image_url
	) VALUES (
		:brand,
		:name,
		:approved_at,
		:image_url
	)`
)

func listBallsQuery(
	sb *sq.StatementBuilderType,
	filter abl.BallFilter,
) (string, []interface{}, int, error) {
	offset, err := calculateOffset(filter.PageToken)
	if err != nil {
		return "", nil, 0, fmt.Errorf("calculateOffset: %w", err)
	}

	q := sb.Select(
		"id",
		"brand",
		"name",
		"approved_at",
		"image_url",
		"COUNT(*) OVER() as count",
	).From(
		"balls",
	)

	if filter.Brand != nil {
		q = q.Where(sq.Eq{"brand": *filter.Brand})
	}

	q = q.OrderBy(
		"id DESC",
	)
	if filter.PageSize > 0 {
		q = q.Limit(
			uint64(filter.PageSize + 1),
		).Offset(
			uint64(offset),
		)
	}

	query, args, err := q.ToSql()
	if err != nil {
		return "", nil, 0, fmt.Errorf("q.ToSql: %w", err)
	}

	return query, args, offset, nil
}

func calculateOffset(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}

	token, err := base64.URLEncoding.DecodeString(pageToken)
	if err != nil {
		return 0, errs.Errorf(errs.Invalid, "invalid page token: %s", err.Error())
	}

	offset, err := strconv.Atoi(string(token))
	if err != nil {
		return 0, errs.Errorf(errs.Invalid, "invalid page token: %s", err.Error())
	}

	return offset, nil
}
