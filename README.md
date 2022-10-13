# Work In Progress

## Build
 - use go 1.19 (1.18 must work to)
 - go build -o vsd && ./vsd serve --debug
 - admin ui http://127.0.0.1:8090/_/

 ## Redirect from 80 to 8090 port (for telegram auth)
`socat TCP-LISTEN:80,fork TCP:127.0.0.1:8090`
http://127.0.0.1/auth?id=1263310&first_name=recoilme&username=recoilme&photo_url=https%3A%2F%2Ft.me%2Fi%2Fuserpic%2F320%2FjKp4n4Lk3i9yDV1dBo3WQrL3mFaQl7bgLgd0Ip_UWZM.jpg&auth_date=1664374146&hash=2ca16295dc2ff2caa029633a0a6e4c3f3eef42daed739ddb69085308041fb766

<img width="994" alt="Снимок экрана 2022-09-20 в 17 32 56" src="https://user-images.githubusercontent.com/417177/191286332-3be6531d-e39f-4fb4-a4b5-7ae3a8b3b48d.png">
