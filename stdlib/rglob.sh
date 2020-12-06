fn dir_glob(dir, pattern, results) {
	files <= find $dir -maxdepth 1 -mindepth 1
	files <= split($files, "\n")

	for f in $files {
		_, status <= test -d $f

		if $status == "0" {
			dir_pattern     <= format("%s/%s", $f, $pattern)
			current_results <= glob($dir_pattern)
			
			for r in $current_results {
				results <= append($results, $r)
			}
			
			results <= dir_glob($f, $pattern, $results)
		}
	}

	return $results
}

fn rglob(pattern) {
	working_dir <= pwd
	result      <= glob($working_dir+"/"+$pattern)
	result      <= dir_glob($working_dir, $pattern, $result)

	return $result
}
