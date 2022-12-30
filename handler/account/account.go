package account

import (
	"database/sql"
	"net/http"

	hdr "github.com/kkgoo-software-engineering/workshop/handler"
	"github.com/kkgoo-software-engineering/workshop/internal/config"
	"github.com/kkgoo-software-engineering/workshop/internal/middleware/zapmw"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Account struct {
	ID      int64   `json:"id"`
	Balance float64 `json:"balance"`
}

type handler struct {
	cfgFlag *config.FeatureFlag
	db      *sql.DB
}

func New(cfgFlag *config.FeatureFlag, db *sql.DB) *handler {
	return &handler{cfgFlag, db}
}

const (
	cStmt         = "INSERT INTO accounts (balance) VALUES ($1) RETURNING id;"
	cBalanceLimit = 10000
)

var (
	hErrBalanceLimitExceed = echo.NewHTTPError(http.StatusBadRequest,
		"create account balance exceed limitation")
)

func (h handler) Create(c echo.Context) error {
	logger := zapmw.Logger(c)
	ctx := c.Request().Context()
	var ac Account
	err := c.Bind(&ac)
	if err != nil {
		logger.Error("bad request body", zap.Error(err))
		return hdr.ErrBadRequest
	}

	if h.cfgFlag.IsLimitMaxBalanceOnCreate && ac.Balance > cBalanceLimit {
		logger.Error("account limit on account creating", zap.Error(hErrBalanceLimitExceed))
		return hErrBalanceLimitExceed
	}

	var lastInsertId int64
	err = h.db.QueryRowContext(ctx, cStmt, ac.Balance).Scan(&lastInsertId)
	if err != nil {
		logger.Error("query row error", zap.Error(err))
		return err
	}

	logger.Info("create successfully", zap.Int64("id", lastInsertId))
	ac.ID = lastInsertId
	return c.JSON(http.StatusCreated, ac)
}
