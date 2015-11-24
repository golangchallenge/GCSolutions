package main

import "fmt"

type Game interface {
	Board() Board
	Solve(b Board) (Board, error)
}

type CoordinateFinder interface {
	NextOpenCoordinate(board Board, xy XY) (*Coord, bool)
}

type SudokuGame struct {
	config    *Config
	events    EventHandler
	useEvents bool
	finder    CoordinateFinder
}

func NewSudokuGame(config *Config) *SudokuGame {
	return &SudokuGame{config: config, events: &NoEventHandler{}, useEvents: false, finder: &ClosestCoordFinder{}}
}

func (game *SudokuGame) SetCoordinateRanker(finder CoordinateFinder) {
	game.finder = finder
}

func (game *SudokuGame) AddEventHandler(handler EventHandler) {
	game.turnOnEvents()
	if game.events == nil {
		game.events = handler
	} else {
		if existing, ok := game.events.(*MultiEventHandler); ok {
			existing.handlers = append(existing.handlers, handler)
		} else {
			if existing == nil {
				game.events = handler
			} else {
				multi := NewMultiEventHandler(existing, handler)
				game.events = multi
			}
		}
	}
}

func (game *SudokuGame) Board() Board {
	game.turnOffEvents()
	defer game.turnOnEvents()
	if game.config.Generate {
		return game.generateBoard()
	} else {
		return game.readBoard()
	}
}

func (game *SudokuGame) turnOffEvents() {
	game.useEvents = false
}
func (game *SudokuGame) turnOnEvents() {
	game.useEvents = true
}

func (game *SudokuGame) readBoard() Board {
	fmt.Println("Reading Board...")

	board, err := BoardFromReader(game.config.UserInput)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if !game.config.InteractiveInput {
		fmt.Printf("%v\n\n", board)
	}

	if board.Conflict() {
		fmt.Println("Board is Invalid")
		return nil
	}

	if _, err := game.Solve(board); err != nil {
		fmt.Println("Board is Not Solvable")
		return nil
	}

	fmt.Println("Board is Valid and Solvable")
	displayDifficulty(board, game.config)
	return board
}

func (game *SudokuGame) generateBoard() Board {
	fmt.Println("Generating Board...")

	var board Board
	var created bool
	for attempts := 1; attempts < 10; attempts++ {
		board = CreateInitialBoardWithStartingCountAndSize(9, 10)

		if board.Conflict() {
			fmt.Println("... still working")
			logger.Debug("Generated board has conflicts [%d]", attempts)
			continue
		}
		if solved, err := game.Solve(board); err != nil {
			fmt.Println("... still working")
			logger.Debug("Generated board unsolvable [%d]: %v", attempts, err)
			continue
		} else {
			board = solved
		}

		ReduceBoardToDifficulty(board, game.config.Difficulty)
		created = true
		break
	}

	if !created {
		fmt.Println("Could not generate a valid and solvable board")
		return nil
	}

	fmt.Printf("%v\n\n", board)
	fmt.Println("Board is Valid and Solvable")
	displayDifficulty(board, game.config)
	return board
}

func displayDifficulty(board Board, config *Config) {
	if config.ShowDifficulty {
		var msg string
		var difficulty BoardDifficulty

		if config.Generate {
			difficulty = BoardDifficulty(config.Difficulty)
		} else {
			difficulty = EstimateDifficulty(board)
		}

		switch {
		case difficulty == EasyBoard:
			msg = "Easy"
		case difficulty > EasyBoard && difficulty < MediumBoard:
			msg = "Easy-Medium"
		case difficulty == MediumBoard:
			msg = "Medium"
		case difficulty > MediumBoard && difficulty < HardBoard:
			msg = "Medium-Hard"
		case difficulty == HardBoard:
			msg = "Hard"
		default:
			msg = "Unknown"
		}
		fmt.Printf("Game Difficulty is %v\n", msg)
	}
}

