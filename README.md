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

<img width="994" alt="Снимок экрана 2022-09-20 в 17 32 56" src="https://user-images.githubusercontent.com/417177/191286332-3be6531d-e39f-4fb4-a4b5-7ae3a8b3b48d.png">
