package main

type EventHandler interface {
	OnAttemptingCoord(board Board, coord XY)

	OnBeforeClearCoord(board Board, coord XY)
	OnAfterClearCoord(board Board, coord XY)

	OnSuccessfulCoord(board Board, coord XY)
	OnFailedCoord(board Board, coord XY)
}

type NoEventHandler struct{}

func (e *NoEventHandler) OnAttemptingCoord(board Board, coord XY)  {}
func (e *NoEventHandler) OnBeforeClearCoord(board Board, coord XY) {}
func (e *NoEventHandler) OnAfterClearCoord(board Board, coord XY)  {}
func (e *NoEventHandler) OnSuccessfulCoord(board Board, coord XY)  {}
func (e *NoEventHandler) OnFailedCoord(board Board, coord XY)      {}
