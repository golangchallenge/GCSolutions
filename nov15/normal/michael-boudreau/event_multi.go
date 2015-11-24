package main

type MultiEventHandler struct {
	handlers []EventHandler
}

func NewMultiEventHandler(handlers ...EventHandler) *MultiEventHandler {
	return &MultiEventHandler{handlers: handlers}
}
func (e *MultiEventHandler) OnAttemptingCoord(board Board, coord XY) {
	for _, h := range e.handlers {
		h.OnAttemptingCoord(board, coord)
	}
}
func (e *MultiEventHandler) OnBeforeClearCoord(board Board, coord XY) {
	for _, h := range e.handlers {
		h.OnBeforeClearCoord(board, coord)
	}
}
func (e *MultiEventHandler) OnAfterClearCoord(board Board, coord XY) {
	for _, h := range e.handlers {
		h.OnAfterClearCoord(board, coord)
	}
}
func (e *MultiEventHandler) OnSuccessfulCoord(board Board, coord XY) {
	for _, h := range e.handlers {
		h.OnSuccessfulCoord(board, coord)
	}
}
func (e *MultiEventHandler) OnFailedCoord(board Board, coord XY) {
	for _, h := range e.handlers {
		h.OnFailedCoord(board, coord)
	}
}
