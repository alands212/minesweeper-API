package minesweeper

import (
	"errors"
	"github.com/google/uuid"
	"github.com/tomiok/minesweeper-API/internal/logs"
	"github.com/tomiok/minesweeper-API/models"
	"github.com/tomiok/minesweeper-API/storage"
)

const (
	defaultRows  = 6
	defaultCols  = 6
	defaultMines = 12
	maxMines     = 25
	maxRows      = 30
	maxCols      = 30
)

type GameService struct {
	gameStorage models.MineSweeperStorage
	userStorage models.UserStorage
}

func NewGameService(db *storage.DB) models.MineSweeperService {
	return &GameService{
		gameStorage: storage.NewGameEngineStorage(db),
		userStorage: storage.NewUserStorage(db),
	}
}

func (s *GameService) CreateGame(game *models.Game) error {
	username := game.Username
	if username == "" {
		return errors.New("username empty is not allowed")
	}

	_, err := s.userStorage.GetByName(username)

	if err != nil {
		return errors.New("cannot find username")
	}

	if game.Name == "" {
		game.Name = getUUIDName()
	}

	if game.Rows == 0 {
		game.Rows = defaultRows
	}

	if game.Cols == 0 {
		game.Cols = defaultCols
	}

	if game.Mines == 0 {
		game.Mines = defaultMines
	}

	if game.Mines > maxMines {
		game.Mines = maxMines
	}

	if game.Rows > maxRows {
		game.Rows = maxRows
	}
	if game.Cols > maxCols {
		game.Cols = maxCols
	}
	if game.Mines > (game.Cols * game.Rows) {
		game.Mines = game.Cols * game.Rows
	}
	game.Status = "new"

	err = s.gameStorage.Create(game)
	return err
}

func (s *GameService) Start(name string) (*models.Game, error) {
	game, err := s.gameStorage.GetByName(name)
	if err != nil {
		return nil, err
	}

	buildBoard(game)

	game.Status = "in_progress"
	err = s.gameStorage.Update(game)
	logs.Sugar().Infof("%#v\n", game.Grid)
	return game, err
}

func (s *GameService) Click(name, clickType string, i, j int) (*models.Game, error) {
	game, err := s.gameStorage.GetByName(name)
	if err != nil {
		return nil, err
	}

	if isNormalClick(clickType) {
		if err := clickCell(game, i, j); err != nil {
			return nil, err
		}
	} else {
		if err := flagOrQuestionMarkCell(game, i, j, clickType); err != nil {
			return nil, err
		}
	}

	if err := s.gameStorage.Update(game); err != nil {
		return nil, err
	}

	return game, nil
}

func isNormalClick(clickType string) bool {
	return clickType == "click"
}

func getUUIDName() string {
	return uuid.New().String()
}