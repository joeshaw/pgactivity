`pgactivity` is a simple Go program that prints activity within a
Postgres database.  It queries the `pg_stat_activity` view on an
interval (defaulting to every second).

Usage is simple:

    $ pgactivity

Configure it by setting `libpq` environment variables.  [See the
documentation](http://www.postgresql.org/docs/9.4/static/libpq-envars.html)
for all possible values.  Typically you'll want to set `PGHOST`,
`PGUSER` and `PGPASSWORD`:

    $ PGHOST=mydatabase.example.com PGUSER=dbuser PGPASSWORD=wolf pgactivity

A timestamp is printed before each result set is printed.
