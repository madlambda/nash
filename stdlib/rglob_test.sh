#!/usr/bin/env nash

import rglob

fn setup() {
	temp_dir <= mktemp -d

	test_txt  = $temp_dir+"/test.txt"
	test_sh   = $temp_dir+"/test.sh"

	_, _      <= touch $test_txt
	_, _      <= touch $test_sh

	hello_dir = $temp_dir+"/hello"
	world_dir = $temp_dir+"/world"

	_, _      <= mkdir $hello_dir
	_, _      <= mkdir $world_dir

	honda_txt = $hello_dir+"/honda.txt"
	honda_sh  = $hello_dir+"/honda.sh"
	civic_txt = $world_dir+"/civic.txt"
	civic_sh  = $world_dir+"/civic.sh"

	_, _      <= touch $honda_txt
	_, _      <= touch $honda_sh
	_, _      <= touch $civic_txt
	_, _      <= touch $civic_sh

	return $temp_dir
}

fn assert_success(expected, got) {
	if len($got) != len($expected) {
		print("expected length to be '%s' but got '%s'\n", len($expected), len($got))
		exit("1")
	}

	found = "0"

	for x in $expected {
		for y in $got {
			if $x == $y {
				found <= echo $found+" + 1" | bc
			}
		}
	}

	if $found != len($expected) {
		print("expected [%s] got [%s]\n", $expected, $got)
		exit("1")
	}
}

fn test_rglob_txt(dir) {
	expected_files = (
		$dir+"/test.txt"
		$dir+"/hello/honda.txt"
		$dir+"/world/civic.txt"
	)

	got_files <= rglob("*.txt")

	assert_success($expected_files, $got_files)
}

fn test_rglob_sh(dir) {
	expected_files = (
		$dir+"/test.sh"
		$dir+"/hello/honda.sh"
		$dir+"/world/civic.sh"
	)

	got_files <= rglob("*.sh")

	assert_success($expected_files, $got_files)
}

old_dir  <= pwd
temp_dir <= setup()

chdir($temp_dir)
test_rglob_txt($temp_dir)
test_rglob_sh($temp_dir)
chdir($old_dir)
