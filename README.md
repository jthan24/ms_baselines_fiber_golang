# MS Baselines Golang

## Start
- task docker:dev
- task docker:start
-  go to localhost:3000

## TODO
- livenes & readines https://github.com/rookie-ninja/rk-entry/blob/master/entry/common_service_entry.go#L226 https://github.com/rookie-ninja/rk-fiber/blob/db18984dcd90f950b9054d1cb26d29efd89abac7/boot/fiber_entry.go#L354
- clean arch/hex arch (More or LEss)
- testing
- timeout https://github.com/rookie-ninja/rk-fiber/tree/master/middleware/timeout
- Document Swagger status codes (500s, 400s, etc)
- Create error messages instead of returing the json representation of the error
- Enable swagger only when passing an specific option
- Paginators for handlers
- Hashids
- Migrate client to the same arch
- Create manage post usecases

## Resources

- [semver otel](https://github.com/open-telemetry/opentelemetry-specification/tree/main/semantic_conventions)
- [semver otel connection attrs](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/trace/semantic_conventions/database.md#connection-level-attributes)
- [uptrace golang instrumentations](https://uptrace.dev/opentelemetry/instrumentations/?lang=go)
