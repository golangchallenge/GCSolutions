//Author: Amauris De Jesus

HOW TO RUN:
	- Execute main.go in root file, ex:
		go run main.go "_ 5 7 8 _ _ 9 _ _ 
						4 1 9 _ _ _ 3 _ _ 
						_ 8 _ _ 9 _ _ _ 1 
						_ _ _ 3 _ 9 5 _ _ 
						_ 9 _ _ 8 _ _ 2 _ 
						_ _ 4 5 _ 7 _ _ _ 
						9 _ _ _ 7 _ _ 1 _ 
						_ _ 1 _ _   _ 4 3 8
						_ _ 8 _ _ 3 6 7 _"

CODE DETAILS:
	- The package includes two sub packages, sodoku, and solutions.
	- The sodoku sub package, implements the game's mechanics, such as building the
	9x9 board, setting values, and checking if a board is complete.
	- The solutions package, implements the alogrithm which solves the sodoku object.
	  Solutionizer is in charge of executing the algorithm described below.
	- The tester.php scripts is a simple script use to fetch results from websudoku.com
	  in order to get a wide range of puzzles to test with. You can call it for example,
	  php tester.php [level_flag]
	  The level flag can range from 1-4 and determines the available difficulties on websodoku.com

ALGORITHM:
	- The solution is solved recursively for simplicity. The algorithm is simple and straighforward.
	- It first looks for entries in the sodoku board which are solvable with 100% certainty. For example,
	1 2 3 4 5 6 7 8 _
	_ _ _ _ 6 _ _ _ _
	_ _ _ _ 7 _ _ _
	_ _ _ _ 8 _ _ _ _
	_ _ _ _ _ _ _ _ _
	_ _ _ _ _ _ _ _ _
	_ _ _ _ _ _ _ _ _
	1 2 3 4 _ _ _ _ _
	_ _ _ _ _ _ _ _ _
	Two columns here are solvable right away. The indices at (row 0, column 8), and (row 7, column 4)
	(counting from zero). This is because for both indices, looking at their respective families, only 
	have one option available. For (row 0, column 8) only 9 is available so we can immediatly insert that.
	And for (row 7, column 4) only 9 is available as well because its the only value left with repect to
	its row and column families.
	- If no entries were found that can be filled with 100% certainty then we try to solve the solution by filling
	the entries with least possiblities to choose from first, and then call the algorithm recursively and see if it solves.

TERMINOLOGY:
	Family: For an index i,j from the multidimensional representation of the sodoku board,
		a family represents the groups in which the value for that index must be unique.
		For example, the family for the value located at index 1,2, will be all the values
		at row 1, all the values at column 2, and the quadrant in which 1,2 is part of.
		So every index has a set of three families for which it's value must be unique with
		respect to each family.
	Relative:
		Every value/member of each family for index i,j, is considered a relative of index i,j.