func (game *SudokuGame) Solve(b Board) (Board, error) {
	board := b.Clone()
	if board.Conflict() {
		return nil, fmt.Errorf("Board is invalid")
	}
	c, ok := game.finder.NextOpenCoordinate(board, &Coord{0, 0})
	if !ok {
		return nil, fmt.Errorf("Could not locate a valid coordinate to start with")
	}

	valuesToTry := board.AvailableValuesAtCoordinate(c)
	if len(valuesToTry) == 0 {
		return nil, fmt.Errorf("Could not determine any valid values for open coordinate at %+v", c)
	}

	var errors []error
	for _, val := range valuesToTry {
		if err := game.solveNextValue(board, c, val); err == nil {
			game.OnSuccessfulCoord(board, c)
			return board, nil
		} else {
			errors = append(errors, err)
		}
	}

	return nil, fmt.Errorf("Error Messages\n%v", errors)
}

func (game *SudokuGame) solveNextValue(board Board, coord XY, testValue byte) (err error) {
	game.OnAttemptingCoord(board, coord)
	if err := board.WriteSafe(coord, testValue); err != nil {
		game.OnFailedCoord(board, coord)
		return newTraceBackError(coord, testValue, err.Error())
	}

	defer func(board Board, coord XY) {
		if err != nil {
			// clear/rollback value if result ends up being false
			game.OnFailedCoord(board, coord)
			game.OnBeforeClearCoord(board, coord)
			board.Clear(coord)
			game.OnAfterClearCoord(board, coord)
		}
	}(board, coord)

	if board.Conflict() {
		err = newTraceBackError(coord, testValue, "Value creates a board conflict")
		return
	}

	logger.Debug("-----------\n%v\n", board)

	nextCoord, hasMore := game.finder.NextOpenCoordinate(board, coord)

	if !hasMore {
		return
	}

	valuesToTry := board.AvailableValuesAtCoordinate(nextCoord)
	if len(valuesToTry) == 0 {
		err = newTraceBackError(coord, testValue, "No available values to try at coordinate")
		return
	}

	nestedErrContainer := newTraceBackError(coord, testValue, "Nested Error")
	for _, v := range valuesToTry {
		if nestedErr := game.solveNextValue(board, nextCoord, v); nestedErr == nil {
			game.OnSuccessfulCoord(board, nextCoord)
			return // success
		} else {
			nestedErrContainer.Add(nestedErr)
		}
	}

	err = nestedErrContainer
	return
}

func (e *SudokuGame) OnAttemptingCoord(board Board, coord XY) {
	if e.useEvents && e.events != nil {
		e.events.OnAttemptingCoord(board, coord)
	}
}
func (e *SudokuGame) OnBeforeClearCoord(board Board, coord XY) {
	if e.useEvents && e.events != nil {
		e.events.OnBeforeClearCoord(board, coord)
	}
}
func (e *SudokuGame) OnAfterClearCoord(board Board, coord XY) {
	if e.useEvents && e.events != nil {
		e.events.OnAfterClearCoord(board, coord)
	}
}
func (e *SudokuGame) OnSuccessfulCoord(board Board, coord XY) {
	if e.useEvents && e.events != nil {
		e.events.OnSuccessfulCoord(board, coord)
	}
}
func (e *SudokuGame) OnFailedCoord(board Board, coord XY) {
	if e.useEvents && e.events != nil {
		e.events.OnFailedCoord(board, coord)
	}
}

type traceBackError struct {
	coord          XY
	attemptedValue byte
	reason         string
	nestedErrors   []error
}

func newTraceBackError(coord XY, value byte, msg string) *traceBackError {
	return &traceBackError{coord: coord, attemptedValue: value, reason: msg}
}
func (e *traceBackError) Add(nestedError error) {
	e.nestedErrors = append(e.nestedErrors, nestedError)
}
func (e *traceBackError) Error() string {
	return fmt.Sprintf("Value %v cannot be placed at coordinate %+v with reason: %v\n%v", e.attemptedValue, e.coord, e.reason, e.nestedErrors)
}
