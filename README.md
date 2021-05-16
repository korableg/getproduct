# Get Product
Web-scraping сервис, который позволяет получать товар по штрихкоду

# Сервис обладает следующим API:
|METHOD|PATH|DESCRIPTION|
| ------ | ------ | ------ |
|GET|/api/barcode/:barcode|Выдает быстрый результат (от самого быстрого провайдера), у остальных провайдеров запрос отменяется. При использовании этого метода ответ не кешируется|
|GET|/api/thebestproduct/:barcode|Сначала получает всевозможные варианты, затем вычисляет самый заполненный. Результат кешируется в локальном хранилище|
|DELETE|/api/localstorage/:barcode|Удаляет результат из локального хранилища|

