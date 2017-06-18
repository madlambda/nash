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

fn le(a, b) {
	var _, st <= test $a -le $b

	return $st
}

fn sqrt(n) {
	var v, _ <= expr $n * $n

	return $v
}

fn range(start, end) {
	var values, _ <= seq $start $end
	var list <= split($values, "\n")

	return $list
}

fn xrange(start, condfn) {
	var out = ()

	if $condfn($start) == "0" {
		out = ($start)
	} else {
		return ()
	}

	var next = $start

	for {
		next, _ <= expr $next "+" 1

		if $condfn($next) == "0" {
			out <= append($out, $next)
		} else {
			return $out
		}
	}

	unreachable
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

	fn untilSqrtRoot(v) {
		return le(sqrt($v), $n)
	}

	for i in xrange("2", $untilSqrtRoot) {
		if $tries[$i] == "1" {
			for j in range("0", $n) {
				# arithmetic seems cryptic without integers =(
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

for prime in sieve($ARGS[1]) {
	print("%s ", $prime)
}

print("\n")
