# Chirpy


This is my eighth Boot.Dev project.


I will be making a HTTP server in Go
that uses a RESTful API,
stores data in an SQL database,
implements authentication/authorization,
and uses a webhook.


# Requirements


1. Go


Install by running `curl -sS https://webi.sh/golang | sh` in your terminal on Linux/WSL/macOS.


Or follow the [official Golang installation instructions](https://go.dev/doc/install)


2. Postgres


Install on macOS with


`brew install postgresql@15`


Install on Linux/WSL with


`sudo apt update`


`sudo apt install postgresql postgresql-contrib`


Or follow the [official Microsoft installation instructions](https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql)


_On Linux/WSL_ Set Password


`sudo passwd postgres`


# How To Use


1. Clone Repository


With GitHub CLI


`gh repo clone smail1111/chirpy`


2. Set Config


In the same directory, create a file named `.env` with the following contents.




```
DB_URL= { YOUR OWN DATABASE CONNECTION STRING }
PLATFORM="dev"
SECRET= { YOUR OWN SECRET STRING }
POLKA_KEY= { YOUR OWN KEY STRING }
```


3. Install Chirpy


While in the repository, run


`go build`


`go install`


This will install chirpy and allow you to activate the chirpy server from any location.


4. Run Chirpy


Run


`chirpy`


This will host chirpy at `http://localhost:8080`


# Resources


1. `/app`


`GET` -> Returns the fileserver hosted from the base directory


2. `admin/metrics`


`GET` -> Returns the number of times the `/app` resource has been visited.


3. `admin/reset`


`POST` -> Resets the entire database. Your platform must be `dev` to use this command.


4. `/api/users`


`POST` -> Registers a new user in the database with the provided email and password.
Returns a JSON response containing the created user's information,
excluding the hashed password.


`PUT` -> Updates a user in the database's information based on the request's access token with the provided email and password.
A valid access token must be in the request's Authorization header to use,
and the user to update will be inferred from it.
Returns a JSON response containing the updated user's information, excluding the hashed password.


Input JSON ->


```
{
    "email": {string},
    "password": {string}
}
```


5. `/api/chirps`


`GET` -> Returns a JSON response containing the information of every chirp registered in the database,
ordered by when they were created.
The `?sort=desc` query parameter will return the chirps in reversed order.


`POST` -> Creates a new chirp in the database with the given body based on the request's access token.
A valid access token must be in the request's Authorization header to use,
and the user the chirp will be assigned to will be inferred from it.
Returns a JSON response containing the created chirp's information.


Input JSON ->


```
{
    "body": {string}
}
```


6. `api/chirps/{id}`


`GET` -> Returns a JSON response containing a specific chirp's information based on the provided id in the url.


`DELETE` -> Deletes a specific chirp in the database based on the provided id in the url.
A valid access token must in request's Authorization header,
and the access token must be for the user that created the chirp to use.


7. `api/login`


`POST` -> If the password for the user in the database with the given email is correct,
returns a JSON response containing the user's information,
a new access token for the user,
and a new refresh token for the user.


Input JSON ->


```
{
    "email": {string},
    "password": {string}
}
```


8. `/api/refresh`


`POST` -> Returns a JSON response containing a new access token for the user assigned to the refresh token in the request's Authorization header.
A valid refresh token must be in the request's Authorization header to use.


9. `/api/revoke`


`POST` -> Revoke the refresh token in the request's Authorization header.


10. `api/polka/webhooks`


`POST` -> Upgrades a user in the database with the given user_id within the given data to chirpyRed.
An ApiKey must be in the request's Authorization header and the ApiKey must match with the secret POLKA_KEY to use.

Input JSON -> 

```
{
    "event": {string},
    "data": {
        "user_id": string
    }
}
```
