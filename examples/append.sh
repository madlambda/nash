#!/usr/bin/env nash

var example_list = ()
echo "appending string 1"
example_list <= append($example_list, "1")
echo $example_list
echo "appending string 2"
example_list <= append($example_list, "2")
echo $example_list
