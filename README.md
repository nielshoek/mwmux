# This is in development and the code is a mess!

# go-middleware-lib
Simple and small utility to register middleware in the same way as in Express.js / ASP.NET (and probably several others). It is basically a wrapper around Go's `http.ServeMux`.

## Todo
- [ ] Middleware pipeline (`next func`)
- [ ] Invocation of actual action / handler
- [ ] Tests that check order of function invocation
- [ ] Cleanup / Refactor
