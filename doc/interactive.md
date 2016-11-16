# Interactive mode

When used as an interactive shell, nash supports a few features to
enhance user experience.

## Line mode

Nash supports line editing with `emacs` and `vim` modes. The default
mode is `emacs` but it can be changed by the command `set mode vim`,
or setting the environment variable `LINEMODE` with desired value.

When in emacs mode, the following shortcuts can be used:

| Shortcut           | Comment                           |
| ------------------ | --------------------------------- |
| `Ctrl`+`A`         | Beginning of line                 |
| `Ctrl`+`B` / `←`   | Backward one character            |
| `Meta`+`B`         | Backward one word                 |
| `Ctrl`+`C`         | Send io.EOF                       |
| `Ctrl`+`D`         | Delete one character/Close nash   |
| `Meta`+`D`         | Delete one word                   |
| `Ctrl`+`E`         | End of line                       |
| `Ctrl`+`F` / `→`   | Forward one character             |
| `Meta`+`F`         | Forward one word                  |
| `Ctrl`+`H`         | Delete previous character         |
| `Ctrl`+`I` / `Tab` | Command line completion           |
| `Ctrl`+`J`         | Line feed                         |
| `Ctrl`+`K`         | Cut text to the end of line       |
| `Ctrl`+`L`         | Clear screen                      |
| `Ctrl`+`M`         | Same as Enter key                 |
| `Ctrl`+`T`         | Transpose characters              |
| `Ctrl`+`U`         | Cut text to the beginning of line |
| `Ctrl`+`W`         | Cut previous word                 |
| `Backspace`        | Delete previous character         |
| `Meta`+`Backspace` | Cut previous word                 |
| `Enter`            | Line feed                         |

## Autocomplete

Nash doesn't have autocomplete built in, but it do has triggers to you
implement it yourself.

Every time the `TAB` or `CTRL-I (in emacs mode)` is pressed, nash
looks for a function called `nash_complete` declared in the
environment and calls it passing the line buffer and cursor position.

The function must make the autocomplete using some external software
(like [fzf fuzzy finder](https://github.com/junegunn/fzf)) and then
return the characters to be completed. Below is a simple example to
autocomplete system binaries using `fzf`:

```sh
fn diffword(complete, line) {
    diff <= echo -n $complete | sed "s#^"+$line+"##g" | tr -d "\n"

    return $diff
}

fn nash_complete(line, pos) {
    ret = ()
    parts <= split($line, "\n")

    choice <= (
		find /bin /usr/bin -maxdepth 1 -type f |
		sed "s#/.*/##g" |
		sort -u |
		-fzf -q "^"+$line
				-1
				-0
				--header "Looking for system-wide binaries"
				--prompt "(λ programs)>"
				--reverse

	)

    if $status != "0" {
        return $ret
    }

    choice <= diffword($choice, $line)

	ret = ($choice+" " "0")

	return $ret
}
```

## Hooks

There are two functions that can be used to update the environment
while typing commands. The function `nash_repl_before` is called every
time in the cli main loop *before* the printing of the `PROMPT`
variable (and before user can type any command). And the function
called `nash_repl_after` is called every time in the cli main loop
too, but *after* the command was interpreted and executed.

See the examples below:

```sh
DEFPROMPT = "λ> "
fn nash_repl_before() {
    # do something before prompt is ready
    datetime <= date "+%d/%m/%y %H:%M:%S"
    PROMPT = "("+$datetime+")"+$DEFPROMPT
    setenv PROMPT
}

fn nash_repl_after(line, status) {
    # do something after command was executed
    # line and status are the command issued and their
    # exit status (if applicable)
}
```
