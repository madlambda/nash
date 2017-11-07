fn map_new() {
        map = ()
        return $map
}

fn map_get(map, key) {
        return map_get_default($map, $key, "")
}

fn map_iter(map, func) {
        for entry in $map {
                $func($entry[0], $entry[1])
        }
}

fn map_get_default(map, key, default) {
        for entry in $map {
                if $entry[0] == $key {
                        return $entry[1]
                }
        }

        return $default
}

fn map_add(map, key, val) {
        for entry in $map {
                if $entry[0] == $key {
                        entry[1] = $val
                        return $map
                }
        }

        tuple = ($key $val)
        map <= append($map, $tuple)
        return $map
}
