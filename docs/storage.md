# PostgreSQL storage tips

Nxpkg stores most data in a
[PostgreSQL database](http://www.postgresql.org). Git repositories,
uploaded user content (e.g., image attachments in issues) are stored
on the filesystem.

## Version requirements

You must use PostgreSQL 9.x (9.6) for development. If you use PostgreSQL 10,
you can't make changes to the database schema and make proper pull requests for
them; we generate files based on the database schema, and the formatting gets
broken. (The changes are minor and cosmetic, but they're a hassle for us in
tracking changes, so we need to all be using compatible-enough versions.)

For Ubuntu 18.04, you will need to add a repository source. Use the
[PostgreSQL.org official repo and instructions.](https://www.postgresql.org/download/linux/ubuntu/)

# Initializing PostgreSQL

Nxpkg assumes it has a dedicated PostgreSQL server, or at least that you
can make global configuration changes, such as changing the timezone. If you
need to use other settings for other databases, use a separate PostgreSQL
instance.

After installing PostgreSQL, set up up a `nxpkg` user and database:

```
sudo su - postgres # this line only needed for Linux
createdb
createuser --superuser nxpkg
psql -c "ALTER USER nxpkg WITH PASSWORD 'nxpkg';"
createdb --owner=nxpkg --encoding=UTF8 --template=template0 nxpkg
```

Then update your `postgresql.conf` default timezone to UTC. Determine the location
of your `postgresql.conf` by running `psql -c 'show config_file;'`. Update the line beginning
with `timezone =` to the following:

```
timezone = 'UTC'
```

Finally, restart your database server (mac: `brew services restart postgresql@9.6`, recent linux, probably `service postgresql restart`)

# Configuring PostgreSQL

The Nxpkg server reads PostgreSQL connection configuration from
the
[`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html);
for example, in your `~/.bashrc`:

```
export PGPORT=5432
export PGHOST=localhost
export PGUSER=nxpkg
export PGPASSWORD=nxpkg
export PGDATABASE=nxpkg
export PGSSLMODE=disable
```

To test the environment's credentials, run `psql` (the PostgreSQL CLI
client) with the `PG*` environment variables set. If you see a
database prompt, then the environment's credentials are valid.

If you get an error message about "peer authentication", you are
probably connecting over the Unix domain socket, rather than over TCP.
Make sure you've set `PGHOST`. (Postgres can do peer authentication
on local sockets, which provides reliable identification but must
be specially configured to authenticate you as a user with a name
different from your account name.)

# Migrations

Migrations get applied automatically at application startup - you
shouldn't need to run anything by hand. For full documentation see
[../migrations/README.md](../migrations/README.md)

# Style guide

Here is the preferred style going forward. Existing tables may be inconsistent with this style.

## Avoiding nullable columns

Use a `NOT NULL` constraint whenever possible to enforce having a value on every column. `NULL` values can easily introduce errors when not handled correctly, and for many fields it makes sense to always have a value anyways.

For example, a `"revision" text` column can use `NULL` to represent no revision, or instead `""` (empty string). On the other hand, it makes sense to represent a `"deleted_at" timestamp` field as `NULL`, meaning "this row has not been deleted".

When NULL fields are necessary, remember to use [`Null*` types](https://golang.org/pkg/database/sql/#NullString) in Go when querying this data. Otherwise `row.Scan` will error after encountering a `NULL` value.

```Go
var s sql.NullString
// Column name can be NULL
err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
...
if s.Valid {
   // use s.String
} else {
   // NULL value
}
```

## Recommended columns for all tables

- `id` auto increment primary key.
- `created_at` not null default `now()` set when a row is first inserted and never updated after that.
- `updated_at` not null default `now()` set when a row is first inserted and updated on every update.
- `deleted_at` set to a not null timestamp to indicate the row is deleted (called soft deleting). This is preferred over hard deleting data from our db (see discussion section below).
  - When querying the db, rows with a non-null `deleted_at` should be excluded.

The timestamps are useful for forensics if something goes wrong, they do not necessarily need to be used or exposed by our graphql APIs. There is no harm in exposing them though.

Example:

```sql
CREATE TABLE "widgets" (
	"id" bigserial NOT NULL PRIMARY KEY,
	"created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"updated_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
	"deleted_at" TIMESTAMP WITH TIME ZONE,
);
```

## Hard vs soft deletes

Definitions:

- A "hard" delete is when rows are deleted using `DELETE FROM table WHERE ...`
- A "soft" delete is when rows are deleted using `UPDATE table SET deleted_at = now() WHERE ...`

Hard deletes are hard to recover from if something goes wrong (application bug, bad migration, manual query, etc.). This usually involves restoring from a backup and it is hard to target only the data affected by the bad delete.

Soft deletes are easier to recover from once you determine what happened. You can simply find the affected rows and `UPDATE table SET deleted_at = null WHERE ...`.

### Dealing with unique constraints

Soft deleting data has implications for unique constraints.

Consider a hypothetical schema:

```sql
CREATE TABLE "orgs" (
	"id" serial NOT NULL PRIMARY KEY
);

CREATE TABLE "users" (
	"id" serial NOT NULL PRIMARY KEY
);

CREATE TABLE "users_orgs" (
	"id" serial NOT NULL PRIMARY KEY,
	"user_id" integer NOT NULL,
	"org_id" integer NOT NULL,

	CONSTRAINT user_orgs_references_orgs
	FOREIGN KEY (org_id)
	REFERENCES orgs (id) ON DELETE RESTRICT,

	CONSTRAINT users_references_users
	FOREIGN KEY (user_id)
	REFERENCES users (id) ON DELETE RESTRICT,

	UNIQUE (user_id, org_id)
);
```

#### Hard delete case

Removing a user from an org deletes the row from `user_orgs`.

Adding a user inserts a row to `user_orgs`. If the user is already a user of the org, the insert fails.

If we wanted to keep a record of membership, it would need to be in a separate audit log table.

#### Soft delete case

Removing a user from an org sets a non-null timestamp on the `deleted_at` column for the row.

Adding a user to an org sets `deleted_at = null` if there is already an existing record for that combination of `user_id` and `org_id`, else a new record is inserted.

Alternatively, we could remove the unique constraint on `user_id` and `org_id` and always insert in the add user case (after checking to see if the user is in the org). This would then function as an audit log table.

The decision here can be made on a table by table basis.

## Use foreign keys

If you have a column that references another column in the database, add a foreign key constraint.

There are reasons to not use foreign keys at scale, but we are not at scale and we can drop these in the future if they become a problem.

### Don't cascade deletes

Foreign key constraints should not cascade deletes for a few reasons:

1.  We don't want to accidentally delete a lot of data (either from our application, or from a manual query in prod).
2.  If we ever add new tables that depend on other tables via foreign key, it is not necessarily the case that cascading the delete is correct for the new table. Explicit application code is better here.
3.  If we ever get to the point of sharding the db, we will probably need to drop all foreign key constraints so it would be great if we did not make our code depend on cascading delete behavior.

Instead of cascading deletes, applications should explicitly delete the rows that would otherwise get deleted if cascading deletes were enabled.

## Table names

Tables are plural (e.g. repositories, users, comments, etc.).

Join tables should be named based on the two tables being joined (e.g. `foo_bar` joins `foo` and `bar`).

## Validation

To the extent that certain fields require validation (e.g. username) we should perform that validation in client AND EITHER the database when possible, OR the graphql api. This results in the best experience for the client, and protects us from corrupt data.

## Trigger functions

Trigger functions perform some action when data is inserted or updated. We don't use trigger functions.
