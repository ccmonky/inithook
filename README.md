# inithook

`inithook` used to register attribute setter, which usually used in library code, for example,
if a `render` lib need to set `app_name` to fill output fields, and also some others libs need to set `app_name`, 
thus in these libs, can call `inithook.RegisterAttrSetter` like this:

```go
// e.g. in apierrors lib
apierrorsLibAppName := ""
err := inithook.RegisterAttrSetter(inithook.AppName, "apierrors", func(ctx context.Context, value string) error {
    apierrorsLibAppName = value
    return nil
})
// e.g. in render lib
renderLibAppName := ""
err = inithook.RegisterAttrSetter(inithook.AppName, "render", func(ctx context.Context, value string) error {
    renderLibAppName = value
    return nil
})
```

then, in your app code, you can call `ExecuteAttrSetters` to set the `app_name` once, then all libs's `app_name` will be set,
also, maybe the attr values comes from some config file, you can call the `GetAttrConstructor` to get the attr value constructor
to assist the unmarshal procedure.

```go
// e.g. in app
data := map[string]json.RawMessage{
    inithook.AppName: []byte(`"xxx"`),
    "your-other-attr": ...,
}
err = inithook.ExecuteMapAttrSetters(context.Background(), data)
```
