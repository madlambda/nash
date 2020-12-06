
# Single assignment
var a = "";
var a = "something";
var a = "/usr/home/ken";
var a = "C:\\Users\\Bill";
var interests = (
    "plan9"
    "go"
    "c"
    "asm"
    "scheme"
);

# MultipleAssign
var a, b = "1", "2";
var a, b, c, d, e, ff, ggg, hhhh = "1", "2", "3", "4", "5", "6", "7", "8";
var A, B = (), ();

var aa, bb = (), ("a" "b");

# MultipleAssign2
var (a="1");

var (
    a = ()
);

var (
    this = "",
    is = "",
    boring = "",
);

var (
    localHost = "localhost",
    targetHost = "victim.tld",
);

# ExecAssign
var out <= boom;
var _, _ <= nuke deploy --location brasilia;

# set assignments
a = "1";
a = ();
a = ("a" "b");

out, sts <= boom --again;
