// Linux don't allow multithreaded applications to enter a namespace, then
// it's impossible to use Go for such task. The only way is defining a cgo
// constructor to run before the runtime starts their threads and invoke
// the setns there. But we have the problem of need the command line
// arguments... The trick used was read the /proc/self/cmdline :D
// Thanks to @minux on #go-nuts.
package main

import (
	"os"
	"fmt"
	"os/exec"
)

/*
#define _GNU_SOURCE
#include <fcntl.h>
#include <sched.h>
#include <unistd.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

__attribute__((constructor)) void init() {
    // call setns here, it's guaranteed to be single-threaded at this time.
    int fd, argsfile, i, n;
    char *progname = NULL, *nsfile = NULL, *binary = NULL;
    char cmdline[1024], *buf;

    memset(cmdline, 0, 1024);

    argsfile = open("/proc/self/cmdline", O_RDONLY);
    if (argsfile == -1) {
        printf("Fail: open cmdline\n");
        exit(1);
    }

    n = read(argsfile, &cmdline, 1024);
    if (n == -1) {
        printf("Fail: cmdline\n");
        exit(1);
    }

    buf = cmdline;

    for (i = 0; i < n; i++) {
        if (cmdline[i] == 0) {
            if (progname == NULL)
                progname = buf;
            else if (nsfile == NULL)
                nsfile = buf;
            else if (binary == NULL)
                binary = buf;

            buf += i + 1;
        }
    }

    if (progname == NULL || nsfile == NULL || binary == NULL) {
        printf("Usage: sudo ./nsExec <namespace file> <program>\n");
        exit(1);
    }

    printf("Nsfile: %s\n", nsfile);

    fd = open(nsfile, O_RDONLY);   
    if (fd == -1) {
        printf("Fail: nsfile\n");
        exit(1);
    }
    
    if (setns(fd, 0) == -1) {
        printf("Fail: setns - requires root\n");
        exit(1);
    }
}
*/
import "C" 

func main() {
	c := exec.Command(os.Args[2])

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Printf(err.Error())
	}	
}
