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

`--indent-style=none` (default) or `--indent-style=tabs` or `--indent-style=spaces`

`--indent-size=2` (default for tabs annd spaces) or `--indent-size=<number>`

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
