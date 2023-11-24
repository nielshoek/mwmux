# In development

# go-middleware-lib
Simple and small utility to register middleware in the same way as in Express.js / ASP.NET (and probably several others). It is basically a wrapper around Go's `http.ServeMux`.

## Todo
- [x] Middleware pipeline (`next func`)
- [x] Invocation of actual action / handler
- [x] Tests that check order of function invocation
- [ ] GitHub Actions
- [ ] Cleanup / Refactor
- [ ] Release package