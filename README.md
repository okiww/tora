# Try Out API 
Minimum build API for Try Out System BE Exercise using Go language powered by [Gin web framework](https://github.com/gin-gonic/gin) and [GORM](https://github.com/jinzhu/gorm) as backend DAL layer.

## Run the code

The first clone or download this project

### The way

You will need Go installed in your local machine

* Install dep dependency manager

  `$ go get -u github.com/golang/dep/cmd/dep`

* Run dep ensure to downloads all dependencies

  `$ dep ensure`

* Copy file default.yml into .env.yml and modify the config to suit your environment

  `$ cp default.yml .env.yml`

* Ensure your database server is running and application table of your choice (by default it is ruangguru, you can change it in .env.yml file) is exist

* Run the app. For first run you may want to add `--migrate` and `--seeder` switch to run auto db migration and seeding data.

  `$ go run main.go --migrate --seeder`

## API Documentation

By default the app will listen on all interface at port `8000`. Here is the list of endpoint curently available

* Login `POST /login` login using `admin@admin.com` or `user@user.com` and password `12345678`
* List Test `GET /api/v1/list-test`
* Detail Test `GET /api/v1/test/:id_test/detail`

### API SPECIFIC FOR ADMIN
* Create Test `POST /api/v1/create-test`
* Create Question  `POST /api/v1/create-question`
* Update Test `POST /api/v1/update-test`
* Update Question `POST /api/v1/update-question`
* Update Choice `POST /api/v1/update-choice`
* Delete Test `DELETE /api/v1/delete`
* Delete Question `DELETE /api/v1/delete-question`
* Delete Choice `DELETE /api/v1/delete-choice`

### API SPECIFIC FOR USER

* User Attempt Test `POST /api/v1/user/attempt-test`
* User Answer Test  `POST /api/v1/user/answer`
* Get Results `GET /api/v1/user/test/:id_test/result`