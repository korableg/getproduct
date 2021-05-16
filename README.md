# Get Product
Web-scraping сервис, который позволяет получать товар по штрихкоду с сайтов-провайдеров

## Сервис обладает следующим API:
|METHOD|PATH|DESCRIPTION|
| ------ | ------ | ------ |
|GET|/api/barcode/:barcode|Выдает быстрый результат (от самого быстрого провайдера), у остальных провайдеров запрос отменяется. При использовании этого метода ответ не кешируется|
|GET|/api/thebestproduct/:barcode|Сначала получает всевозможные варианты, затем вычисляет самый заполненный. Результат кешируется в локальном хранилище|
|DELETE|/api/localstorage/:barcode|Удаляет результат из локального хранилища|

## Модель товара
Если запрос выполнен успешно, то сервис возвращает JSON с HTTP кодом 200.
Поля в модели могут быть заполнены не все (в зависимости от предоставленной информации сайтом-провайдером).

|PROPERTY|TYPE|DESCRIPTION|
| ------ | ------ | ------ |
|barcode|string|Штрихкод|
|article|string|Артикул|
|name|string|Наименование|
|description|string|Описание товара|
|manufacturer|string|Производитель товара|
|unit|string|Единица измерения товара|
|weight|number|Вес 1 unit товара|
|picture|base64|Изображение товара|

## Сайты-провайдеры
Сайты-провайдеры информации по товару 
