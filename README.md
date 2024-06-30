# Branch

`branch` is a simple demux display tool for the CLI.
It reads from stdin and runs multiple commands in parallel, piping the input to each of them.

## Example

```bash
go run ./sample/rand.go | go run . "grep a" "grep b" --default 
```
