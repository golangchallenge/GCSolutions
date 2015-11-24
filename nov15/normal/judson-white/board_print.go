package main

import (
	"bytes"
	"fmt"
)

func (b *board) Print() {
	for i := 0; i < len(b.solved); i++ {
		if b.solved[i] == 0 {
			fmt.Print("_")
		} else {
			fmt.Printf("%d", b.solved[i])
		}
		if (i+1)%9 == 0 {
			fmt.Println()
		} else {
			fmt.Print(" ")
		}
	}
}

func (b *board) PrintPretty() {
	fmt.Print("|-------|-------|-------|\n| ")
	for i := 0; i < len(b.solved); i++ {
		if b.solved[i] == 0 {
			fmt.Print("_ ")
		} else {
			fmt.Printf("%d ", b.solved[i])
		}
		if (i+1)%9 == 0 {
			fmt.Print("|\n|")
			if (i+1)%27 == 0 {
				fmt.Print("-------|-------|-------|\n")
				if i != 80 {
					fmt.Print("| ")
				}
			} else {
				fmt.Print(" ")
			}
		} else if (i+1)%3 == 0 {
			fmt.Print("| ")
		}
	}
}

func (b *board) PrintCompact() {
	fmt.Println(b.GetCompact())
}

func (b *board) GetCompact() string {
	buf := bytes.NewBufferString("")
	for i := 0; i < 81; i++ {
		buf.WriteByte('0' + byte(b.solved[i]))
	}
	return buf.String()
}

func (b *board) PrintHints() {
	fmt.Printf(b.GetTextBoardWithHints())
}

func (b *board) GetTextBoardWithHints() string {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|\n"))
	buf.WriteString(fmt.Sprintf("|r,c| %15d %15d %15d | %15d %15d %15d | %15d %15d %15d |\n", 1, 2, 3, 4, 5, 6, 7, 8, 9))
	buf.WriteString(fmt.Sprintf("|---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|\n| A | "))
	for i := 0; i < len(b.solved); i++ {
		if b.solved[i] == 0 {
			buf.WriteString(fmt.Sprintf("%15s ", fmt.Sprintf("(%s)", GetBitsString(b.blits[i]))))
		} else {
			buf.WriteString(fmt.Sprintf("%15d ", b.solved[i]))
		}
		if (i+1)%9 == 0 {
			buf.WriteString("|\n|")
			textRow := getTextRow((i + 1) / 9)
			if (i+1)%27 == 0 {
				buf.WriteString("---|-------------------------------------------------|-------------------------------------------------|-------------------------------------------------|\n")
				if i != 80 {
					buf.WriteString(fmt.Sprintf("| %c | ", textRow))
				}
			} else {
				buf.WriteString(fmt.Sprintf(" %c | ", textRow))
			}
		} else if (i+1)%3 == 0 {
			buf.WriteString("| ")
		}
	}

	return buf.String()
}

func printCompactToStandard(b string) {
	i := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if c != 0 {
				fmt.Print(" ")
			}
			if b[i] == '0' {
				fmt.Print("_")
			} else {
				fmt.Print(string(b[i]))
			}
			i++
		}
		fmt.Println()
	}
}
