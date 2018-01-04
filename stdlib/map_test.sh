import map

fn expect(map, key, want) {
        got <= map_get($map, $key)

        if $got != $want {
                echo "error: got["+$got+"] want["+$want+"]"
                exit("1")
        }
}

fn test_adding_keys() {
        map <= map_new()

        map <= map_add($map, "key", "value")
        expect($map, "key", "value")

        map <= map_add($map, "key2", "value2")
        expect($map, "key2", "value2")

        map <= map_add($map, "key", "override")
        expect($map, "key", "override")
}

fn test_absent_key_will_have_empty_string_value() {
        map <= map_new()
        expect($map, "absent", "")
}

fn test_absent_key_with_custom_default_value() {
        map <= map_new()
        want = "hi"
        got <= map_get_default($map, "absent", $want)
        if $got != $want {
                echo "error: got["+$got+"] want["+$want+"]"
                exit("1")
        }
}

fn test_iterates_map() {
        map <= map_new()
        map <= map_add($map, "key", "value")
        map <= map_add($map, "key2", "value2")

        got <= map_new()

	fn iter(key, val) {
		got <= map_add($got, $key, $val)
	}

        map_iter($map, $iter)

        expect($map, "key", "value")
        expect($map, "key2", "value2")
}

fn test_removing_key() {
        map <= map_new()

        map <= map_add($map, "key", "value")
        map <= map_add($map, "key2", "value2")

        expect($map, "key", "value")
        expect($map, "key2", "value2")

        map <= map_del($map, "key")
        expect($map, "key", "")
        expect($map, "key2", "value2")
}

fn test_removing_absent_key() {
        map <= map_new()

        expect($map, "key", "")
        map <= map_del($map, "key")
        expect($map, "key", "")
}

test_adding_keys()
test_absent_key_will_have_empty_string_value()
test_absent_key_with_custom_default_value()
test_iterates_map()
test_removing_key()
test_removing_absent_key()
