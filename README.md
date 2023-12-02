# subserver
A simple config of clash server


### Config
Exec cmd
```shell
cp config-example.yaml config.yaml
```
Change config.yaml file.

### Run
Docker run
```shell
docker-compose -p subserver up -d
```

Stop & Delete
```shell
docker-compose stop
docker-compose rm -f
```


### Vercel
Env config
```shell
touch .env
```
```shell
TOKEN='xxx'
CLASH_SUB_FMT='http://localhost:25500/sub?target=clash&new_name=true&insert=false&config=https%3A%2F%2Fraw.githubusercontent.com%2FACL4SSR%2FACL4SSR%2Fmaster%2FClash%2Fconfig%2FACL4SSR_Online_Full_MultiMode.ini&append_type=true&emoji=true&list=false&tfo=false&scv=false&fdn=true&sort=false&udp=true'
CLASH_SUB_URLS='{"tag":"https://abc.com/subscribe"}'
```

Develop
```shell
vercel dev
```

Deploy
```shell
vercel --prod
```


### Test

```shell
# subserver
curl http://localhost:8008/convert?token=<request-token>&sub_type=a

# subconverter
curl http://localhost:25500/sub?target=clash&new_name=true&insert=false&config=https%3A%2F%2Fraw.githubusercontent.com%2FACL4SSR%2FACL4SSR%2Fmaster%2FClash%2Fconfig%2FACL4SSR_Online_Full_MultiMode.ini&append_type=true&emoji=true&list=false&tfo=false&scv=false&fdn=true&sort=false&udp=true&filename=Clash_t.yaml&url=https%3A%2F%2Fxxx.a.com%2Fsub_url
```

