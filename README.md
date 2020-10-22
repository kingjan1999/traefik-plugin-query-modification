# Traefik Plugin for Modifying Query Parameters

This Traefik plugin allows you to modify the query parameters of an incoming request, by either adding new, deleting or modifying existing query parameters.
E.g. you can transform `?a=b&c=d&e=f` to `?a=g&h=i` by using this plugin (multiple times).

## Sample Complete Configuration

The following code snippet is a sample configuration for the dynamic file based provider (TOML), but as usual, this plugin should work with all other configuration providers as well.

```toml
[http]
  [http.routers]
    [http.routers.router0]
      entryPoints = ["http"]
      service = "service-foo"
      rule = "Path(`/foo`)"
      middlewares = ["my-plugin"]

  [http.middlewares]
    [http.middlewares.my-plugin.plugin.dev]
      type = "modify"
      paramName = "password"
      newValue = "censored"
```

## Configuration Overview

This plugin knows three different modifications:

### Adding new parameters (`type = "add"`)

Specify the type (`add`), the name / key of the new query parameter (`paramName`) and the value of the new parameter (`newValue`).

Example: 
```toml
type = "add"
paramName = "authenticated"
newValue = "true"
```

Transforms this querystring: `?some=other&stuff=here` into: `?some=other&stuff=here&authenticated=true`

*Note*: Existing query params with the same name are not replaced, instead a new param with the same name is added. Using the previous example:
`?authenticated=false` becomes `?authenticatd=false&authenticated=true`. The handling of such query strings depends on your upstream server. To replace existing values, use `modify`.

### Modifying existing parameters (`type = "modify"`)

This is the most complex mode, as it supports multiple configuration types. You always need to specify which parameters to modify and how the new value should be computed. Avoid configuration more of one way for each of these (e.g. `paramName` and `paramNameRegex`) as this might result in unexpected behavior.

#### Specifying parameter

You have three choices:

- `paramName` matches the plain name / key of the parameter (e.g. `paramName="test"` matches the `test=1234` param in `?test=1234&othertest=5678`)
- `paramNameRegex` matches the name / key of the parameter with a regex (e.g. `paramNameRegex="^.*test$"` matches `test=1234` and `othertest=5678` in `?test=1234&othertest=5678`)
- `paramValueRegex` matches the value of the parameter with a regex (e.g. `paramValueRegex="^1234$"` matches `test=1234` in `?test=1234&othertest=5678`)

Note: While always all matched parameters are handled, you might want to consider just using this middleware plugin multiple times instead of trying to create complex regexes for your situation.

#### Specifying substitution

There are two nays:

- `newValue` replaces the old value with the specifying value. `$1` is replaced by the old value (note: as of now, this is not escapable) (e.g. `paramName="test",newValue="bar-$1"` transforms `test=foo` into `test=bar-foo`)
- `newValueRegex` allows you to use the capture groups from `paramValueRegex` to create the replacement value (e.g. `paramValueRegex="^(.*)oo$",newValueRegex="$1"` transforms `test=foo&test2=poo` into `test=f&test=p`)


### Deleting existing parameters (`type = "delete"`)

This deletes an existing parameters including all of it's values. Specifying the affected parameters works the same [as above](https://github.com/kingjan1999/traefik-plugin-query-modification#specifying-parameter).
Example: `type="delete",paramValueRegex="password"` transforms `?secret=password&othersecret=other-password&tracker=1234` into `tracker=1234`