package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type ProgressEventHandler struct {
	ShowColors       bool
	Delay            time.Duration
	Writer           io.Writer
	goodCoords       []*Coord
	failedCoords     []*Coord
	inProgressCoords []*Coord
}

func NewProgressEventHandler(colors bool, delay time.Duration) *ProgressEventHandler {
	return &ProgressEventHandler{ShowColors: colors, Delay: delay, Writer: os.Stdout}
}

func (e *ProgressEventHandler) OnAttemptingCoord(board Board, coord XY) {
	e.delay()
	e.inProgressCoords = append(e.inProgressCoords, CoordXY(coord))
	inprogressCodeSet := NewColorCoordSet(e.inProgressCoords, StatusInProgressColor)
	successCodeSet := NewColorCoordSet(e.goodCoords, StatusSuccessColor)

	e.clearConsole()
	e.write(fmt.Sprintf("Attempting Coordinate %v\n%v", coord, ColorizeBoard(board, inprogressCodeSet, successCodeSet)))
}
func (e *ProgressEventHandler) OnBeforeClearCoord(board Board, coord XY) {}
func (e *ProgressEventHandler) OnAfterClearCoord(board Board, coord XY)  {}
func (e *ProgressEventHandler) OnSuccessfulCoord(board Board, coord XY) {
	e.delay()
	e.removeInProgressCoord(coord)
	e.goodCoords = append(e.goodCoords, CoordXY(coord))
	//failedCodeSet := NewColorCoordSet([]*Coord{CoordXY(coord)}, StatusFailedColor)
	inprogressCodeSet := NewColorCoordSet(e.inProgressCoords, StatusInProgressColor)
	successCodeSet := NewColorCoordSet(e.goodCoords, StatusSuccessColor)

	e.clearConsole()
	e.write(fmt.Sprintf("Successful Coordinate %v\n%v", coord, ColorizeBoard(board, inprogressCodeSet, successCodeSet)))
}
func (e *ProgressEventHandler) OnFailedCoord(board Board, coord XY) {
	e.delay()
	e.removeInProgressCoord(coord)
	failedCodeSet := NewColorCoordSet([]*Coord{CoordXY(coord)}, StatusFailedColor)
	inprogressCodeSet := NewColorCoordSet(e.inProgressCoords, StatusInProgressColor)
	successCodeSet := NewColorCoordSet(e.goodCoords, StatusSuccessColor)

	e.clearConsole()
	e.write(fmt.Sprintf("Failed Coordinate %v\n%v", coord, ColorizeBoard(board, inprogressCodeSet, successCodeSet, failedCodeSet)))
}

func (e *ProgressEventHandler) clearConsole() {
	e.write(fmt.Sprintf("%v%v", ClearConsole, ResetCursor))
}
func (e *ProgressEventHandler) delay() {
	time.Sleep(e.Delay)
}
func (e *ProgressEventHandler) write(output string) {
	if e.Writer != nil {
		strings.NewReader(output).WriteTo(e.Writer)
	}
}
func (e *ProgressEventHandler) removeInProgressCoord(coord XY) {
	var index int
	for i, p := range e.inProgressCoords {
		if p.x == coord.X() && p.y == coord.Y() {
			index = i
		}
	}

	if index > 0 {
		e.inProgressCoords = append(e.inProgressCoords[:index], e.inProgressCoords[index+1:]...)
	}
}
