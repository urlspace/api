module github.com/urlspace/api

go 1.26.0

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
	github.com/resend/resend-go/v3 v3.0.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.18.0
	go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/trace v1.44.0
	golang.org/x/crypto v0.53.0
)

require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/dave/dst v0.27.4 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/urfave/cli/v3 v3.10.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/tools v0.47.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/log v0.20.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otelc v1.0.1 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)

tool go.opentelemetry.io/otelc/tool/cmd/otelc

replace go.opentelemetry.io/otelc/pkg/runtime => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/pkg/runtime

replace go.opentelemetry.io/otelc/instrumentation/log/slog => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/log/slog

replace go.opentelemetry.io/otelc/instrumentation/go.opentelemetry.io/otel => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/go.opentelemetry.io/otel

replace go.opentelemetry.io/otelc/instrumentation/net/http/client => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/net/http/client

replace go.opentelemetry.io/otelc/instrumentation/database/sql => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/database/sql

replace go.opentelemetry.io/otelc/instrumentation/log => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/log

replace go.opentelemetry.io/otelc/instrumentation/net/http/server => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/net/http/server

replace go.opentelemetry.io/otelc/pkg => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/pkg

replace go.opentelemetry.io/otelc/instrumentation => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation

replace go.opentelemetry.io/otelc/instrumentation/go.opentelemetry.io/otel/trace => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/go.opentelemetry.io/otel/trace

replace go.opentelemetry.io/otelc/instrumentation/go.opentelemetry.io/otel/init => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/go.opentelemetry.io/otel/init

replace go.opentelemetry.io/otelc/instrumentation/runtime => /Users/pawelgrzybek/Developer/url.space/api/.otelc-build/instrumentation/runtime
