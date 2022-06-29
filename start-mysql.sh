#! /bin/sh

docker run --rm --name some-mysql -e MYSQL_ROOT_PASSWORD=test -e MYSQL_DATABASE=gitpod -e MYSQL_USER=gitpod -e MYSQL_PASSWORD=test -p 3306:3306 -d mysql:5.7

# Now run ` yarn typeorm migration:generate -n New` from the `gitpod-db` directory.
# Chop out everything that doesn't pertain to the new table.
