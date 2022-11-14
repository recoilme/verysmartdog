# Work In Progress

## Build

- use go 1.18+
- go build -o vsd && ./vsd serve --debug
- run socat (redirect 8080 -> 80) for auth via telegram locally `socat TCP-LISTEN:80,fork TCP:127.0.0.1:8090`
- admin ui http://127.0.0.1/_/
- user ui http://127.0.0.1 (clean localstorage and cookie for 127.0.0.1 before 1st run)
- development - use live reload with air (https://github.com/cosmtrek/air): air

## UI

![ui](https://user-images.githubusercontent.com/417177/200583250-8404bef3-418b-490a-93ba-827fdc662807.jpg)
