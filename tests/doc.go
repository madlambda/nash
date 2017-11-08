// Package tests contains all nash tests that are blackbox.
// What would be blackbox ? These are tests that are targeted
// directly on top of the language using only the shell API,
// they are end to end in the sense that they will exercise
// a lot of different packages on a single test.
//
// The objective of these tests is to have a compreensive set
// of tests that are coupled only the language specification
// and not to how the language is implemented. These allows
// extremely aggressive refactorings to be made without
// incurring in any changes on the tests.
//
// There are disadvantages but discussing integration VS unit
// testing here is not the point (there are also unit tests).
//
// Here even tests that involves the script calling syscalls like
// exit are allowed without interfering with the results of other tests
package tests
