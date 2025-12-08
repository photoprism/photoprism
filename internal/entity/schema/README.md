# Database Schema

*This schema description is for illustrative purposes only, e.g. to generate visual relationship diagrams. It should not be used to update or replace an existing production database.*

## Entity-Relationship Diagram

↪ [docs.photoprism.app/developer-guide/database/schema/](https://docs.photoprism.app/developer-guide/database/schema/)

## Mermaid Markup

With [Mermaid.js](https://mermaid-js.github.io/) you can generate visual diagrams from this markup file:

↪ [mariadb.mmd](mariadb.mmd)

## MariaDB SQL Dump

An SQL schema dump can be created using the command shown below, for example:

↪ [mariadb.sql](mariadb.sql)

To create a database schema dump, run the following command in your [development environment](https://docs.photoprism.app/developer-guide/setup/):

```bash
mariadb-dump --no-data --skip-add-locks --skip-comments \
 --skip-opt --skip-set-charset photoprism > mariadb.sql
```

If needed, you can use `grep` to remove magic comments or other unwanted lines from the `mariadb.sql` file:

```bash
cat mariadb.sql | grep -v '^\/\*![0-9]\{5\}.*\/;$' > photoprism-mariadb-database-schema.sql
```

Please note that the dump we provide is only updated at irregular intervals and should therefore not be used to update or replace an existing production database.

## Schema Migrations

↪ [docs.photoprism.app/developer-guide/database/migrations/](https://docs.photoprism.app/developer-guide/database/migrations/)

↪ [github.com/photoprism/photoprism/tree/develop/internal/migrate](https://github.com/photoprism/photoprism/tree/develop/internal/migrate)