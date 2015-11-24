<?php

//stress tester. Retrieved table from websodoku, parses it and passes it 
//into my own sodoku solver

$inputTable = getSodokuTable($argv[1]);

echo "running with the following input\n".$inputTable." \n";
$output = shell_exec("cd /Users/amauris/workspace/go/sodoku_solutioner; go run main.go \"".$inputTable."\"");
echo "\n\nFinished with the following output\n".$output."\n";

function getSodokuTable($difficultyLevel) {

	$html = file_get_contents("http://show.websudoku.com/?level=$difficultyLevel");
	preg_match("/\"puzzle_grid\"(.*?)<\/table/si", $html, $table);
	$table = $table[1];

	preg_match_all("/<input(.*?)>/si", $table, $inputs);
	$inputs = $inputs[1];
	
	$inputString = "";
	$i = 0;
	foreach($inputs as $input) {
		preg_match("/VALUE\=\"(.*?)\"/si", $input, $value);

		if($i>=9)
		{
			$i = 0;
			$inputString .= "\n";
		}

		
		if(empty($value)) 
			$inputString .= "_ ";
		else 
			$inputString .= trim($value[1])." ";


		$i += 1;
	}

	return $inputString;
}
