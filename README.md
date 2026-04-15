# gotmplfmt

A formatter for Go text/template files.

## Install

```
go install github.com/johejo/gotmplfmt@latest
```

## Spacing

`--spacing=false` (default) or `--spacing=true`

```
{{.foo}}
{{range .foo}}
```

to

```
{{ .foo }}
{{ range .foo }}
```

## Indent Style

`--indent-style=none` (default), `--indent-style=tabs`, or `--indent-style=spaces`

`--indent-style=tabs` uses one tab per indent level.

`--indent-style=spaces` uses `--indent-size=<positive-number>` spaces per indent level. The default is `--indent-size=2`.

For example, with `--indent-style=spaces` and `--indent-size=2`, it will change

```
{{range .foo}}
{{.bar}}
{{end}}
```

to 

```
{{range .foo}}
  {{.bar}}
{{end}}
```

## License

MIT License
