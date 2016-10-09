package parser

import "testing"

type fmtTestTable struct {
	input, expected string
}

func testFmt(input string, expected string, t *testing.T) {
	p := NewParser("fmt test", input)

	tree, err := p.Parse()

	if err != nil {
		t.Error(err)
		return
	}

	fmtval := tree.String()

	if fmtval != expected {
		t.Errorf("Fmt differ: '%s' != '%s'", fmtval, expected)
		return
	}
}

func testFmtTable(testTable []fmtTestTable, t *testing.T) {
	for _, test := range testTable {
		testFmt(test.input, test.expected, t)
	}
}

func TestFmtVariables(t *testing.T) {
	testTable := []fmtTestTable{

		// correct adjust of spaces
		{`test = "a"`, `test = "a"`},
		{`test="a"`, `test = "a"`},
		{`test= "a"`, `test = "a"`},
		{`test  ="a"`, `test = "a"`},
		{`test =    "a"`, `test = "a"`},
		{`test	="a"`, `test = "a"`},
		{`test		="a"`, `test = "a"`},
		{`test =	"a"`, `test = "a"`},
		{`test =		"a"`, `test = "a"`},
		{`test = ()`, `test = ()`},
		{`test=()`, `test = ()`},
		{`test =()`, `test = ()`},
		{`test	=()`, `test = ()`},
		{`test=	()`, `test = ()`},
		{`test = (plan9)`, `test = (plan9)`},
		{`test=(plan9)`, `test = (plan9)`},
		{`test      = (plan9)`, `test = (plan9)`},
		{`test	= (plan9)`, `test = (plan9)`},
		{`test	=	(plan9)`, `test = (plan9)`},
		{`test = (	plan9)`, `test = (plan9)`},
		{`test = (     plan9)`, `test = (plan9)`},
		{`test = (plan9     )`, `test = (plan9)`},
		{`test = (plan9 from bell labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9         from bell labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9         from         bell         labs)`, `test = (plan9 from bell labs)`},
		{`test = (plan9	from	bell	labs)`, `test = (plan9 from bell labs)`},
		{`test = (
	plan9
	from
	bell
	labs
)`, `test = (plan9 from bell labs)`},
		{`test = (plan9 from bell labs windows linux freebsd netbsd openbsd)`, `test = (
	plan9
	from
	bell
	labs
	windows
	linux
	freebsd
	netbsd
	openbsd
)`},

		{`IFS = ("\n")`, `IFS = ("\n")`},

		// multiple variables
		{`test = "a"
testb = "b"`, `test  = "a"
testb = "b"`},
	}

	testFmtTable(testTable, t)
}

