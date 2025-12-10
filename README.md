# tls-connector
## Описание
Прокси с tls авторизацией по сертификатам  
**Клиент:**
Подключается к серверу используя сертификаты  
Слушает локальный порт, весь трафик пришедший на этот порт без изменений отправляется на сервер к которому произведено подключение  

**Сервер:**  
Слушает порт, разрешает подключение только от клиентов с правильным сертификатом  
Пренаправляет поступающие запросы в другой порт (порт бекенда)  

**Пример использования**
Использование в качестве защиты от посторонних подключений на сервисах без авторизации.  
Например собственный Mincraft сервер для друзей. 
"Minrcraft client" --localhost--> "tls-connector_client" ===Internet===> "tls-connector_server" --localhost--> "Minrcraft server"

## Сборка
```
# Client
cd tls-connector/client
go build -o tls-connector-client

# Server
cd tls-connector/client
go build -o tls-connector-server
```

## Выпуск сертификатов
```
cd tls-connector/certs
# Create CA.crt CA.key
00_generate_ca.sh

# Create serever crt key, out dir tls-connector/certs/server
01_generate_server.sh

# Create client crt key, out dir tls-connector/certs/client
01_generate_server.sh

```
