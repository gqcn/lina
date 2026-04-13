# Manifest Resources

`plugin-demo-dynamic` keeps install-time resources under this directory.

The current sample does not ship SQL migrations, but the directory remains part
of the embedded resource contract so the standard `go:embed` declaration stays
stable when SQL assets are added later.
