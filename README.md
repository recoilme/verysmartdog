# Work In Progress

## Build

- use go 1.18+
- go build -o vsd && ./vsd serve --debug
- run socat (redirect 8080 -> 80) `socat TCP-LISTEN:80,fork TCP:127.0.0.1:8090` (needed for auth via telegram locally)
- remove cookie for 127.0.0.1 (if not first run)
- open admin ui http://127.0.0.1/_/ or user ui http://127.0.0.1
- development - use live reload with air (https://github.com/cosmtrek/air): just run air
- dump db: `sqlite3 pb_data/data.db .dump > data.sql`

## UI

![ui](https://user-images.githubusercontent.com/417177/200583250-8404bef3-418b-490a-93ba-827fdc662807.jpg)
