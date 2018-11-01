# Mapper-Server

Microservice which process source and template files to create multiple results. 🎬

## What is this for?

For example you have `docx` template and you want to get 100 `pdf` files. 🌹
So you could take `CSV` file with 100 rows 🚣. Every row has column titled as `DATA_TITLE` and contains data to fill template in.
Then you could add special markers 🔖 like `%DATA_TITLE%` to `docx` file.
So when you send this two files (source and template file) 📜 to microservice and get result of processing.

## Use cases

* Generating similar `pdf` files to send emails 📫
* Printing documents and forms 📑
* Anything else 👾

## Powered by

- 🐨 `Go lang`
- 🌺 `Heroku`
