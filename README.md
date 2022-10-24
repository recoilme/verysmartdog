# Work In Progress

## Build

- use go 1.19 (1.18 must work to)
- go build -o vsd && ./vsd serve --debug
- admin ui http://127.0.0.1:8090/_/ (open in private window)
- live reload with air (https://github.com/cosmtrek/air): air

## Redirect from 80 to 8090 port (for telegram auth)

`socat TCP-LISTEN:80,fork TCP:127.0.0.1:8090`

<img width="994" alt="Снимок экрана 2022-09-20 в 17 32 56" src="https://user-images.githubusercontent.com/417177/191286332-3be6531d-e39f-4fb4-a4b5-7ae3a8b3b48d.png">
