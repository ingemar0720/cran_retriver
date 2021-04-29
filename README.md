## Cran Retriver
This repository will index packages in https://cran.r-project.org and seed package information into postgresql DB. It also provide a query API to query the packages by package name.
- It accept an environment variable `numbefOfPkgs` to fetch the number of packages and feed into DB. It's an one-time job which execute upon service boot up.
- The query URL is `http://localhost:5000/packages`, you could run http get to query DB and get all package information in JSON format

### How to execute it.
- start service: `docker-compose up`
- stop service: `docker-compose down`

