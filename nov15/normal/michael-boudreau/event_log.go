package main

import (
	"log"
)

type LogEventHandler struct {
}

func NewLogEventHandler() *LogEventHandler {
	return &LogEventHandler{}
}
func (e *LogEventHandler) OnAttemptingCoord(board Board, coord XY) {
	log.Printf("Attempting Coordinate %+v", coord)
}
func (e *LogEventHandler) OnBeforeClearCoord(board Board, coord XY) {
	log.Printf("Before Clearing Coordinate %+v", coord)
}
func (e *LogEventHandler) OnAfterClearCoord(board Board, coord XY) {
	log.Printf("After Clearing Coordinate %+v", coord)
}
func (e *LogEventHandler) OnSuccessfulCoord(board Board, coord XY) {
	log.Printf("Successful Coordinate %+v", coord)
}
func (e *LogEventHandler) OnFailedCoord(board Board, coord XY) {
	log.Printf("Failed Coordinate %+v", coord)
}
