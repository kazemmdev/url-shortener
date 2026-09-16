# URL shortener

## Requirements

1. We don't have custom link
2. Links live forever
3. Analytics on each link

## Estimations

1. How many links created per day? about 1m links which means 12 write per secounds
2. How many clicks does each links get? about 1k clicks mean => each writes get 1000 reads => Read heavy
3. How many links will we ever store? 1m per days => 5.6b for 10 years

## API

1. POST /shorten: send long URL
2. GET /{code}: redirect to long URL

## System Design

- The main db model is `urls_tabel` with two `short_code` and `long_url` columns at least
- for each long url we can asign an autoincrement number and hash the number with base62 to prevent any collision
- 0-9, a-z, A-Z can create 62 possiblities where a 7 chracter code can create about 3.5T

### Some security issue

- If someone decode short code can access index number and then can find the next url never share with him
- If using sharding or has same time request counter might return back same number

*Solution*: Add some fixed rules

- Shuffle number
- Add fixed amount to it
- hash with base62

### Heavy Reads

- Click per link = 1000
- links per day = 1000000
- click per day = 1000000000
- 11000 reads/sec

It get worse even by an hot links -> using cache here is mandatory -> first user use db and rest get data from redis (store once, served forever)

## STALENESS => inconsistency between cach and databases

- Here since we never changed the a record => cache each record aggresively - the data never changes

## Redirect without using browser history

- 301 - moved permanently
- 302 - ask me again next time ( if there's analytics let's use this one )

```text
If the business wants click analytics, or the power to kill a link - go with 302 
```

## Bottlenecks

- Redis: it's for performance not source of truth
- Database must not lose any of data so need Replica and Backup

## Questions

1. Two creates at the same time - same code? it can be locked until a counter resolved a request
2. That counter is a one shared thing What if it goes down? we can have multiple replica and set min and max for each pof server
3. We want link to expire, What will you do? add expire data into url table!