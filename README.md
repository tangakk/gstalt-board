Все бд в init.sql

# Методы для постов

### /create POST
Создаёт пост. Если есть JWT в заголовке или куках, использует его. Приоритет у заголовка. Возвращает ОК или ошибку
```
Content-Type: multipart/form-data

Body:
Author = Автор поста. По умолчанию Аноним. Переписывается, если есть токен. Когда-нибудь я начну резать символов 20, но не сейчас
Text = Собственно текст поста
Data = Данные файлика, пока что максимум 2 МБ
Parentid = Айди поста-родителя
Board = /имя-доски
```
### /get-{id} GET
JSON с постом с соответствующим id. В Data будет base64 строка.
### /{board}/get-{offset}-{n} GET
JSON с {n} последних постов на доске /{board}, начиная с {offset}. Только посты с ParentId=0
### /get-responses-{id}-{offset}-{n} GET
Как предыдущее, но ответы к посту, начиная с самого старого.

# Методы для пользователей
### /create-user POST
Создаёт пользователя. Возвращает ОК или ошибку.
```
Content-Type: application/x-www-form-urlencoded

Body:
Name = Имя пользователя
Pass = Пароль
```
### /login POST
Возвращает JWT токен или ошибку.
```
Content-Type: application/x-www-form-urlencoded

Body:
Name = Имя пользователя
Pass = Пароль
```
Ответ:
```
{ "JWT": "многабукав"}
```
