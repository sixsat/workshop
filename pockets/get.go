package pocket

import (
	"net/http"
	"regexp"

	"github.com/kkgo-software-engineering/workshop/mlog"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	gStmt   = `SELECT id, amount, name, "accountId", currency FROM tbl_pockets;`
	gbiStmt = `SELECT id, amount, name, "accountId", currency FROM tbl_pockets WHERE id = $1;`
)

func (h handler) Get(c echo.Context) error {

	logger := mlog.L(c)
	ctx := c.Request().Context()
	var ps []Pocket = []Pocket{}
	rows, err := h.db.QueryContext(ctx, gStmt)
	if err != nil {
		logger.Error("Query error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
	}
	defer rows.Close()
	for rows.Next() {
		var p Pocket
		err = rows.Scan(&p.ID, &p.Amount, &p.Name, &p.AccountId, &p.Currency)
		if err != nil {
			logger.Error("Scan error", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		}
		ps = append(ps, p)
	}
	return c.JSON(http.StatusOK, ps)
}

func (h handler) GetByID(c echo.Context) error {
	logger := mlog.L(c)
	ctx := c.Request().Context()

	id := c.Param("id")
	if id == "" {
		return c.JSON(
			http.StatusUnprocessableEntity,
			ErrorResponse{Message: "Param id is empty"},
		)
	}
	var p Pocket
	err := h.db.QueryRowContext(ctx, gbiStmt, id).Scan(&p.ID, &p.Amount, &p.Name, &p.AccountId, &p.Currency)
	if err != nil {
		match, errMatch := regexp.MatchString("invalid input syntax", err.Error())
		if match {
			logger.Error("Param id must be integer", zap.Error(err))
			return c.JSON(
				http.StatusUnprocessableEntity,
				ErrorResponse{Message: "Param id must be integer"},
			)
		}
		if errMatch != nil {
			logger.Error("Match fail", zap.Error(err))
			return c.JSON(
				http.StatusInternalServerError,
				ErrorResponse{Message: err.Error()},
			)
		}
		match, errMatch = regexp.MatchString("no rows in result set", err.Error())
		if match {
			logger.Error("Record not found", zap.Error(err))
			return c.JSON(
				http.StatusNotFound,
				ErrorResponse{Message: "Record not found"},
			)
		}
		if errMatch != nil {
			logger.Error("Match fail", zap.Error(err))
			return c.JSON(
				http.StatusInternalServerError,
				ErrorResponse{Message: err.Error()},
			)
		}
		logger.Error("Internal error", zap.Error(err))
		return c.JSON(
			http.StatusInternalServerError,
			ErrorResponse{Message: err.Error()},
		)
	}
	return c.JSON(http.StatusOK, p)
}
