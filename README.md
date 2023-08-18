# Тестовое задание на должность junior golang разработчика

### Запуск приложения

Для запуска приложения необходимо клонировать репозиторий: `git clone https://github.com/amaretur/auth-service.git`

Создать конфигурационный файл в директории `config/` по примеру конфигурационного файла `config/example/config.toml`. В случае если вы создали конфигурационный файл в другом месте, необходимо указать путь к нему при запуске приложения: `go run cmd/main.go -config <путь к файлу>`.

Запуск приложения выполняется командой `go run cmd/main.go` или `go run cmd/main.go -config <путь к файлу>`.

Для корректной работы приложения, в БД необходимо заранее создать ttl индекс:
```
db.collection.createIndex(
   { "expire_at": 1 },
   { expireAfterSeconds: 0 }
)
```

### Описание API
Приложение реализует две конечные точки: для создания пары авторизационных токенов на основе идентификатора пользователя и для обновления этих токенов.

Конечная точка №1:
Пример запроса: 
```
curl -X POST -i 'http://localhost:8085/api/v1/sign-in?uuid=61f0c404-5cb3-11e7-907b-a6006ad3dba0'
```
Согласно требованиям к тестовому заданию, идентификатор передается через параметры запроса.
Пример ответа:
``` js
{
	"access":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTIzNTc1OTEsInV1aWQiOiI2MWYwYzQwNC01Y2IzLTExZTctOTA3Yi1hNjAwNmFkM2RiYTAiLCJyX2lkIjoiNjRkZjUwNTNhNWYyYWVjMDc4OTcwYjdlIn0.idbArbEE3GXXvI4saeJaDrVvWwhu9rBA_-viOjkX3btGIcuAndPeyJFz_wfKaymdwz5NZYg6ltAmahOFFR3Cbw",
	"refresh":"5co4RgMhwDa2WCjV68JKYhbuXFabCG2B"
}
```

Конечная точка №2:
Пример запроса: 
```
curl -X POST -i http://localhost:8085/api/v1/refresh --data '{"access":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTIzNTc1OTEsInV1aWQiOiI2MWYwYzQwNC01Y2IzLTExZTctOTA3Yi1hNjAwNmFkM2RiYTAiLCJyX2lkIjoiNjRkZjUwNTNhNWYyYWVjMDc4OTcwYjdlIn0.idbArbEE3GXXvI4saeJaDrVvWwhu9rBA_-viOjkX3btGIcuAndPeyJFz_wfKaymdwz5NZYg6ltAmahOFFR3Cbw","refresh":"5co4RgMhwDa2WCjV68JKYhbuXFabCG2B"}'
```
Пример ответа: 
``` js
{
	"access":"eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2OTIzNTc2NzcsInV1aWQiOiI2MWYwYzQwNC01Y2IzLTExZTctOTA3Yi1hNjAwNmFkM2RiYTAiLCJyX2lkIjoiNjRkZjUwYTlhNWYyYWVjMDc4OTcwYjgwIn0.dMb_R-5DfQDSgpIc4Gc1YohtONdZHrfrDfuY73rAyv1I6U8hZj2eMFmT_QHPf-mGBWB3og-1ny1ppvBYgotNZA",
	"refresh":"rbLSw2DPTdE8rw-jbg1BkS1EqkJEH1J1"
}
```

### Особенности реализации
Access токен представляет собой JWT токен для создания которого используется алгоритм создания подписи HMAC512, который использует алгоритм SHA512 для создания хеша (согласно требованиям). Время его жизни опрелеляется настройками приложения.

Refresh токен представляет из себя случайный набор байт, закодированных в base64. Длина токена - 32 символа. Refresh токен хранится в MongoDB и автоматически удаляется по истечении срока его жизни. После обновления токенов, refresh токен удаляется из БД. Таким образом реализуется защита от повторного использования.

Id документа в MongoDB, в котором хранится refresh токен, добавляется в access токен. Таким образом реализуется связывание двух токенов. За счет этого, обновлять пару авторизационных токенов можно только той парой access и refresh токенов, которые были выданы вместе.


### Примечания
Используемые в этом проекте пакеты `pkg/errors`, `pkg/log`, `pkg/reqid` были реализованы до этого проекта.
