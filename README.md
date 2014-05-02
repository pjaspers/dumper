## What does it do?

I'm working with a varying amount of Rails apps, and switching between them regularly requires me to dump a database to a sql file, or load a sql file into the database.

These commands are not hard but I always had to `C-r` to find the correct command, then look in the yaml for database settings before I could issue the command.

Now I just issue `dumper staging` and it will print the command to dump the `staging` database to a sql file.

At the moment it knows `postgres`, `mysql` and `sqlite`.

Example:

```
❯ dumper staging
Dump:

PGPASSWORD=lepass pg_dump -Fc --no-acl --no-owner --clean -U leuser -h lehost ledatabase > app_name_sta_20140503.dump
```

## What else does it do?

```
❯ dumper -h
usage: dumper <environment>
  -F=false: Show restore operation
  -i=[]: comma-separated list of tables to ignore
  -p="": Path to yaml (otherwise config/database.yml)
  -v=false: Prints current dumper version
```

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Do some awesome computering
4. Commit your changes (`git commit -am 'Add some feature'`)
5. Push to the branch (`git push origin my-new-feature`)
6. Create new Pull Request
