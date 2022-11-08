# Work In Progress

## Build

- use go 1.18+
- go build -o vsd && ./vsd serve --debug
- open admin ui http://127.0.0.1:8090/_/ create admin
- import db shema: pb_shema.json
- run socat for auth via telegram localy `socat TCP-LISTEN:80,fork TCP:127.0.0.1:8090`
- now you may use as user - open http://127.0.0.1 (clean localstorage and cookie for 127.0.0.1 before 1st run)
- for admin ui - open it in private window (user and admin don't may live together)
- development - use live reload with air (https://github.com/cosmtrek/air): air

## Update feeds and posts

- ./vsd checkfeeds

<img width="1274" alt="Снимок экрана 2022-11-08 в 16 41 16" src="https://user-images.githubusercontent.com/417177/200582500-47ce9af4-9ad5-4eaa-91b6-204158d377d5.png">
