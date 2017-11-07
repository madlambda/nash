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
        map <= map_add($map, "key", "value2")
        expect($map, "key", "value2")
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

test_adding_keys()
test_absent_key_will_have_empty_string_value()
test_absent_key_with_custom_default_value()

# TODO: test iteration
