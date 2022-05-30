---
title: Migrating to older version of Nebraska.
weight: 11
---
 

When migrating down to an older version of Nebraska one might have to rollback database migrations that were applied as a part of the release. 

First get the migrations that were part of the version that you want to downgrade to using the following command. 

> git show <NEBRASKA-VERSION>:./backend/pkg/api/db/migrations | tail -n1

For example to find the migrations that were part of `2.4.0` release clone the Nebraska repo and run.

> git show 2.4.0:../backend/pkg/api/db/migrations | tail -n1
>
> 0013_add_stats_indexes.sql

To rollback to 0013_add_stats_indexes.sql use the following command

> /nebraska/nebraska --rollback-db-to 0013
