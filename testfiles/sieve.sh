#!/usr/bin/env nash

# Sieve of Erathostenes

fn lt(a, b) {
	var _, st <= test $a -lt $b

	return $st
}

fn gt(a, b) {
	var _, st <= test $a -gt $b

	return $st
}

fn gte(a, b) {
	var _, st <= test $a -ge $b

	return $st
}

fn range(start, end) {
	var values, _ <= seq $start $end
	var list <= split($values, "\n")

	return $list
}

fn sieve(n) {
	if lt($n, "2") == "0" {
		return ()
	}
	if $n == "2" {
		return ("2")
	}

	var tries = ("0" "0")

	for i in range("2", $n) {
		tries <= append($tries, "1")
	}
	for i in range("2", $n) {
		if $tries[$i] == "1" {
			for j in range("0", $n) {
				var k, _ <= expr $i * $i "+" "(" $j * $i ")"

				if gt($k, $n) != "0" {
					tries[$k] = "0"
				}
			}
		}
	}

	var primes = ()

	for i in range("2", $n) {
		if $tries[$i] == "1" {
			primes <= append($primes, $i)
		}
	}

	return $primes
}

for p in sieve($ARGS[1]) {
	print("%s ", $p)
}

print("\n")