func TestFmtGroupVariables(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`test = "a"

test2 = "b"

fn cd() { echo "hello" }`,
			`test  = "a"
test2 = "b"

fn cd() {
	echo "hello"
}`,
		},
		{
			`#!/usr/bin/env nash
echo "hello"`,
			`#!/usr/bin/env nash

echo "hello"`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtFn(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`fn lala() { echo hello }
fn lele() { echo lele }`,
			`fn lala() {
	echo hello
}

fn lele() {
	echo lele
}`,
		},
		{
			`vv = ""
fn t() {
	echo t
}`,
			`vv = ""

fn t() {
	echo t
}`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtImports(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`import test
import test

import test`,
			`import test
import test
import test`,
		},
		{
			`import nashlib/all
import klb/aws/all

vpcTags = ((Name klb-vpc-example) (Env testing))
`,
			`import nashlib/all
import klb/aws/all

vpcTags = (
	(Name klb-vpc-example)
	(Env testing)
)`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtFnComments(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`PATH = "/bin"

# isolated comment

# Comment for fn
fn test() {
	echo "hello"
}
`,
			`PATH = "/bin"

# isolated comment

# Comment for fn
fn test() {
	echo "hello"
}`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtSamples(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`#!/usr/bin/env nash
import nashlib/all
import klb/aws/all
vpcTags = ((Name klb-vpc-example) (Env testing))
igwTags = ((Name klb-igw-example) (Env testing))
routeTblTags = ((Name klb-rtbl-example) (Env testing))
appSubnetTags = ((Name klb-app-subnet-example) (Env testing))
dbSubnetTags = ((Name klb-db-subnet-example) (Env testing))
sgTags = ((Name klb-sg-example) (Env testing))
fn print_resource(name, id) {
	printf "Created %s: %s%s%s\n" $name $NASH_GREEN $id $NASH_RESET
}
fn create_prod() {
	vpcid <= aws_vpc_create("10.0.0.1/16", $vpcTags)
	appnet <= aws_subnet_create($vpcid, "10.0.1.0/24", $appSubnetTags)
	dbnet <= aws_subnet_create($vpcid, "10.0.2.0/24", $dbSubnetTags)
	igwid <= aws_igw_create($igwTags)
	tblid <= aws_routetbl_create($vpcid, $routeTblTags)
	aws_igw_attach($igwid, $vpcid)
	aws_route2igw($tblid, "0.0.0.0/0", $igwid)
	grpid <= aws_secgroup_create("klb-default-sg", "sg description", $vpcid, $sgTags)
	print_resource("VPC", $vpcid)
	print_resource("app subnet", $appnet)
	print_resource("db subnet", $dbnet)
	print_resource("Internet Gateway", $igwid)
	print_resource("Routing table", $tblid)
	print_resource("Security group", $grpid)
}
create_prod()
`,
			`#!/usr/bin/env nash

import nashlib/all
import klb/aws/all

vpcTags = (
	(Name klb-vpc-example)
	(Env testing)
)

igwTags = (
	(Name klb-igw-example)
	(Env testing)
)

routeTblTags = (
	(Name klb-rtbl-example)
	(Env testing)
)

appSubnetTags = (
	(Name klb-app-subnet-example)
	(Env testing)
)

dbSubnetTags = (
	(Name klb-db-subnet-example)
	(Env testing)
)

sgTags = (
	(Name klb-sg-example)
	(Env testing)
)

fn print_resource(name, id) {
	printf "Created %s: %s%s%s\n" $name $NASH_GREEN $id $NASH_RESET
}

fn create_prod() {
	vpcid  <= aws_vpc_create("10.0.0.1/16", $vpcTags)
	appnet <= aws_subnet_create($vpcid, "10.0.1.0/24", $appSubnetTags)
	dbnet  <= aws_subnet_create($vpcid, "10.0.2.0/24", $dbSubnetTags)
	igwid  <= aws_igw_create($igwTags)
	tblid  <= aws_routetbl_create($vpcid, $routeTblTags)

	aws_igw_attach($igwid, $vpcid)
	aws_route2igw($tblid, "0.0.0.0/0", $igwid)

	grpid <= aws_secgroup_create("klb-default-sg", "sg description", $vpcid, $sgTags)

	print_resource("VPC", $vpcid)
	print_resource("app subnet", $appnet)
	print_resource("db subnet", $dbnet)
	print_resource("Internet Gateway", $igwid)
	print_resource("Routing table", $tblid)
	print_resource("Security group", $grpid)
}

create_prod()`,
		},
	}

	testFmtTable(testTable, t)
}

func TestFmtPipes(t *testing.T) {
	testTable := []fmtTestTable{
		{
			`echo hello | grep "he" > test`,
			`echo hello | grep "he" > test`,
		},
		{
			`(echo hello | sed "s/he/wo/g" >[1] /tmp/test >[2] /dev/null)`,
			`(
	echo hello |
	sed "s/he/wo/g"
		>[1] /tmp/test
		>[2] /dev/null
)`,
		},
		{
			`choice <= (
                -find $dir+"/" -maxdepth 1 |
                sed "s#.*/##" |
                sort |
                uniq |
                -fzf --exact -q "^"+$query -1 -0 --inline-info --header "select file: "
       )`,
			`choice <= (
	-find $dir+"/"
		-maxdepth
		1 |
	sed "s#.*/##" |
	sort |
	uniq |
	-fzf --exact
		-q
		"^"+$query -1
		-0
		--inline-info
		--header "select file: "
)`,
		},
	}

	testFmtTable(testTable, t)
}
