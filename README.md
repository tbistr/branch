# Branch

`branch` is a simple demux display tool for the CLI.
It reads from stdin and runs multiple commands in parallel, piping the input to each of them.

## Example

```bash
head -c 1000000 /dev/urandom | pv -q -L 1k | hexdump -C | go run . --grep=A --grep=B
```
