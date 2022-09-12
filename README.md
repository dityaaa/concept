## TODO

### Create "version" command
self describing

### Create "clean" command
_clean_ is used for removing database objects (table, view, stored procedures, etc).
this command has a flag that specify which object need to be cleaned. by default, clean
will remove all database object

### Create "dump" command
same like database schema dumping (ex: mysqldump)

### Create "baseline" command
_baseline_ command is used to trim overly populated migration files. this command works
by dumping current database schema, then using it as initial migration file. this command
can also be used for initializing migration on migration-less database
