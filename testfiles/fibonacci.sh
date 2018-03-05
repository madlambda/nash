#!/usr/bin/env nash

# Recursive fibonacci implementation to find the value
# at index n in the sequence.

# Some times:
#   λ> time ./testfiles/fibonacci.sh 1
#   1
#   0.00u 0.01s 0.01r 	 ./testfiles/fibonacci.sh 1
#   λ> time ./testfiles/fibonacci.sh 2
#   1 2
#   0.01u 0.01s 0.02r 	 ./testfiles/fibonacci.sh 2
#   λ> time ./testfiles/fibonacci.sh 3
#   1 2 3
#   0.01u 0.03s 0.03r 	 ./testfiles/fibonacci.sh 3
#   λ> time ./testfiles/fibonacci.sh 4
#   1 2 3 5
#   0.04u 0.04s 0.07r 	 ./testfiles/fibonacci.sh 4
#   λ> time ./testfiles/fibonacci.sh 5
#   1 2 3 5 8
#   0.09u 0.07s 0.13r 	 ./testfiles/fibonacci.sh 5
#   λ> time ./testfiles/fibonacci.sh 10
#   1 2 3 5 8 13 21 34 55 89
#   1.31u 1.18s 2.03r 	 ./testfiles/fibonacci.sh 10
#   λ> time ./testfiles/fibonacci.sh 15
#   1 2 3 5 8 13 21 34 55 89 144 233 377 610 987
#   15.01u 13.49s 22.55r 	 ./testfiles/fibonacci.sh 15
#   λ> time ./testfiles/fibonacci.sh 20
#   1 2 3 5 8 13 21 34 55 89 144 233 377 610 987 1597 2584 4181 6765 10946
#   169.27u 155.50s 265.19r 	 ./testfiles/fibonacci.sh 20

# a is lower or equal than b?
fn lte(a, b) {
	var _, st <= test $a -le $b

	return $st
}

fn fib(n) {
	if lte($n, "1") == "0" {
		return "1"
	}

	var a, _ <= expr $n - 1
	var b, _ <= expr $n - 2
	var _a <= fib($a)
	var _b <= fib($b)
	var ret, _ <= expr $_a "+" $_b

	return $ret
}

fn range(start, end) {
	var seq, _ <= seq $start $end
	var lst <= split($seq, "\n")

	return $lst
}

for i in range("1", $ARGS[1]) {
	print("%s ", fib($i))
}

print("\n")
